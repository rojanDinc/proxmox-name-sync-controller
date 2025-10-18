package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/luthermonson/go-proxmox"
)

// Config holds the Proxmox connection configuration
type Config struct {
	URL      string
	Username string
	Password string
	TokenID  string
	Secret   string
	Insecure bool
}

// Client wraps the Proxmox API client
type Client struct {
	client *proxmox.Client
}

// VM represents a virtual machine with its ID and name
type VM struct {
	ID   int
	Name string
	Node string
}

// NewClient creates a new Proxmox client
func NewClient(config Config) (*Client, error) {
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid Proxmox URL: %w", err)
	}

	var client *proxmox.Client

	if config.TokenID != "" && config.Secret != "" {
		// Use API token authentication
		client = proxmox.NewClient(parsedURL.String(),
			proxmox.WithAPIToken(config.TokenID, config.Secret),
		)
	} else if config.Username != "" && config.Password != "" {
		// Use username/password authentication
		credentials := &proxmox.Credentials{
			Username: config.Username,
			Password: config.Password,
		}
		client = proxmox.NewClient(parsedURL.String(),
			proxmox.WithCredentials(credentials),
		)
	} else {
		return nil, fmt.Errorf("either API token (TokenID and Secret) or credentials (Username and Password) must be provided")
	}

	return &Client{client: client}, nil
}

// GetVMs retrieves all VMs from all nodes in the Proxmox cluster
func (c *Client) GetVMs(ctx context.Context) ([]VM, error) {
	nodes, err := c.client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var allVMs []VM
	for _, nodeStatus := range nodes {
		node, err := c.client.Node(ctx, nodeStatus.Node)
		if err != nil {
			return nil, fmt.Errorf("failed to get node %s: %w", nodeStatus.Node, err)
		}

		vms, err := node.VirtualMachines(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get VMs from node %s: %w", nodeStatus.Node, err)
		}

		for _, vm := range vms {
			allVMs = append(allVMs, VM{
				ID:   int(vm.VMID),
				Name: vm.Name,
				Node: nodeStatus.Node,
			})
		}
	}

	return allVMs, nil
}

// UpdateVMName updates the name of a VM
func (c *Client) UpdateVMName(ctx context.Context, nodeName string, vmid int, newName string) error {
	// Get the node
	node, err := c.client.Node(ctx, nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	// Get the VM
	vm, err := node.VirtualMachine(ctx, vmid)
	if err != nil {
		return fmt.Errorf("failed to get VM %d on node %s: %w", vmid, nodeName, err)
	}

	// Update the VM name using Config method with VirtualMachineOption
	task, err := vm.Config(ctx, proxmox.VirtualMachineOption{
		Name:  "name",
		Value: newName,
	})
	if err != nil {
		return fmt.Errorf("failed to update VM %d name: %w", vmid, err)
	}

	// Wait for the task to complete
	if task != nil {
		// The task is returned, but for name changes it's usually quick
		return nil
	}

	return nil
}

// FindVMByName finds a VM by its current name
func (c *Client) FindVMByName(ctx context.Context, name string) (*VM, error) {
	vms, err := c.GetVMs(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Name == name {
			return &vm, nil
		}
	}

	return nil, nil // VM not found
}

// GetVMIDByName finds VM ID by searching for VMs that might correspond to a node
func (c *Client) GetVMIDByName(ctx context.Context, nodeName string) (*VM, error) {
	vms, err := c.GetVMs(ctx)
	if err != nil {
		return nil, err
	}

	// Try different matching strategies
	for _, vm := range vms {
		// Exact match
		if vm.Name == nodeName {
			return &vm, nil
		}

		// Case-insensitive match
		if strings.EqualFold(vm.Name, nodeName) {
			return &vm, nil
		}

		// Check if VM name contains the node name or vice versa
		if strings.Contains(strings.ToLower(vm.Name), strings.ToLower(nodeName)) ||
			strings.Contains(strings.ToLower(nodeName), strings.ToLower(vm.Name)) {
			return &vm, nil
		}
	}

	return nil, nil
}
