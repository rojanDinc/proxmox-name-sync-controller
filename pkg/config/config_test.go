package config

import (
	"os"
	"testing"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		rawConfig string
		expected  *proxmox.ClusterConfig
	}{
		{
			name: "read config successfully",
			rawConfig: `---
hostUrls: ["https://192.168.1.111", "https://192.168.1.112"]
username: test
password: test
insecure: true
tokenId: test@pve!test
secret: 1234-23-12323-45235-353`,
			expected: &proxmox.ClusterConfig{
				HostURLs: []string{"https://192.168.1.111", "https://192.168.1.112"},
				Username: "test",
				Password: "test",
				TokenID:  "test@pve!test",
				Secret:   "1234-23-12323-45235-353",
				Insecure: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := os.CreateTemp(os.TempDir(), "*.yaml")
			require.NoError(t, err)
			tempFile.WriteString(tt.rawConfig)
			defer tempFile.Close()
			defer os.Remove(tempFile.Name())

			actual, err := LoadProxmoxConfig(tempFile.Name())
			assert.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}
