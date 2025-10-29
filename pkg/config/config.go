package config

import (
	"fmt"
	"io"
	"os"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
	"sigs.k8s.io/yaml"
)

func LoadProxmoxConfig(configPath string) (*proxmox.ClusterConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path empty")
	}

	cfg, err := loadFromYAML(configPath)
	if err != nil {
		return nil, err
	}

	if err := validateConfig(*cfg); err != nil {
		return nil, err
	}

	return cfg, err
}

func validateConfig(config proxmox.ClusterConfig) error {
	if len(config.HostURLs) == 0 {
		return fmt.Errorf("at least one Proxmox URL must be provided")
	}

	hasTokenAuth := config.TokenID != "" && config.Secret != ""
	hasPasswordAuth := config.Username != "" && config.Password != ""

	if !hasTokenAuth && !hasPasswordAuth {
		return fmt.Errorf("authentication credentials are required (token or username/password)")
	}

	return nil
}

func loadFromYAML(filePath string) (*proxmox.ClusterConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open YAML config at %s: %w", filePath, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML config at %s: %w", filePath, err)
	}

	var cfg proxmox.ClusterConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &cfg, nil
}
