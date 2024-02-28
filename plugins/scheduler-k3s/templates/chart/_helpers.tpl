{{- define "print.annotations" }}
{{- if and .config.annotations (hasKey .config.annotations .key) }}
{{- $annotations := get .config.annotations .key }}
{{- range $k, $v := $annotations }}
{{ $k }}: {{ $v | quote }}
{{- end }}
{{- end }}
{{- end }}

{{- define "print.labels" }}
{{- if and .config.labels (hasKey .config.labels .key) }}
{{- $labels := get .config.labels .key }}
{{- range $k, $v := $labels }}
{{ $k }}: {{ $v | quote }}
{{- end }}
{{- end }}
{{- end }}
