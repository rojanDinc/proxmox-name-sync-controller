# Proxmox Name Sync Controller Helm Chart

This Helm chart deploys the Proxmox Name Sync Controller on a Kubernetes cluster using the Helm package manager.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Access to a Proxmox VE cluster with API access

## Deploying the Chart

### 1. Provide Proxmox configuration

You can either let the chart create the Secret from values (default), or provide an existing Secret containing `proxmox.yaml`.

**Option A: Use an existing Secret (managed outside Helm):**

Create a Secret with a key named `proxmox.yaml`:

```yaml
# proxmox.yaml
hostUrls:
  - "https://pve.example.com:8006"
tokenId: "root@pam!k8s-controller" # API tokens are recommended over username/password.
# username: user1
# password: mypassword
secret: "your-api-token-secret"
insecure: false
```

```bash
kubectl create secret generic my-proxmox-config \
  --from-file=proxmox.yaml=./proxmox.yaml
```

Reference it in values:

```yaml
proxmox:
  createSecret: false
  existingSecret: my-proxmox-config
```

**Option B: Let the chart create the Secret (from values) — API Token:**

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

**Option B: Let the chart create the Secret (from values) — Username/Password:**

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

### 2. Install the chart

```bash
helm install proxmox-name-sync-controller oci://ghcr.io/rojandinc/charts/proxmox-name-sync-controller
```

### 4. Verify the installation

```bash
# Check pod status
kubectl get pods -n default -l app.kubernetes.io/name=proxmox-name-sync-controller

# Check logs
kubectl logs -f deployment/proxmox-name-sync-controller
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

