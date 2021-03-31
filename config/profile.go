package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type Profile struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	ApiKey  string `json:"-"`
}

func (p Profile) Save() error {
	if err := p.ensureDir(); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("error generating profile file: %w", err)
	}
	if err := ioutil.WriteFile(p.ConfigFilename(), raw, 0644); err != nil {
		return fmt.Errorf("error saving profile configuration: %w", err)
	}
	if err := ioutil.WriteFile(p.ApiKeyFilename(), []byte(p.ApiKey), 0644); err != nil {
		return fmt.Errorf("error saving api key: %w", err)
	}
	return nil
}

func LoadProfile(name string) (*Profile, error) {
	p := &Profile{
		Name: name,
	}

	if err := p.ensureDir(); err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadFile(p.ConfigFilename())
	if err != nil {
		return nil, fmt.Errorf("error reading profile configuration: %w", err)
	}
	if err := json.Unmarshal(raw, p); err != nil {
		return nil, fmt.Errorf("invalid profile configuration: %w", err)
	}
	// The name in the configuration file should not override the requested profile
	p.Name = name

	if raw, err := ioutil.ReadFile(p.ApiKeyFilename()); err != nil {
		return nil, fmt.Errorf("error reading api key: %w", err)
	} else {
		p.ApiKey = string(raw)
	}
	return p, nil
}

func (p Profile) Directory() string {
	return path.Join(NullstoneDir, p.Name)
}

func (p Profile) ConfigFilename() string {
	return path.Join(p.Directory(), "config")
}

func (p Profile) ApiKeyFilename() string {
	return path.Join(p.Directory(), "key")
}

func (p Profile) ensureDir() error {
	if err := os.MkdirAll(p.Directory(), 0755); !os.IsExist(err) {
		return err
	}
	return nil
}
