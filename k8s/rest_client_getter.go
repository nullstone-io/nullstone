package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type RestClientGetter struct {
	Config *rest.Config
}

// ToRESTConfig returns restconfig
func (g RestClientGetter) ToRESTConfig() (*rest.Config, error) {
	return g.Config, nil
}

// ToDiscoveryClient returns discovery client
func (g RestClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return nil, nil
}

// ToRESTMapper returns a restmapper
func (g RestClientGetter) ToRESTMapper() (meta.RESTMapper, error) { return nil, nil }

// ToRawKubeConfigLoader return kubeconfig loader as-is
func (g RestClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig { return nil }
