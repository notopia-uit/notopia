{{- define "notopia-common.memoryEnv" -}}
{{- $runtime := default "" .Values.runtime -}}
{{- if or (eq $runtime "go") (eq $runtime "node") -}}
{{- $quantity := toString (dig "resources" "limits" "memory" "" (merge (dict) .Values)) -}}
{{- if not (regexMatch "^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)?$" $quantity) -}}{{ fail (printf "unsupported or missing memory limit %q" $quantity) }}{{- end -}}
{{- $digits := regexFind "^[0-9]+" $quantity -}}
{{- if gt (len $digits) 15 -}}{{ fail (printf "memory limit %q is too large" $quantity) }}{{- end -}}
{{- $number := int64 $digits -}}
{{- $suffix := trimPrefix $digits $quantity -}}
{{- $factors := dict "" 1 "Ki" 1024 "Mi" 1048576 "Gi" 1073741824 "Ti" 1099511627776 "K" 1000 "M" 1000000 "G" 1000000000 "T" 1000000000000 -}}
{{- if gt $number (div 1024819115206086200 (get $factors $suffix)) -}}{{ fail (printf "memory limit %q is too large" $quantity) }}{{- end -}}
{{- $bytes := mul $number (get $factors $suffix) -}}
{{- if eq $runtime "go" }}
- name: GOMEMLIMIT
  value: {{ div (mul $bytes 9) 10 | quote }}
{{- else }}
- name: NODE_OPTIONS
  value: {{ printf "--max-old-space-size=%d" (div (mul $bytes 3) 4194304) | quote }}
{{- end -}}{{- end -}}{{- end -}}
{{- define "notopia-common.otelEnv" -}}
{{- if .Values.otel.enabled }}
- name: OTEL_SERVICE_NAME
  value: {{ include "notopia-common.name" . | quote }}
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: {{ .Values.otel.endpoint | quote }}
- name: OTEL_EXPORTER_OTLP_PROTOCOL
  value: {{ .Values.otel.protocol | quote }}
- name: OTEL_K8S_NAMESPACE_NAME
  valueFrom: { fieldRef: { fieldPath: metadata.namespace } }
- name: OTEL_K8S_POD_NAME
  valueFrom: { fieldRef: { fieldPath: metadata.name } }
- name: OTEL_K8S_NODE_NAME
  valueFrom: { fieldRef: { fieldPath: spec.nodeName } }
- name: OTEL_RESOURCE_ATTRIBUTES
  value: {{ $attrs := list "k8s.namespace.name=$(OTEL_K8S_NAMESPACE_NAME)" "k8s.pod.name=$(OTEL_K8S_POD_NAME)" "k8s.node.name=$(OTEL_K8S_NODE_NAME)" }}{{ range $key, $value := .Values.otel.resourceAttributes }}{{ $attrs = append $attrs (printf "%s=%v" $key $value) }}{{ end }}{{ join "," $attrs | quote }}
{{- end -}}{{- end -}}
{{- define "notopia-common.env" -}}
{{- $all := list -}}
{{- range $name, $value := .Values.env -}}{{- $all = append $all (dict "name" $name "value" (toString $value)) -}}{{- end -}}
{{- range .Values.extraEnv -}}{{- $all = append $all . -}}{{- end -}}
{{- range (list (include "notopia-common.otelEnv" .) (include "notopia-common.memoryEnv" .)) -}}
{{- if . -}}{{- $all = concat $all (fromYamlArray .) -}}{{- end -}}{{- end -}}
{{- $seen := dict -}}{{- range $all -}}{{- if hasKey $seen .name -}}{{ fail (printf "duplicate environment variable %s" .name) }}{{- end -}}{{- $_ := set $seen .name true -}}{{- end -}}
{{- toYaml $all -}}
{{- end -}}
