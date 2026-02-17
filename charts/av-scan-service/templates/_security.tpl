{{/*
Helper to detect if running on OpenShift
Checks for OpenShift-specific API groups
Returns "true" or "false"
*/}}
{{- define "isOpenShift" -}}
{{- if or (.Capabilities.APIVersions.Has "security.openshift.io/v1") 
         (.Capabilities.APIVersions.Has "apps.openshift.io/v1") 
         (.Capabilities.APIVersions.Has "route.openshift.io/v1") 
         (.Capabilities.APIVersions.Has "project.openshift.io/v1") 
         (.Capabilities.APIVersions.Has "image.openshift.io/v1") -}}
true
{{- else -}}
false
{{- end -}}
{{- end -}}

{{/*
Helper to determine the effective platform setting
Returns "openshift" or "kubernetes" based on:
1. Manual override (if platform is explicitly set)
2. Auto-detection (if platform is "auto" or not set)
*/}}
{{- define "effectivePlatform" -}}
{{- if and .Values.platform (ne .Values.platform "auto") -}}
{{- .Values.platform -}}
{{- else -}}
{{- $isOpenShift := include "isOpenShift" . | trim -}}
{{- if eq $isOpenShift "true" -}}
openshift
{{- else -}}
kubernetes
{{- end -}}
{{- end -}}
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
