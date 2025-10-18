package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
)

// LoadProxmoxConfig loads Proxmox configuration from environment variables
func LoadProxmoxConfig() (proxmox.Config, error) {
	config := proxmox.Config{}

	// Required: Proxmox URL
	config.URL = os.Getenv("PROXMOX_URL")
	if config.URL == "" {
		return config, fmt.Errorf("PROXMOX_URL environment variable is required")
	}

	// Ensure URL has proper scheme
	if !strings.HasPrefix(config.URL, "http://") && !strings.HasPrefix(config.URL, "https://") {
		config.URL = "https://" + config.URL
	}

	// Authentication method 1: API Token (preferred)
	config.TokenID = os.Getenv("PROXMOX_TOKEN_ID")
	config.Secret = os.Getenv("PROXMOX_SECRET")

	// Authentication method 2: Username/Password (fallback)
	if config.TokenID == "" || config.Secret == "" {
		config.Username = os.Getenv("PROXMOX_USERNAME")
		config.Password = os.Getenv("PROXMOX_PASSWORD")

		if config.Username == "" || config.Password == "" {
			return config, fmt.Errorf("either PROXMOX_TOKEN_ID/PROXMOX_SECRET or PROXMOX_USERNAME/PROXMOX_PASSWORD must be provided")
		}
	}

	// Optional: Insecure SSL
	if insecure := os.Getenv("PROXMOX_INSECURE"); insecure != "" {
		var err error
		config.Insecure, err = strconv.ParseBool(insecure)
		if err != nil {
			return config, fmt.Errorf("invalid PROXMOX_INSECURE value: %w", err)
		}
	} else {
		config.Insecure = true // Default to true for self-signed certificates
	}

	return config, nil
}

// ValidateConfig validates the Proxmox configuration
func ValidateConfig(config proxmox.Config) error {
	if config.URL == "" {
		return fmt.Errorf("Proxmox URL cannot be empty")
	}

	// Check if we have valid authentication
	hasTokenAuth := config.TokenID != "" && config.Secret != ""
	hasPasswordAuth := config.Username != "" && config.Password != ""

	if !hasTokenAuth && !hasPasswordAuth {
		return fmt.Errorf("authentication credentials are required")
	}

	return nil
}
