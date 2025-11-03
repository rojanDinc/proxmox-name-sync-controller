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
{{- .Release.Namespace }}
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
{{- if and .Values.proxmox.secret.create .Values.proxmox.existingSecret }}
{{- fail "proxmox.secret.create cannot be true when proxmox.existingSecret is set" }}
{{- end }}
{{- if .Values.proxmox.secret.create }}
{{- $secret := .Values.proxmox.secret | default dict }}
{{- $hasHostUrls := $secret.hostUrls }}
{{- $hasUrl := $secret.url }}
{{- if and (not $hasHostUrls) (not $hasUrl) }}
{{- fail "Proxmox URL is required when proxmox.secret.create=true. Set proxmox.secret.hostUrls or proxmox.secret.url in values.yaml" }}
{{- end }}
{{- $hasToken := and $secret.tokenId $secret.secret }}
{{- $hasPassword := and $secret.username $secret.password }}
{{- if not (or $hasToken $hasPassword) }}
{{- fail "Proxmox authentication is required. Set either tokenId/secret or username/password in values.yaml" }}
{{- end }}
{{- else }}
{{- if not .Values.proxmox.existingSecret }}
{{- fail "proxmox.existingSecret is required when proxmox.secret.create=false" }}
{{- end }}
{{- end }}
{{- end }}
