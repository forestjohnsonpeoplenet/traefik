/*
The MIT License (MIT)

Copyright (c) 2016-2019 Containous SAS

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "github.com/containous/traefik/pkg/provider/kubernetes/crd/generated/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// IngressRoutes returns a IngressRouteInformer.
	IngressRoutes() IngressRouteInformer
	// IngressRouteTCPs returns a IngressRouteTCPInformer.
	IngressRouteTCPs() IngressRouteTCPInformer
	// Middlewares returns a MiddlewareInformer.
	Middlewares() MiddlewareInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// IngressRoutes returns a IngressRouteInformer.
func (v *version) IngressRoutes() IngressRouteInformer {
	return &ingressRouteInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// IngressRouteTCPs returns a IngressRouteTCPInformer.
func (v *version) IngressRouteTCPs() IngressRouteTCPInformer {
	return &ingressRouteTCPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Middlewares returns a MiddlewareInformer.
func (v *version) Middlewares() MiddlewareInformer {
	return &middlewareInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
