{{- define "embedded-market-services.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: embedded-market
{{- end }}

{{- define "embedded-market-services.serviceLabels" -}}
application: service
app: {{ .name }}
tier: {{ .tier }}
{{- end }}
