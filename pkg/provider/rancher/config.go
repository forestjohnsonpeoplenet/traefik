package rancher

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/label"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider"
)

func (p *Provider) buildConfiguration(ctx context.Context, services []rancherData) *config.Configuration {
	configurations := make(map[string]*config.Configuration)

	for _, service := range services {
		ctxService := log.With(ctx, log.Str("service", service.Name))

		if !p.keepService(ctx, service) {
			continue
		}

		logger := log.FromContext(ctxService)

		confFromLabel, err := label.DecodeConfiguration(service.Labels)
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			err := p.buildTCPServiceConfiguration(ctxService, service, confFromLabel.TCP)
			if err != nil {
				logger.Error(err)
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxService, confFromLabel.TCP)
			if len(confFromLabel.HTTP.Routers) == 0 &&
				len(confFromLabel.HTTP.Middlewares) == 0 &&
				len(confFromLabel.HTTP.Services) == 0 {
				configurations[service.Name] = confFromLabel
				continue
			}
		}

		err = p.buildServiceConfiguration(ctx, service, confFromLabel.HTTP)
		if err != nil {
			logger.Error(err)
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   service.Name,
			Labels: service.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, service.Name, p.defaultRuleTpl, model)

		configurations[service.Name] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) buildTCPServiceConfiguration(ctx context.Context, service rancherData, configuration *config.TCPConfiguration) error {
	serviceName := service.Name

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*config.TCPService)
		lb := &config.TCPLoadBalancerService{}
		configuration.Services[serviceName] = &config.TCPService{
			LoadBalancer: lb,
		}
	}

	for _, confService := range configuration.Services {
		err := p.addServerTCP(ctx, service, confService.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildServiceConfiguration(ctx context.Context, service rancherData, configuration *config.HTTPConfiguration) error {

	serviceName := service.Name

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*config.Service)
		lb := &config.LoadBalancerService{}
		lb.SetDefaults()
		configuration.Services[serviceName] = &config.Service{
			LoadBalancer: lb,
		}
	}

	for _, confService := range configuration.Services {
		err := p.addServers(ctx, service, confService.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) keepService(ctx context.Context, service rancherData) bool {
	logger := log.FromContext(ctx)

	if !service.ExtraConf.Enable {
		logger.Debug("Filtering disabled service.")
		return false
	}

	if ok, failingConstraint := p.MatchConstraints(service.ExtraConf.Tags); !ok {
		if failingConstraint != nil {
			logger.Debugf("service pruned by %q constraint", failingConstraint.String())
		}
		return false
	}

	if p.EnableServiceHealthFilter {
		if service.Health != "" && service.Health != healthy && service.Health != updatingHealthy {
			logger.Debugf("Filtering service %s with healthState of %s \n", service.Name, service.Health)
			return false
		}
		if service.State != "" && service.State != active && service.State != updatingActive && service.State != upgraded && service.State != upgrading {
			logger.Debugf("Filtering service %s with state of %s \n", service.Name, service.State)
			return false
		}
	}

	return true
}

func (p *Provider) addServerTCP(ctx context.Context, service rancherData, loadBalancer *config.TCPLoadBalancerService) error {
	log.FromContext(ctx).Debugf("Trying to add servers for service  %s \n", service.Name)

	serverPort := ""

	if loadBalancer != nil && len(loadBalancer.Servers) > 0 {
		serverPort = loadBalancer.Servers[0].Port
	}

	port := getServicePort(service)

	if len(loadBalancer.Servers) == 0 {
		server := config.TCPServer{}

		loadBalancer.Servers = []config.TCPServer{server}
	}

	if serverPort != "" {
		port = serverPort
		loadBalancer.Servers[0].Port = ""
	}

	if port == "" {
		return errors.New("port is missing")
	}

	var servers []config.TCPServer
	for _, containerIP := range service.Containers {
		servers = append(servers, config.TCPServer{
			Address: net.JoinHostPort(containerIP, port),
		})
	}

	loadBalancer.Servers = servers
	return nil

}

func (p *Provider) addServers(ctx context.Context, service rancherData, loadBalancer *config.LoadBalancerService) error {
	log.FromContext(ctx).Debugf("Trying to add servers for service  %s \n", service.Name)

	serverPort := getLBServerPort(loadBalancer)
	port := getServicePort(service)

	if len(loadBalancer.Servers) == 0 {
		server := config.Server{}
		server.SetDefaults()

		loadBalancer.Servers = []config.Server{server}
	}

	if serverPort != "" {
		port = serverPort
		loadBalancer.Servers[0].Port = ""
	}

	if port == "" {
		return errors.New("port is missing")
	}

	var servers []config.Server
	for _, containerIP := range service.Containers {
		servers = append(servers, config.Server{
			URL: fmt.Sprintf("%s://%s", loadBalancer.Servers[0].Scheme, net.JoinHostPort(containerIP, port)),
		})
	}

	loadBalancer.Servers = servers
	return nil
}

func getLBServerPort(loadBalancer *config.LoadBalancerService) string {
	if loadBalancer != nil && len(loadBalancer.Servers) > 0 {
		return loadBalancer.Servers[0].Port
	}
	return ""
}

func getServicePort(data rancherData) string {
	rawPort := strings.Split(data.Port, "/")[0]
	hostPort := strings.Split(rawPort, ":")

	if len(hostPort) >= 2 {
		return hostPort[1]
	}
	if len(hostPort) > 0 && hostPort[0] != "" {
		return hostPort[0]
	}
	return rawPort
}
