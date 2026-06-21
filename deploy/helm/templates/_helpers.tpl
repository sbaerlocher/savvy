{{/*
Expand the name of the chart.
*/}}
{{- define "savvy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "savvy.fullname" -}}
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
{{- define "savvy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "savvy.labels" -}}
helm.sh/chart: {{ include "savvy.chart" . }}
{{ include "savvy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: savvy
{{- end }}
{{/*
component label intentionally omitted from common labels: per-resource
templates set their own app.kubernetes.io/component (config, secrets,
backend, ...). Hardcoding it here produced a duplicate map key wherever a
resource also declared its own — invalid YAML, caught by helm-template
rendered-manifest validation.
*/}}

{{/*
Selector labels
*/}}
{{- define "savvy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "savvy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "savvy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "savvy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
