{{/*
Expand the name of the chart.
*/}}
{{- define "proxmox-name-sync-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "proxmox-name-sync-controller.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "proxmox-name-sync-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "proxmox-name-sync-controller.labels" -}}
helm.sh/chart: {{ include "proxmox-name-sync-controller.chart" . }}
{{ include "proxmox-name-sync-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "proxmox-name-sync-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "proxmox-name-sync-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "proxmox-name-sync-controller.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "proxmox-name-sync-controller.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the namespace to use
*/}}
{{- define "proxmox-name-sync-controller.namespace" -}}
{{- if .Values.namespace.name }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Create the secret name for Proxmox credentials
*/}}
{{- define "proxmox-name-sync-controller.secretName" -}}
{{- if .Values.proxmox.existingSecret }}
{{- .Values.proxmox.existingSecret }}
{{- else }}
{{- printf "%s-proxmox-credentials" (include "proxmox-name-sync-controller.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Validate Proxmox configuration
*/}}
{{- define "proxmox-name-sync-controller.validateProxmox" -}}
{{- if not .Values.proxmox.url }}
{{- fail "Proxmox URL is required. Set proxmox.url in values.yaml" }}
{{- end }}
{{- if .Values.proxmox.createSecret }}
{{- $hasToken := and .Values.proxmox.tokenId .Values.proxmox.secret }}
{{- $hasPassword := and .Values.proxmox.username .Values.proxmox.password }}
{{- if not (or $hasToken $hasPassword) }}
{{- fail "Proxmox authentication is required. Set either tokenId/secret or username/password in values.yaml" }}
{{- end }}
{{- end }}
{{- end }}
