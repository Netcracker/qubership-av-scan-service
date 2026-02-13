{{/*
Helper to detect if running on OpenShift
*/}}
{{- define "isOpenShift" -}}
{{- if or (.Capabilities.APIVersions.Has "security.openshift.io/v1") (.Capabilities.APIVersions.Has "apps.openshift.io/v1") (.Capabilities.APIVersions.Has "route.openshift.io/v1") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Template to generate av-scan-service Pod SecurityContext
*/}}
{{- define "avScanService.securityContext" -}}
{{- if .Values.avScanService.securityContext -}}
{{- $ctx := .Values.avScanService.securityContext -}}
{{- $isOpenShift := include "isOpenShift" . | trim -}}
{{- if eq $isOpenShift "false" -}}
{{- if not (hasKey $ctx "runAsUser") -}}
{{- $_ := set $ctx "runAsUser" 100 -}}
{{- end -}}
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
{{- $isOpenShift := include "isOpenShift" . | trim -}}
{{- if eq $isOpenShift "false" -}}
{{- if not (hasKey $ctx "runAsUser") -}}
{{- $_ := set $ctx "runAsUser" 100 -}}
{{- end -}}
{{- end -}}
{{- toYaml $ctx | nindent 0 }}
{{- end -}}
{{- end -}}
