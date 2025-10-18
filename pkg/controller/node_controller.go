package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/rojanDinc/proxmox-name-sync-controller/pkg/proxmox"
)

// ProxmoxClientInterface defines the interface for Proxmox client operations
type ProxmoxClientInterface interface {
	GetVMIDByName(ctx context.Context, nodeName string) (*proxmox.VM, error)
	UpdateVMName(ctx context.Context, nodeName string, vmid int, newName string) error
}

// NodeReconciler reconciles Node objects and syncs VM names in Proxmox
type NodeReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ProxmoxClient ProxmoxClientInterface
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=get

// Reconcile handles the reconciliation of Node resources
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Node instance
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		// Node was deleted, nothing to do
		logger.Info("Node not found, probably deleted", "node", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling node", "node", node.Name)

	// Skip if this is a control plane node (optional filter)
	if r.isControlPlaneNode(&node) {
		logger.Info("Skipping control plane node", "node", node.Name)
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
	}

	// Try to find corresponding VM in Proxmox
	vm, err := r.ProxmoxClient.GetVMIDByName(ctx, node.Name)
	if err != nil {
		logger.Error(err, "Failed to search for VM in Proxmox", "node", node.Name)
		return ctrl.Result{RequeueAfter: time.Minute * 2}, err
	}

	if vm == nil {
		logger.Info("No corresponding VM found in Proxmox for node", "node", node.Name)
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
	}

	// Check if VM name matches node name
	if vm.Name == node.Name {
		logger.Info("VM name already matches node name", "node", node.Name, "vmid", vm.ID)
		return ctrl.Result{RequeueAfter: time.Minute * 30}, nil
	}

	// Update VM name to match node name
	logger.Info("Updating VM name to match node name",
		"node", node.Name,
		"vmid", vm.ID,
		"currentVMName", vm.Name,
		"newVMName", node.Name)

	err = r.ProxmoxClient.UpdateVMName(ctx, vm.Node, vm.ID, node.Name)
	if err != nil {
		logger.Error(err, "Failed to update VM name in Proxmox",
			"node", node.Name,
			"vmid", vm.ID)
		return ctrl.Result{RequeueAfter: time.Minute * 2}, err
	}

	logger.Info("Successfully updated VM name in Proxmox",
		"node", node.Name,
		"vmid", vm.ID)

	// Requeue after successful update to verify the change
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// isControlPlaneNode checks if a node is a control plane node
func (r *NodeReconciler) isControlPlaneNode(node *corev1.Node) bool {
	// Check for common control plane labels
	if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; exists {
		return true
	}
	if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
		return true
	}

	// Check for control plane taints
	for _, taint := range node.Spec.Taints {
		if taint.Key == "node-role.kubernetes.io/control-plane" ||
			taint.Key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false
}

// SetupWithManager sets up the controller with the Manager
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

// NewNodeReconciler creates a new NodeReconciler instance
func NewNodeReconciler(client client.Client, scheme *runtime.Scheme, proxmoxClient ProxmoxClientInterface) *NodeReconciler {
	return &NodeReconciler{
		Client:        client,
		Scheme:        scheme,
		ProxmoxClient: proxmoxClient,
	}
}
