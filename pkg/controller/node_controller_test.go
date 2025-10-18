package controller

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
)

// MockProxmoxClient is a mock implementation of the Proxmox client for testing
type MockProxmoxClient struct {
	VMs             []proxmox.VM
	UpdateCallCount int
	LastUpdatedVM   *proxmox.VM
	LastUpdatedName string
}

func (m *MockProxmoxClient) GetVMs(ctx context.Context) ([]proxmox.VM, error) {
	return m.VMs, nil
}

func (m *MockProxmoxClient) UpdateVMName(ctx context.Context, nodeName string, vmid int, newName string) error {
	m.UpdateCallCount++
	for i := range m.VMs {
		if m.VMs[i].ID == vmid {
			m.LastUpdatedVM = &m.VMs[i]
			m.LastUpdatedName = newName
			m.VMs[i].Name = newName // Update the mock VM name
			break
		}
	}
	return nil
}

func (m *MockProxmoxClient) GetVMIDByName(ctx context.Context, nodeName string) (*proxmox.VM, error) {
	for _, vm := range m.VMs {
		// Exact match
		if vm.Name == nodeName {
			return &vm, nil
		}
		// Case-insensitive match
		if vm.Name == "vm-"+nodeName || vm.Name == nodeName+"-vm" {
			return &vm, nil
		}
	}
	return nil, nil
}

func TestNodeReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	err := corev1.AddToScheme(scheme)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                string
		node                *corev1.Node
		existingVMs         []proxmox.VM
		expectedUpdateCount int
		expectedVMName      string
	}{
		{
			name: "VM name matches node name - no update needed",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-1",
				},
			},
			existingVMs: []proxmox.VM{
				{ID: 100, Name: "worker-1", Node: "pve1"},
			},
			expectedUpdateCount: 0,
		},
		{
			name: "VM name doesn't match node name - update needed",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-1",
				},
			},
			existingVMs: []proxmox.VM{
				{ID: 100, Name: "vm-worker-1", Node: "pve1"},
			},
			expectedUpdateCount: 1,
			expectedVMName:      "worker-1",
		},
		{
			name: "Control plane node - should be skipped",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "control-plane-1",
					Labels: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
					},
				},
			},
			existingVMs: []proxmox.VM{
				{ID: 101, Name: "vm-control-plane-1", Node: "pve1"},
			},
			expectedUpdateCount: 0,
		},
		{
			name: "No matching VM found",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-2",
				},
			},
			existingVMs: []proxmox.VM{
				{ID: 100, Name: "different-vm", Node: "pve1"},
			},
			expectedUpdateCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with the node
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.node).
				Build()

			// Create mock Proxmox client
			mockProxmox := &MockProxmoxClient{
				VMs: tt.existingVMs,
			}

			// Create reconciler
			reconciler := &NodeReconciler{
				Client:        fakeClient,
				Scheme:        scheme,
				ProxmoxClient: mockProxmox,
			}

			// Create reconcile request
			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: tt.node.Name,
				},
			}

			// Perform reconciliation
			ctx := context.TODO()
			result, err := reconciler.Reconcile(ctx, req)

			// Verify results
			if err != nil {
				t.Errorf("Reconcile() error = %v", err)
				return
			}

			if mockProxmox.UpdateCallCount != tt.expectedUpdateCount {
				t.Errorf("Expected %d update calls, got %d", tt.expectedUpdateCount, mockProxmox.UpdateCallCount)
			}

			if tt.expectedUpdateCount > 0 && mockProxmox.LastUpdatedName != tt.expectedVMName {
				t.Errorf("Expected VM name to be updated to %s, got %s", tt.expectedVMName, mockProxmox.LastUpdatedName)
			}

			// Verify that the result indicates when to requeue
			_, isControlPlane := tt.node.Labels["node-role.kubernetes.io/control-plane"]
			if result.RequeueAfter == 0 && tt.expectedUpdateCount == 0 && !isControlPlane {
				// Should requeue for regular checks unless it's a control plane node
				if result.RequeueAfter < time.Minute*10 {
					t.Errorf("Expected requeue time for no-op scenarios")
				}
			}
		})
	}
}

func TestNodeReconciler_isControlPlaneNode(t *testing.T) {
	reconciler := &NodeReconciler{}

	tests := []struct {
		name     string
		node     *corev1.Node
		expected bool
	}{
		{
			name: "Node with control-plane label",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"node-role.kubernetes.io/control-plane": "",
					},
				},
			},
			expected: true,
		},
		{
			name: "Node with master label",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"node-role.kubernetes.io/master": "",
					},
				},
			},
			expected: true,
		},
		{
			name: "Node with control-plane taint",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "node-role.kubernetes.io/control-plane",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "Worker node",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"node-role.kubernetes.io/worker": "",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconciler.isControlPlaneNode(tt.node)
			if result != tt.expected {
				t.Errorf("isControlPlaneNode() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
