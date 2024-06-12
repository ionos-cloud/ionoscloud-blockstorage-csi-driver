{{/*
Expand the name of the chart.
*/}}
{{- define "csi-driver.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "csi-driver.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "csi-driver.labels" -}}
helm.sh/chart: {{ include "csi-driver.chart" . }}
{{ include "csi-driver.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "csi-driver.selectorLabels" -}}
app.kubernetes.io/name: {{ include "csi-driver.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Controller component labels
*/}}
{{- define "csi-driver.controllerComponentLabels" -}}
component: controller
{{- end }}

{{/*
Node component labels
*/}}
{{- define "csi-driver.nodeComponentLabels" -}}
component: node
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "csi-driver.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "csi-driver.name" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Render container args from extraArgs.
*/}}
{{- define "csi-driver.extraArgs" -}}
{{- range $key, $value := .extraArgs }}
{{- if not (kindIs "invalid" $value) }}
- --{{ $key | mustRegexFind "^[^_]+" }}={{ $value }}
{{- else }}
- --{{ $key | mustRegexFind "^[^_]+" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Render default storage class annotation.
*/}}
{{- define "csi-driver.storageClassAnnotation" -}}
{{- $root := index . 0 }}
{{- $name := index . 1 }}
{{- $defaultValue := index . 2 }}
{{- if $root.Release.IsInstall }}
storageclass.kubernetes.io/is-default-class: "{{ $defaultValue }}"
{{- else }}
{{- $annotations := ((lookup "storage.k8s.io/v1" "StorageClass" "" $name).metadata).annotations }}
storageclass.kubernetes.io/is-default-class: {{ get $annotations "storageclass.kubernetes.io/is-default-class" | quote }}
{{- end }}
{{- end }}

{{- define "image.registry" -}}
{{- if .Values.registry -}}
{{ .Values.registry }}/
{{- end -}}
{{- end -}}
