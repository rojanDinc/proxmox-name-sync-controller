package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type MockProxmoxClient struct {
	GetVMByUUIDFn  func(ctx context.Context, uuid string) (*proxmox.VM, error)
	UpdateVMNameFn func(ctx context.Context, nodeName string, vmid int, newName string) error
}

func (mock *MockProxmoxClient) GetVMByUUID(ctx context.Context, uuid string) (*proxmox.VM, error) {
	return mock.GetVMByUUIDFn(ctx, uuid)
}

func (mock *MockProxmoxClient) UpdateVMName(ctx context.Context, nodeName string, vmid int, newName string) error {
	return mock.UpdateVMNameFn(ctx, nodeName, vmid, newName)
}

func TestNodeReconciler_Reconcile_Scenarios(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		node            corev1.Node
		expectedError   error
		expectedNewName string
		mock            *MockProxmoxClient
	}{
		{
			name: "no update when VM name matches node name",
			node: corev1.Node{
				ObjectMeta: testNodeMeta("worker-01"),
				Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{SystemUUID: "uuid-1"}},
			},
			expectedError: nil,
			mock: &MockProxmoxClient{
				GetVMByUUIDFn: func(ctx context.Context, uuid string) (*proxmox.VM, error) {
					return &proxmox.VM{ID: 100, Name: "worker-01", Node: "pve-1", UUID: "uuid-1"}, nil
				},
				UpdateVMNameFn: func(ctx context.Context, nodeName string, vmid int, newName string) error {
					return nil
				},
			},
		},
		{
			name: "update when VM name differs from node name",
			node: corev1.Node{
				ObjectMeta: testNodeMeta("k8s-node-02"),
				Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{SystemUUID: "uuid-2"}},
			},
			expectedNewName: "k8s-node-02",
			expectedError:   nil,
			mock: &MockProxmoxClient{
				GetVMByUUIDFn: func(ctx context.Context, uuid string) (*proxmox.VM, error) {
					return &proxmox.VM{ID: 200, Name: "old-name", Node: "pve-2", UUID: "uuid-2"}, nil
				},
				UpdateVMNameFn: func(ctx context.Context, nodeName string, vmid int, newName string) error {
					return nil
				},
			},
		},
		{
			name: "no VM found by UUID - no update, no error",
			node: corev1.Node{
				ObjectMeta: testNodeMeta("worker-03"),
				Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{SystemUUID: "uuid-3"}},
			},
			expectedError: nil,
			mock: &MockProxmoxClient{
				GetVMByUUIDFn: func(ctx context.Context, uuid string) (*proxmox.VM, error) { return nil, nil },
				UpdateVMNameFn: func(ctx context.Context, nodeName string, vmid int, newName string) error {
					return nil
				},
			},
		},
		{
			name: "GetVMByUUID error bubbles up",
			node: corev1.Node{
				ObjectMeta: testNodeMeta("worker-04"),
				Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{SystemUUID: "uuid-4"}},
			},
			expectedError: proxmoxInternalErr,
			mock: &MockProxmoxClient{
				GetVMByUUIDFn: func(ctx context.Context, uuid string) (*proxmox.VM, error) { return nil, proxmoxInternalErr },
				UpdateVMNameFn: func(ctx context.Context, nodeName string, vmid int, newName string) error {
					return nil
				},
			},
		},
		{
			name: "UpdateVMName error bubbles up",
			node: corev1.Node{
				ObjectMeta: testNodeMeta("worker-05"),
				Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{SystemUUID: "uuid-5"}},
			},
			expectedError: proxmoxInternalErr,
			mock: &MockProxmoxClient{
				GetVMByUUIDFn: func(ctx context.Context, uuid string) (*proxmox.VM, error) {
					return &proxmox.VM{ID: 500, Name: "wrong-name", Node: "pve-5", UUID: "uuid-5"}, nil
				},
				UpdateVMNameFn: func(ctx context.Context, nodeName string, vmid int, newName string) error {
					return proxmoxInternalErr
				},
			},
		},
		{
			name: "control plane node is skipped (no update)",
			node: func() corev1.Node {
				n := corev1.Node{
					ObjectMeta: testNodeMeta("cp-01"),
					Status:     corev1.NodeStatus{NodeInfo: corev1.NodeSystemInfo{SystemUUID: "uuid-cp"}},
				}
				n.Labels["node-role.kubernetes.io/control-plane"] = ""
				return n
			}(),
			expectedError: nil,
			mock: &MockProxmoxClient{
				GetVMByUUIDFn: func(ctx context.Context, uuid string) (*proxmox.VM, error) {
					return &proxmox.VM{ID: 600, Name: "different", Node: "pve-6", UUID: "uuid-cp"}, nil
				},
				UpdateVMNameFn: func(ctx context.Context, nodeName string, vmid int, newName string) error {
					return nil
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.node.DeepCopy()).Build()
			r := NewNodeReconciler(c, scheme, tc.mock)

			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: tc.node.Name}}
			res, err := r.Reconcile(t.Context(), req)

			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, requeueDuration, res.RequeueAfter)
		})
	}
}

func testNodeMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   name,
		Labels: map[string]string{},
	}
}
