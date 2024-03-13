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

{{- define "primary.port" -}}
{{- $found := dict -}}
{{- range $idx, $port_map := . -}}
{{- if hasPrefix "https-443-" $port_map.name -}}
{{- $found := set $found "https-443" $port_map.name -}}
{{- end -}}
{{- if hasPrefix "http-80-" $port_map.name -}}
{{- $found := set $found "http-80" $port_map.name -}}
{{- end -}}
{{- if and (eq "https" $port_map.scheme) (not (hasKey $found "https")) -}}
{{- $found := set $found "https" $port_map.name -}}
{{- end -}}
{{- if and (eq "http" $port_map.scheme) (not (hasKey $found "http")) -}}
{{- $found := set $found "http" $port_map.name -}}
{{- end -}}
{{- if not (hasKey $found "any") -}}
{{- $found := set $found "any" $port_map.name -}}
{{- end -}}
{{- end -}}
{{- if hasKey $found "https-443" -}}
{{- get $found "https-443" -}}
{{- else if hasKey $found "http-80" -}}
{{- get $found "http-80" -}}
{{- else if hasKey $found "https" -}}
{{- get $found "https" -}}
{{- else if hasKey $found "http" -}}
{{- get $found "http" -}}
{{- else if hasKey $found "any" -}}
{{- get $found "any" -}}
{{- end -}}
{{- end -}}
