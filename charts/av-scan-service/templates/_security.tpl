{{/*
Helper to determine the platform
Reads PAAS_PLATFORM from values (set by deployer via values.schema)
Converts to lowercase for consistency
Returns: "openshift" or "kubernetes"
Default: "openshift" (safe default that works on OpenShift)
*/}}
{{- define "effectivePlatform" -}}
{{- .Values.PAAS_PLATFORM | default "OPENSHIFT" | lower -}}
{{- end -}}

{{/*
Template to generate av-scan-service Pod SecurityContext
*/}}
{{- define "avScanService.securityContext" -}}
{{- if .Values.avScanService.securityContext -}}
{{- $ctx := .Values.avScanService.securityContext -}}
{{- $effectivePlatform := include "effectivePlatform" . | trim -}}
{{- if and (eq $effectivePlatform "kubernetes") (not (hasKey $ctx "runAsUser")) -}}
{{- $_ := set $ctx "runAsUser" 100 -}}
{{- end -}}
{{- toYaml $ctx | nindent 0 }}
{{- end -}}
{{- end -}}

{{/*
Template to generate clamav Pod SecurityContext
*/}}
{{- define "clamav.securityContext" -}}
{{- if .Values.clamav.securityContext -}}
{{- $ctx := .Values.clamav.securityContext -}}
{{- $effectivePlatform := include "effectivePlatform" . | trim -}}
{{- if and (eq $effectivePlatform "kubernetes") (not (hasKey $ctx "runAsUser")) -}}
{{- $_ := set $ctx "runAsUser" 100 -}}
{{- end -}}
{{- toYaml $ctx | nindent 0 }}
{{- end -}}
{{- end -}}
