{{/*
Template to generate av-scan-service Pod SecurityContext
*/}}
{{- define "avScanService.securityContext" -}}
  {{- if .Values.avScanService.securityContext -}}
    {{- toYaml .Values.avScanService.securityContext | nindent 6 }}
  {{- end }}
  {{- if not (.Capabilities.APIVersions.Has "apps.openshift.io/v1") }}
    {{- if not .Values.avScanService.securityContext.runAsUser }}
      runAsUser: 100
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Template to generate clamav Pod SecurityContext
*/}}
{{- define "clamav.securityContext" -}}
  {{- if .Values.clamav.securityContext -}}
    {{- toYaml .Values.clamav.securityContext | nindent 6 }}
  {{- end }}
  {{- if not (.Capabilities.APIVersions.Has "apps.openshift.io/v1") }}
    {{- if not .Values.clamav.securityContext.runAsUser }}
      runAsUser: 100
    {{- end }}
  {{- end }}
{{- end -}}