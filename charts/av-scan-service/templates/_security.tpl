{{/*
Template to generate av-scan-service Pod SecurityContext
*/}}
{{- define "avScanService.securityContext" -}}
{{- if .Values.avScanService.securityContext -}}
{{- $ctx := .Values.avScanService.securityContext -}}
{{- if and (eq .Values.platform "kubernetes") (not (hasKey $ctx "runAsUser")) -}}
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
{{- if and (eq .Values.platform "kubernetes") (not (hasKey $ctx "runAsUser")) -}}
{{- $_ := set $ctx "runAsUser" 100 -}}
{{- end -}}
{{- toYaml $ctx | nindent 0 }}
{{- end -}}
{{- end -}}
