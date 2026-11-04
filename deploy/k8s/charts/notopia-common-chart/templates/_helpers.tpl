{{- define "notopia-common.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "notopia-common.fullname" -}}
{{- if .Values.fullnameOverride -}}{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}{{- else -}}
{{- $name := include "notopia-common.name" . -}}
{{- if contains $name .Release.Name -}}{{ .Release.Name | trunc 63 | trimSuffix "-" }}{{- else -}}{{ printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}{{- end -}}
{{- end -}}{{- end -}}
{{- define "notopia-common.chart" -}}{{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}{{- end -}}
{{- define "notopia-common.selectorLabels" -}}
app.kubernetes.io/name: {{ include "notopia-common.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "notopia-common.labels" -}}
helm.sh/chart: {{ include "notopia-common.chart" . }}
{{ include "notopia-common.selectorLabels" . }}
app.kubernetes.io/part-of: notopia
app.kubernetes.io/component: {{ include "notopia-common.name" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
{{- end -}}
{{- define "notopia-common.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}{{ default (include "notopia-common.fullname" .) .Values.serviceAccount.name }}{{- else -}}{{ default "default" .Values.serviceAccount.name }}{{- end -}}
{{- end -}}
{{- define "notopia-common.image" -}}
{{- if .Values.image.digest -}}{{ printf "%s@%s" .Values.image.repository .Values.image.digest }}{{- else -}}{{ printf "%s:%s" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) }}{{- end -}}
{{- end -}}
{{- define "notopia-common.podSecurityContext" -}}
{{- $context := .Values.podSecurityContext | default dict -}}
runAsNonRoot: true
runAsUser: {{ required "podSecurityContext.runAsUser is required" $context.runAsUser }}
runAsGroup: {{ required "podSecurityContext.runAsGroup is required" $context.runAsGroup }}
fsGroup: {{ required "podSecurityContext.fsGroup is required" $context.fsGroup }}
seccompProfile:
  type: RuntimeDefault
{{- end -}}
{{- define "notopia-common.containerSecurityContext" -}}
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop: [ALL]
{{- end -}}
