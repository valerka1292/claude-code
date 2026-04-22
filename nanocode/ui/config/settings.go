package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	SpinnerHexagons = "hexagons"
	SpinnerCircles  = "circles"
)

type Settings struct {
	SpinnerStyle string `json:"spinner_style"`
}

func DefaultSettings() Settings {
	return Settings{SpinnerStyle: SpinnerHexagons}
}

func LoadSettings() (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultSettings(), nil
		}
		return DefaultSettings(), err
	}

	cfg := DefaultSettings()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultSettings(), err
	}
	if cfg.SpinnerStyle != SpinnerHexagons && cfg.SpinnerStyle != SpinnerCircles {
		cfg.SpinnerStyle = SpinnerHexagons
	}
	return cfg, nil
}

func SaveSettings(cfg Settings) error {
	path, err := settingsPath()
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
	return os.WriteFile(path, data, 0o644)
}

func settingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".nanocode", "settings.json"), nil
}
