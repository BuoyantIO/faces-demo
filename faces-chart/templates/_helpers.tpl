{{- define "partials.default-image" -}}
  {{- if .root.Values.defaultImage -}}
    {{- .root.Values.defaultImage -}}
  {{- else -}}
    {{- .root.Values.defaultRegistry -}}/{{- .root.Values.defaultImageName -}}:{{- .root.Values.defaultImageTag -}}
  {{- end -}}
{{- end -}}

{{- define "partials.image-tag" -}}
  {{- if .source.imageTag -}}
    {{- .source.imageTag -}}
  {{- else if (and .default .default.imageTag) }}
    {{- .default.imageTag -}}
  {{- else if .root.Values.defaultImageTag -}}
    {{- .root.Values.defaultImageTag -}}
  {{- else -}}
    {{- .root.Chart.AppVersion -}}
  {{- end -}}
{{- end -}}

{{- define "partials.select-image" -}}
  {{- if .source.image -}}
    {{- .source.image -}}
  {{- else if .source.imageName -}}
    {{- .source.imageName -}}:{{- include "partials.image-tag" . -}}
  {{- else if .default -}}
    {{- include "partials.select-image" (dict "source" .default "root" .root) -}}
  {{- else -}}
    {{- include "partials.default-image" . -}}
  {{- end -}}
{{- end -}}

{{- define "partials.select-imagePullPolicy" -}}
  {{- if .source.imagePullPolicy -}}
    {{- .source.imagePullPolicy -}}
  {{- else if .default -}}
    {{- include "partials.select-imagePullPolicy" (dict "source" .default "root" .root) -}}
  {{- else -}}
    {{ .root.Values.defaultImagePullPolicy -}}
  {{- end -}}
{{- end -}}

{{- define "partials.select-delayBuckets" -}}
  {{ $buckets := "" }}
  {{- if .source.delayBuckets -}}
    {{- $buckets = .source.delayBuckets -}}
  {{- else if (and .default .default.delayBuckets) -}}
    {{- $buckets = .default.delayBuckets -}}
  {{- end -}}
  {{- if $buckets }}
        - name: DELAY_BUCKETS
          value: {{ $buckets | quote }}
  {{- end -}}
{{- end -}}

{{- define "partials.select-errorFraction" -}}
  {{ $fraction := "" }}
  {{- if .source.errorFraction -}}
    {{- $fraction = .source.errorFraction -}}
  {{- else if (and .default .default.errorFraction) -}}
    {{- $fraction = .default.errorFraction -}}
  {{- end -}}
  {{- if $fraction }}
        - name: ERROR_FRACTION
          value: {{ $fraction | quote }}
  {{- end -}}
{{- end -}}

# Use all the above to provide nicer helpers for each of the workloads...
{{- define "partials.gui-image" -}}
  {{- include "partials.select-image" (dict "source" .Values.gui "root" .) -}}
{{- end -}}

{{- define "partials.gui-imagePullPolicy" -}}
  {{- include "partials.select-imagePullPolicy" (dict "source" .Values.gui "root" .) -}}
{{- end -}}

{{- define "partials.face-image" -}}
  {{- include "partials.select-image" (dict "source" .Values.face "root" .) -}}
{{- end -}}

{{- define "partials.face-imagePullPolicy" -}}
  {{- include "partials.select-imagePullPolicy" (dict "source" .Values.face "root" .) -}}
{{- end -}}

{{- define "partials.face-delayBuckets" -}}
  {{- include "partials.select-delayBuckets" (dict "source" .Values.face) -}}
{{- end -}}

{{- define "partials.face-errorFraction" -}}
  {{- include "partials.select-errorFraction" (dict "source" .Values.face) -}}
{{- end -}}

{{- define "partials.color-image" -}}
  {{- include "partials.select-image" (dict "source" .Values.color "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.color-imagePullPolicy" -}}
  {{- include "partials.select-imagePullPolicy" (dict "source" .Values.color "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.color-delayBuckets" -}}
  {{- include "partials.select-delayBuckets" (dict "source" .Values.color "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.color-errorFraction" -}}
  {{- include "partials.select-errorFraction" (dict "source" .Values.color "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.color2-image" -}}
  {{- include "partials.select-image" (dict "source" .Values.color2 "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.color2-imagePullPolicy" -}}
  {{- include "partials.select-imagePullPolicy" (dict "source" .Values.color2 "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.color2-delayBuckets" -}}
  {{- include "partials.select-delayBuckets" (dict "source" .Values.color2 "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.color2-errorFraction" -}}
  {{- include "partials.select-errorFraction" (dict "source" .Values.color2 "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.smiley-image" -}}
  {{- include "partials.select-image" (dict "source" .Values.smiley "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.smiley-imagePullPolicy" -}}
  {{- include "partials.select-imagePullPolicy" (dict "source" .Values.smiley "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.smiley-delayBuckets" -}}
  {{- include "partials.select-delayBuckets" (dict "source" .Values.smiley "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.smiley-errorFraction" -}}
  {{- include "partials.select-errorFraction" (dict "source" .Values.smiley "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.smiley2-image" -}}
  {{- include "partials.select-image" (dict "source" .Values.smiley2 "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.smiley2-imagePullPolicy" -}}
  {{- include "partials.select-imagePullPolicy" (dict "source" .Values.smiley2 "default" .Values.backend "root" .) -}}
{{- end -}}

{{- define "partials.smiley2-delayBuckets" -}}
  {{- include "partials.select-delayBuckets" (dict "source" .Values.smiley2 "default" .Values.backend) -}}
{{- end -}}

{{- define "partials.smiley2-errorFraction" -}}
  {{- include "partials.select-errorFraction" (dict "source" .Values.smiley2 "default" .Values.backend) -}}
{{- end -}}
