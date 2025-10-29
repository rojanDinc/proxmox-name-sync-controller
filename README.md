# Proxmox Name Sync Controller

This Kubernetes controller synchronizes VM names in Proxmox with Kubernetes node names. When a node is added to the cluster or its name changes, the controller will update the corresponding VM name in Proxmox to match.

## Description

This controller has one simple but important job: it ensures that VM names in your Proxmox cluster match the names of your Kubernetes nodes. This is particularly useful in environments where VMs are provisioned with generic names but later join the Kubernetes cluster with more meaningful hostnames.

## Features

- Automatically detects new Kubernetes nodes
- Finds corresponding VMs in Proxmox using flexible matching
- Updates VM names to match node names
- Skips control plane nodes (configurable)
- Supports both API token and username/password authentication
- Handles multiple Proxmox nodes/clusters

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

