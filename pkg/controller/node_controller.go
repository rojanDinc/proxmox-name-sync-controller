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

const requeueDuration = time.Second * 30
const proxmoxInternalErr = ProxmoxErr("proxmox internal error")

type ProxmoxErr string

func (pe ProxmoxErr) Error() string {
	return string(pe)
}

type ProxmoxClientInterface interface {
	GetVMByUUID(ctx context.Context, uuid string) (*proxmox.VM, error)
	UpdateVMName(ctx context.Context, nodeName string, vmid int, newName string) error
}

type NodeReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ProxmoxClient ProxmoxClientInterface
}

func NewNodeReconciler(k8sClient client.Client, scheme *runtime.Scheme, proxmoxClient ProxmoxClientInterface) *NodeReconciler {
	return &NodeReconciler{
		Client:        k8sClient,
		Scheme:        scheme,
		ProxmoxClient: proxmoxClient,
	}
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		logger.Info("Node not found, probably deleted", "node", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if r.isControlPlaneNode(&node) {
		logger.Info("Skipping control plane node", "node", node.Name)
		return ctrl.Result{RequeueAfter: requeueDuration}, nil
	}

	logger.Info("Reconciling node", "node", node.Name)
	vm, err := r.ProxmoxClient.GetVMByUUID(ctx, node.Status.NodeInfo.SystemUUID)
	if err != nil {
		logger.Error(err, "Failed to search for VM in Proxmox", "node", node.Name)
		return ctrl.Result{}, proxmoxInternalErr
	}

	if vm == nil {
		logger.Info("No corresponding VM found in Proxmox for node", "node", node.Name)
		return ctrl.Result{RequeueAfter: requeueDuration}, nil
	}

	if vm.Name == node.Name {
		logger.Info("VM name already matches node name", "node", node.Name, "vmid", vm.ID)
		return ctrl.Result{RequeueAfter: requeueDuration}, nil
	}

	logger.Info("Updating VM name to match node name",
		"node", node.Name,
		"vmid", vm.ID,
		"currentVMName", vm.Name,
		"newVMName", node.Name)

	if err := r.ProxmoxClient.UpdateVMName(ctx, vm.Node, vm.ID, node.Name); err != nil {
		logger.Error(err, "Failed to update VM name in Proxmox",
			"node", node.Name,
			"vmid", vm.ID)
		return ctrl.Result{}, proxmoxInternalErr
	}

	logger.Info("Successfully updated VM name in Proxmox",
		"node", node.Name,
		"vmid", vm.ID)

	return ctrl.Result{RequeueAfter: requeueDuration}, nil
}

func (r *NodeReconciler) isControlPlaneNode(node *corev1.Node) bool {
	if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; exists {
		return true
	}

	if _, exists := node.Labels["node-role.kubernetes.io/master"]; exists {
		return true
	}

	for _, taint := range node.Spec.Taints {
		if taint.Key == "node-role.kubernetes.io/control-plane" ||
			taint.Key == "node-role.kubernetes.io/master" {
			return true
		}
	}

	return false
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
