{{/*
Expand the name of the chart.
*/}}
{{- define "x-pdb.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-"}}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "x-pdb.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | replace "." "-" | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Namespace for all resources to be installed into
If not defined in values file then the helm release namespace is used
By default this is not set so the helm release namespace will be used

This gets around an problem within helm discussed here
https://github.com/helm/helm/issues/5358
*/}}
{{- define "x-pdb.namespace" -}}
    {{ .Values.namespace | default .Release.Namespace }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "x-pdb.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "." "-" | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
State labels
*/}}
{{- define "x-pdb.stateLabels" -}}
helm.sh/chart: {{ include "x-pdb.chart" . }}
{{ include "x-pdb.stateSelectorLabels" . -}}
{{ with .Values.state.extraLabels }}
{{- toYaml . }}
{{- end }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
State Selector labels
*/}}
{{- define "x-pdb.stateSelectorLabels" -}}
app.kubernetes.io/name: {{ include "x-pdb.name" . }}-state
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Webhook labels
*/}}
{{- define "x-pdb.webhookLabels" -}}
helm.sh/chart: {{ include "x-pdb.chart" . }}
{{ include "x-pdb.webhookSelectorLabels" . -}}
{{ with .Values.webhook.extraLabels }}
{{- toYaml . }}
{{- end }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Webhook Selector labels
*/}}
{{- define "x-pdb.webhookSelectorLabels" -}}
app.kubernetes.io/name: {{ include "x-pdb.name" . }}-webhook
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "x-pdb.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "x-pdb.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
