package provider

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

var _ Provider = (*File)(nil)

// File holds configurations of the File provider.
type File struct {
	BaseProvider `mapstructure:",squash"`
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *File) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {

	file, err := os.Open(provider.Filename)
	if err != nil {
		log.Error("Error opening file", err)
		return err
	}
	file.Close()

	fileHash, err := hashFile(file.Name())

	if err != nil {
		log.Error("Error hashing file", err)
		return err
	}

	if provider.Watch {
		// Process events
		pool.Go(func(stop chan bool) {
			for {
				select {
				case <-stop:
					return
				default:
					newFileHash, err := hashFile(file.Name())
					if err == nil && newFileHash != fileHash {

						configuration := provider.loadFileConfig(file.Name())
						configurationChan <- types.ConfigMessage{
							ProviderName:  "file",
							Configuration: configuration,
						}

						fileHash = newFileHash
					}
					time.Sleep(time.Second * 5)
				}

			}
		})
	}

	configuration := provider.loadFileConfig(file.Name())
	configurationChan <- types.ConfigMessage{
		ProviderName:  "file",
		Configuration: configuration,
	}
	return nil
}

func (provider *File) loadFileConfig(filename string) *types.Configuration {
	configuration := new(types.Configuration)
	if _, err := toml.DecodeFile(filename, configuration); err != nil {
		log.Error("Error reading file:", err)
		return nil
	}
	return configuration
}

func hashFile(string filename) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("hashFile: Error reading file", err)
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(bytes)), nil
}
