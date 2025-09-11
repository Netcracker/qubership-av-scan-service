{{/*
Expand the name of the chart.
*/}}
{{- define "av-scan-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "av-scan-service.fullname" -}}
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
{{- define "av-scan-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "av-scan-service.labels" -}}
helm.sh/chart: {{ include "av-scan-service.chart" . }}
{{ include "av-scan-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
name: {{ include "av-scan-service.name" . }}
app.kubernetes.io/part-of: "av-scan-service"
app.kubernetes.io/technology: go
{{- end }}

{{/*
Selector labels
*/}}
{{- define "av-scan-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "av-scan-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "av-scan-service.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "av-scan-service.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Liveness probe for av-scan-service
*/}}
{{- define "av-scan-service.livenessProbe" -}}
{{ omit .Values.avScanService.livenessProbe "httpGet" | toYaml }}
{{- if .Values.avScanService.livenessProbe.httpGet }}
httpGet: 
{{ toYaml .Values.avScanService.livenessProbe.httpGet | indent 2 }}
{{- if not .Values.avScanService.livenessProbe.httpGet.scheme }}
  scheme: {{ ternary "HTTPS" "HTTP" .Values.tls.enabled }}
{{- end }}
{{- if not .Values.avScanService.livenessProbe.httpGet.port }}
{{- if .Values.tls.enabled }}
  port: https
{{- else }}
  port: http
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Readiness probe for av-scan-service
*/}}
{{- define "av-scan-service.readinessProbe" -}}
{{ omit .Values.avScanService.readinessProbe "httpGet" | toYaml }}
{{- if .Values.avScanService.readinessProbe.httpGet }}
httpGet: 
{{ toYaml .Values.avScanService.readinessProbe.httpGet | indent 2 }}
{{- if not .Values.avScanService.readinessProbe.httpGet.scheme }}
  scheme: {{ ternary "HTTPS" "HTTP" .Values.tls.enabled }}
{{- end }}
{{- if not .Values.avScanService.livenessProbe.httpGet.port }}
{{- if .Values.tls.enabled }}
  port: https
{{- else }}
  port: http
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{/*
DNS names used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "av-scan-service.certDnsNames" -}}
  {{- $dnsNames := list "localhost"  (include "av-scan-service.fullname" .) (printf "%s.%s" (include "av-scan-service.fullname" .) .Release.Namespace)  (printf "%s.%s.svc" (include "av-scan-service.fullname" .) .Release.Namespace) -}}
  {{- $dnsNames = concat $dnsNames .Values.tls.generateCerts.subjectAlternativeName.additionalDnsNames -}}
  {{- $dnsNames | toYaml -}}
{{- end -}}

{{/*
IP addresses used to generate SSL certificate with "Subject Alternative Name" field
*/}}
{{- define "av-scan-service.certIpAddresses" -}}
  {{- $ipAddresses := list "127.0.0.1" -}}
  {{- $ipAddresses = concat $ipAddresses .Values.tls.generateCerts.subjectAlternativeName.additionalIpAddresses -}}
  {{- $ipAddresses | toYaml -}}
{{- end -}}
