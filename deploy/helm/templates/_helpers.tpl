{{/*
Expand the name of the chart.
*/}}
{{- define "savvy-system.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "savvy-system.fullname" -}}
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
{{- define "savvy-system.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "savvy-system.labels" -}}
helm.sh/chart: {{ include "savvy-system.chart" . }}
{{ include "savvy-system.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "savvy-system.selectorLabels" -}}
app.kubernetes.io/name: {{ include "savvy-system.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "savvy-system.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "savvy-system.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database URL
*/}}
{{- define "savvy-system.databaseUrl" -}}
{{- if .Values.database.external.enabled }}
{{- printf "postgres://%s:$(DATABASE_PASSWORD)@%s:%d/%s?sslmode=%s" .Values.database.external.user .Values.database.external.host (.Values.database.external.port | int) .Values.database.external.name .Values.database.external.sslMode }}
{{- else if .Values.database.postgresql.enabled }}
{{- printf "postgres://%s:$(DATABASE_PASSWORD)@%s-postgresql:5432/%s?sslmode=disable" .Values.database.postgresql.auth.username (include "savvy-system.fullname" .) .Values.database.postgresql.auth.database }}
{{- else }}
{{- fail "Either database.external.enabled or database.postgresql.enabled must be true" }}
{{- end }}
{{- end }}
