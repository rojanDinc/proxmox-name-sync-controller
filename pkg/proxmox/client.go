package proxmox

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
)

const (
	taskInterval = 5 * time.Second
	taskTimeout  = 30 * time.Second
)

type Config ClusterConfig

type ClusterConfig struct {
	HostURLs []string
	Username string
	Password string
	TokenID  string
	Secret   string
	Insecure bool
}

type ClientPool struct {
	clients []*proxmox.Client
}

type VM struct {
	ID   int
	Name string
	Node string
	UUID string
}

func NewClient(clusterConfig *ClusterConfig) (*ClientPool, error) {
	clientPool := &ClientPool{clients: make([]*proxmox.Client, 0)}
	for _, hostURL := range clusterConfig.HostURLs {
		parsedURL, err := url.Parse(hostURL)
		if err != nil {
			return nil, fmt.Errorf("invalid Proxmox URL: %w", err)
		}

		httpClient := &http.Client{}
		if clusterConfig.Insecure {
			httpClient.Transport = &http.Transport{
				// #nosec G402
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}

		var client *proxmox.Client

		if clusterConfig.TokenID != "" && clusterConfig.Secret != "" {
			client = proxmox.NewClient(parsedURL.String(), proxmox.WithAPIToken(clusterConfig.TokenID, clusterConfig.Secret))
		} else if clusterConfig.Username != "" && clusterConfig.Password != "" {
			credentials := &proxmox.Credentials{
				Username: clusterConfig.Username,
				Password: clusterConfig.Password,
			}
			client = proxmox.NewClient(parsedURL.String(),
				proxmox.WithCredentials(credentials),
				proxmox.WithHTTPClient(httpClient),
			)
		} else {
			return nil, fmt.Errorf("either API token (TokenID and Secret) or credentials (Username and Password) must be provided")
		}

		if client != nil {
			clientPool.clients = append(clientPool.clients, client)
		}
	}

	return clientPool, nil
}

func (c *ClientPool) GetVMs(ctx context.Context) ([]VM, error) {
	client, err := c.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	nodes, err := client.Nodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var allVMs []VM
	for _, nodeStatus := range nodes {
		node, err := client.Node(ctx, nodeStatus.Node)
		if err != nil {
			return nil, fmt.Errorf("failed to get node %s: %w", nodeStatus.Node, err)
		}

		vms, err := node.VirtualMachines(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get VMs from node %s: %w", nodeStatus.Node, err)
		}

		for _, partialVM := range vms {
			vm, err := node.VirtualMachine(ctx, int(partialVM.VMID))
			if err != nil {
				return nil, err
			}
			if vm.VirtualMachineConfig == nil {
				slog.Info("Skipping VM with nil configuration", "vmid", vm.VMID, "node", nodeStatus.Node)
				continue
			}
			ok, uuid := extractUUIDFrom(vm.VirtualMachineConfig.SMBios1)
			if !ok {
				slog.Info("Skipping VM with no uuid", "vmid", vm.VMID, "node", nodeStatus.Node)
				continue
			}

			allVMs = append(allVMs, VM{
				ID:   int(vm.VMID),
				Name: vm.Name,
				Node: nodeStatus.Node,
				UUID: uuid,
			})
		}
	}

	return allVMs, nil
}

func (c *ClientPool) UpdateVMName(ctx context.Context, nodeName string, vmid int, newName string) error {
	client, err := c.getClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	node, err := client.Node(ctx, nodeName)
	if err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	vm, err := node.VirtualMachine(ctx, vmid)
	if err != nil {
		return fmt.Errorf("failed to get VM %d on node %s: %w", vmid, nodeName, err)
	}

	task, err := vm.Config(ctx, proxmox.VirtualMachineOption{
		Name:  "name",
		Value: newName,
	})
	if err != nil {
		return fmt.Errorf("failed to update VM %d name: %w", vmid, err)
	}

	if err := task.Wait(ctx, taskInterval, taskTimeout); err != nil {
		return fmt.Errorf("failed to wait for VM %d name update task: %w", vmid, err)
	}

	return nil
}

func (c *ClientPool) FindVMByName(ctx context.Context, name string) (*VM, error) {
	vms, err := c.GetVMs(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Name == name {
			return &vm, nil
		}
	}

	return nil, nil
}

func (c *ClientPool) GetVMIDByName(ctx context.Context, nodeName string) (*VM, error) {
	vms, err := c.GetVMs(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Name == nodeName {
			return &vm, nil
		}

		if strings.EqualFold(vm.Name, nodeName) {
			return &vm, nil
		}

		if strings.Contains(strings.ToLower(vm.Name), strings.ToLower(nodeName)) ||
			strings.Contains(strings.ToLower(nodeName), strings.ToLower(vm.Name)) {
			return &vm, nil
		}
	}

	return nil, nil
}

func (c *ClientPool) GetVMByUUID(ctx context.Context, uuid string) (*VM, error) {
	vms, err := c.GetVMs(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		slog.Debug("found vm", "id", vm.UUID)
		if vm.UUID == uuid {
			return &vm, nil
		}
	}

	return nil, nil
}

func (c *ClientPool) getClient(ctx context.Context) (*proxmox.Client, error) {
	for _, client := range c.clients {
		if _, err := client.Version(ctx); err == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("no client found")
}

func extractUUIDFrom(smbios string) (bool, string) {
	splits := strings.Split(smbios, ",")
	for _, split := range splits {
		if strings.Contains(split, "uuid=") {
			return true, strings.Split(split, "uuid=")[1]
		}
	}

	return false, ""
}
