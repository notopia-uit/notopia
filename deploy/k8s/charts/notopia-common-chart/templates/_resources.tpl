{{- define "notopia-common.hookJob" -}}
{{- $root := .root -}}
{{- $hookValues := default dict (get (default dict $root.Values.hooks) .name) -}}
{{- $resources := default $root.Values.resources $hookValues.resources -}}
{{- $hookRoot := merge (dict) $root -}}
{{- $_ := set $hookRoot "Values" (set (deepCopy $root.Values) "resources" $resources) -}}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ printf "%s-%s" (include "notopia-common.fullname" $root | trunc (int (sub 62 (len .name))) | trimSuffix "-") .name }}
  labels:
    {{- include "notopia-common.labels" $root | nindent 4 }}
  annotations:
    helm.sh/hook: {{ .hook | quote }}
    helm.sh/hook-weight: {{ .weight | quote }}
    helm.sh/hook-delete-policy: before-hook-creation
spec:
  backoffLimit: 0
  activeDeadlineSeconds: {{ default 300 $hookValues.activeDeadlineSeconds }}
  template:
    spec:
      restartPolicy: Never
      serviceAccountName: {{ include "notopia-common.serviceAccountName" $root }}
      securityContext:
        {{- include "notopia-common.podSecurityContext" $root | nindent 8 }}
      {{- with $root.Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      containers:
        - name: {{ .name }}
          image: {{ include "notopia-common.image" $root | quote }}
          imagePullPolicy: {{ default "IfNotPresent" $root.Values.image.pullPolicy }}
          command:
            {{- toYaml .command | nindent 12 }}
          securityContext:
            {{- include "notopia-common.containerSecurityContext" $root | nindent 12 }}
          env:
            {{- include "notopia-common.env" $hookRoot | nindent 12 }}
          {{- with $root.Values.envFrom }}
          envFrom:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          resources:
            {{- toYaml $resources | nindent 12 }}
{{- end -}}
{{- define "notopia-common.istioWorkloadPolicy" -}}
{{- if .Values.istio.enabled }}
{{- if or (empty .Values.istio.allowedPrincipals) (empty .Values.istio.ports) }}
{{- fail "istio workload policy requires non-empty allowedPrincipals and ports" }}
{{- end }}
{{- range .Values.istio.allowedPrincipals }}
{{- if or (not (kindIs "string" .)) (empty (trim (toString .))) (regexMatch "\\s" (toString .)) }}
{{- fail "istio workload policy requires non-empty principal entries" }}
{{- end }}
{{- end }}
{{- range .Values.istio.ports }}
{{- if or (not (kindIs "string" .)) (not (regexMatch "^[0-9]{1,5}$" (toString .))) (lt (int .) 1) (gt (int .) 65535) }}
{{- fail "istio workload policy requires valid port entries from 1 to 65535" }}
{{- end }}
{{- end }}
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: {{ include "notopia-common.fullname" . }}
spec:
  selector:
    matchLabels:
      {{- include "notopia-common.selectorLabels" . | nindent 6 }}
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
              {{- toYaml .Values.istio.allowedPrincipals | nindent 14 }}
      to:
        - operation:
            ports:
              {{- toYaml .Values.istio.ports | nindent 14 }}
{{- end -}}
{{- end -}}
