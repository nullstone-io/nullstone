package config

import (
	"encoding/json"
	"fmt"
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
	if err := os.WriteFile(p.ConfigFilename(), raw, 0600); err != nil {
		return fmt.Errorf("error saving profile configuration: %w", err)
	}
	if err := os.WriteFile(p.ApiKeyFilename(), []byte(p.ApiKey), 0600); err != nil {
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
	// Profiles written by older CLI versions were created world-readable;
	// tighten them in place so the API key is only readable by the owner.
	TightenPerms(p.Directory(), 0700)
	TightenPerms(p.ConfigFilename(), 0600)
	TightenPerms(p.ApiKeyFilename(), 0600)

	raw, err := os.ReadFile(p.ConfigFilename())
	if err != nil {
		// If profile configuration file does not exist, just return our defaults
		if os.IsNotExist(err) {
			return p, nil
		}
		return nil, fmt.Errorf("error reading profile configuration: %w", err)
	}
	if err := json.Unmarshal(raw, p); err != nil {
		return nil, fmt.Errorf("invalid profile configuration: %w", err)
	}
	// The name in the configuration file should not override the requested profile
	p.Name = name

	if raw, err := os.ReadFile(p.ApiKeyFilename()); err != nil {
		return nil, fmt.Errorf("error reading api key: %w", err)
	} else {
		p.ApiKey = CleanseApiKey(string(raw))
	}
	return p, nil
}

func (p Profile) LoadOrg() (string, error) {
	raw, err := os.ReadFile(path.Join(p.Directory(), "org"))
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (p Profile) SaveOrg(org string) error {
	if err := p.ensureDir(); err != nil {
		return err
	}
	return os.WriteFile(path.Join(p.Directory(), "org"), []byte(org), 0600)
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
	if err := os.MkdirAll(p.Directory(), 0700); !os.IsExist(err) {
		return err
	}
	return nil
}

// TightenPerms narrows filename to mode if it exists and is readable by group or
// other. It is a best-effort migration for files created by older CLI versions:
// a missing file or a failed chmod is silently ignored.
func TightenPerms(filename string, mode os.FileMode) {
	fi, err := os.Stat(filename)
	if err != nil {
		return
	}
	if fi.Mode().Perm()&0077 == 0 {
		return
	}
	_ = os.Chmod(filename, mode)
}
