# Proxmox Name Sync Controller Helm Chart

This Helm chart deploys the Proxmox Name Sync Controller on a Kubernetes cluster using the Helm package manager.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Access to a Proxmox VE cluster with API access

## Installing the Chart

### 1. Add the Helm Repository (if published)

```bash
helm repo add proxmox-sync https://your-registry.com/helm-charts
helm repo update
```

### 2. Create a values file with your Proxmox credentials

**For API Token authentication (recommended):**

```yaml
# my-values.yaml
image:
  repository: your-registry.com/proxmox-name-sync-controller
  tag: "v0.1.0"

proxmox:
  url: "https://pve.example.com:8006"
  tokenId: "root@pam!k8s-controller"
  secret: "your-api-token-secret"
```

**For Username/Password authentication:**

```yaml
# my-values.yaml
image:
  repository: your-registry.com/proxmox-name-sync-controller
  tag: "v0.1.0"

proxmox:
  url: "https://pve.example.com:8006"
  username: "root@pam"
  password: "your-password"
```

### 3. Install the chart

```bash
# Install from repository
helm install my-proxmox-sync proxmox-sync/proxmox-name-sync-controller -f my-values.yaml

# Or install from local chart
helm install my-proxmox-sync ./helm/proxmox-name-sync-controller -f my-values.yaml
```

### 4. Verify the installation

```bash
# Check pod status
kubectl get pods -n default -l app.kubernetes.io/name=proxmox-name-sync-controller

# Check logs
kubectl logs -f deployment/my-proxmox-sync-controller-manager

# Test node access
kubectl auth can-i get nodes --as=system:serviceaccount:default:my-proxmox-sync-proxmox-name-sync-controller
```

## Configuration

### Proxmox Authentication

#### API Token (Recommended)

1. Log into Proxmox web interface
2. Go to Datacenter → Permissions → API Tokens
3. Click "Add" and create a token
4. Uncheck "Privilege Separation" for full permissions
5. Copy the token ID and secret

```yaml
proxmox:
  url: "https://pve.example.com:8006"
  tokenId: "root@pam!mytoken"
  secret: "your-token-secret"
```

#### Username/Password

```yaml
proxmox:
  url: "https://pve.example.com:8006"
  username: "root@pam"
  password: "your-password"
```

### High Availability Configuration

For production deployments with multiple replicas:

```yaml
replicaCount: 2
controller:
  leaderElection:
    enabled: true

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - proxmox-name-sync-controller
        topologyKey: kubernetes.io/hostname
```

### Monitoring Configuration

Enable Prometheus monitoring:

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    labels:
      app: prometheus
```

### Security Configuration

For enhanced security:

```yaml
networkPolicy:
  enabled: true

securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `image.repository` | string | `""` | Container image repository |
| `image.tag` | string | `""` | Container image tag (defaults to chart appVersion) |
| `image.pullPolicy` | string | `"IfNotPresent"` | Image pull policy |
| `replicaCount` | int | `1` | Number of replicas |
| `controller.leaderElection.enabled` | bool | `false` | Enable leader election |
| `controller.logLevel` | string | `"info"` | Log level (debug, info, warn, error) |
| `controller.developmentMode` | bool | `false` | Enable development mode |
| `proxmox.url` | string | `""` | Proxmox server URL (required) |
| `proxmox.tokenId` | string | `""` | API token ID |
| `proxmox.secret` | string | `""` | API token secret |
| `proxmox.username` | string | `""` | Username (alternative to token) |
| `proxmox.password` | string | `""` | Password (alternative to token) |
| `proxmox.insecure` | bool | `true` | Accept self-signed certificates |
| `proxmox.createSecret` | bool | `true` | Create secret from values |
| `proxmox.existingSecret` | string | `""` | Use existing secret |
| `metrics.enabled` | bool | `true` | Enable metrics endpoint |
| `metrics.port` | int | `8080` | Metrics port |
| `metrics.serviceMonitor.enabled` | bool | `false` | Create ServiceMonitor |
| `rbac.create` | bool | `true` | Create RBAC resources |
| `serviceAccount.create` | bool | `true` | Create service account |
| `resources.limits.cpu` | string | `"500m"` | CPU limit |
| `resources.limits.memory` | string | `"128Mi"` | Memory limit |
| `resources.requests.cpu` | string | `"10m"` | CPU request |
| `resources.requests.memory` | string | `"64Mi"` | Memory request |
| `nodeSelector` | object | `{}` | Node selector |
| `tolerations` | list | `[]` | Tolerations |
| `affinity` | object | `{}` | Affinity rules |
| `networkPolicy.enabled` | bool | `false` | Enable network policy |

## Example Deployments

### Development Environment

```bash
helm install dev-proxmox-sync ./helm/proxmox-name-sync-controller \
  -f ./helm/proxmox-name-sync-controller/values-development.yaml \
  --set proxmox.url="https://pve-dev.local:8006" \
  --set proxmox.username="root@pam" \
  --set proxmox.password="devpassword"
```

### Production Environment

```bash
helm install prod-proxmox-sync ./helm/proxmox-name-sync-controller \
  -f ./helm/proxmox-name-sync-controller/values-production.yaml \
  --set image.repository="your-registry.com/proxmox-name-sync-controller" \
  --set image.tag="v0.1.0" \
  --set proxmox.url="https://pve.example.com:8006" \
  --set proxmox.tokenId="root@pam!k8s-controller" \
  --set proxmox.secret="your-secret-here"
```

## Upgrading

```bash
helm upgrade my-proxmox-sync ./helm/proxmox-name-sync-controller -f my-values.yaml
```

## Uninstalling

```bash
helm uninstall my-proxmox-sync
```

Note: This will not remove the ClusterRole and ClusterRoleBinding. To remove them:

```bash
kubectl delete clusterrole my-proxmox-sync-manager-role
kubectl delete clusterrolebinding my-proxmox-sync-manager-rolebinding
```

## Troubleshooting

### Common Issues

1. **Controller can't access nodes**
   ```bash
   kubectl auth can-i get nodes --as=system:serviceaccount:default:my-proxmox-sync-proxmox-name-sync-controller
   ```

2. **Proxmox connection failed**
   ```bash
   kubectl logs deployment/my-proxmox-sync-controller-manager
   ```

3. **Secret not found**
   ```bash
   kubectl get secrets | grep proxmox
   kubectl describe secret my-proxmox-sync-proxmox-credentials
   ```

### Debug Mode

Enable debug logging:

```yaml
controller:
  logLevel: debug
  developmentMode: true
```

### Testing Connectivity

```bash
# Port-forward to test locally
kubectl port-forward service/my-proxmox-sync-metrics-service 8080:8080

# Check metrics
curl http://localhost:8080/metrics

# Check health
curl http://localhost:8080/healthz
```
