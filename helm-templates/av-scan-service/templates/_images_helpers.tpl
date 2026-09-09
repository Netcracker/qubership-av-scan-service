{{- define "antivirus.image" -}}
  {{ printf "%s" .Values.avScanService.image }}
{{- end -}}

{{- define "clamav.image" -}}
  {{ printf "%s" .Values.clamav.image }}
{{- end -}}