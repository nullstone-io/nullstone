package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"golang.org/x/oauth2"
	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type TokenSourcer interface {
	TokenSource(ctx context.Context, scopes ...string) (oauth2.TokenSource, error)
}

type ClusterInfoer interface {
	ClusterInfo() ClusterInfo
}

type ClusterInfo struct {
	ID            string
	Endpoint      string
	CACertificate string
}

// ConfigCreator constructs a kubernetes configuration from a token source and cluster information
type ConfigCreator struct {
	TokenSourcer  TokenSourcer
	ClusterInfoer ClusterInfoer
}

func (f *ConfigCreator) Create(ctx context.Context, scopes ...string) (*restclient.Config, error) {
	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	clusterInfo := f.ClusterInfoer.ClusterInfo()

	decodedCACert, err := base64.StdEncoding.DecodeString(clusterInfo.CACertificate)
	if err != nil {
		return nil, fmt.Errorf("invalid cluster CA certificate: %w", err)
	}

	overrides.ClusterInfo.CertificateAuthorityData = decodedCACert
	host, _, err := restclient.DefaultServerURL(clusterInfo.Endpoint, "", apimachineryschema.GroupVersion{}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GKE cluster host %q: %w", clusterInfo.Endpoint, err)
	}
	overrides.ClusterInfo.Server = host.String()

	kubeTokenSource, err := f.TokenSourcer.TokenSource(ctx, scopes...)
	if err != nil {
		return nil, err
	}
	token, err := kubeTokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("error retrieving kubernetes access token from google cloud: %w", err)
	}
	overrides.AuthInfo.Token = token.AccessToken

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	return cc.ClientConfig()
}
