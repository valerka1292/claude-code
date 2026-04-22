package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Provider struct {
	Name        string `json:"name"`
	BaseURL     string `json:"base_url"`
	Model       string `json:"model"`
	APIKey      string `json:"api_key"`
	ContextSize int    `json:"context_size"`
	Active      bool   `json:"active"`
}

type ProvidersFile struct {
	Providers map[string]Provider `json:"providers"`
}

func LoadProviders() (ProvidersFile, error) {
	path, err := providersPath()
	if err != nil {
		return ProvidersFile{Providers: map[string]Provider{}}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ProvidersFile{Providers: map[string]Provider{}}, nil
		}
		return ProvidersFile{Providers: map[string]Provider{}}, err
	}

	cfg := ProvidersFile{Providers: map[string]Provider{}}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ProvidersFile{Providers: map[string]Provider{}}, err
	}
	if cfg.Providers == nil {
		cfg.Providers = map[string]Provider{}
	}
	return cfg, nil
}

func SaveProviders(cfg ProvidersFile) error {
	path, err := providersPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func ActiveProvider(cfg ProvidersFile) (Provider, bool) {
	for _, provider := range cfg.Providers {
		if provider.Active {
			return provider, true
		}
	}
	return Provider{}, false
}

func ProviderNames(cfg ProvidersFile) []string {
	names := make([]string, 0, len(cfg.Providers))
	for name := range cfg.Providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func NormalizeBaseURL(raw string) string {
	value := strings.TrimSpace(raw)
	return strings.TrimRight(value, "/")
}

func providersPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "nanocode", "providers.json"), nil
}
