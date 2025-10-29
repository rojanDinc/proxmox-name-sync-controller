package config

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
	"sigs.k8s.io/yaml"
)

func LoadProxmoxConfig() (*proxmox.ClusterConfig, error) {
	// 1) Prefer YAML-based configuration if provided via file or inline env
	if cfg, err := loadFromYAML(); err != nil {
		return nil, err
	} else if cfg != nil {
		return cfg, nil
	}

	// 2) Fallback to environment variables for backward compatibility
	var cfg proxmox.ClusterConfig

	// Support both PROXMOX_URLS (comma-separated) and PROXMOX_URL (single), the latter from existing chart
	urls := os.Getenv("PROXMOX_URLS")
	if urls == "" {
		urls = os.Getenv("PROXMOX_URL")
	}
	if urls != "" {
		// Split on comma and trim spaces
		parts := strings.Split(urls, ",")
		hostURLs := make([]string, 0, len(parts))
		for _, p := range parts {
			s := strings.TrimSpace(p)
			if s != "" {
				hostURLs = append(hostURLs, s)
			}
		}
		cfg.HostURLs = hostURLs
	}

	if insecure := os.Getenv("PROXMOX_INSECURE"); insecure != "" {
		v, err := strconv.ParseBool(insecure)
		if err != nil {
			return nil, fmt.Errorf("invalid PROXMOX_INSECURE value: %w", err)
		}
		cfg.Insecure = v
	}

	// Credentials can be token or username/password
	cfg.TokenID = os.Getenv("PROXMOX_TOKEN_ID")
	cfg.Secret = os.Getenv("PROXMOX_SECRET")
	if cfg.TokenID == "" || cfg.Secret == "" {
		cfg.Username = os.Getenv("PROXMOX_USERNAME")
		cfg.Password = os.Getenv("PROXMOX_PASSWORD")
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
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

func loadFromYAML() (*proxmox.ClusterConfig, error) {
	filePath := os.Getenv("PROXMOX_CONFIG_PATH")

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

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
