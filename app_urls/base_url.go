package app_urls

import (
	"gopkg.in/nullstone-io/go-api-client.v0"
	"net/url"
	"strings"
)

func GetBaseUrl(cfg api.Config) *url.URL {
	u, err := url.Parse(cfg.BaseAddress)
	if err != nil {
		u = &url.URL{Scheme: "https", Host: "app.nullstone.io"}
	}
	u.Host = strings.Replace(u.Host, "api", "app", 1)
	if u.Host == "localhost:8443" {
		u.Scheme = "http"
		u.Host = "localhost:8090"
	}
	return u
}
