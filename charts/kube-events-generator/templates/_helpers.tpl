{{/*
Find a kube-events image in various places.
Image can be found from:
* specified by user from .Values.image.repository and .Values.image.tag
* default value
*/}}
{{- define "kube-events.image" -}}
  {{- if and (not (empty .Values.image.repository)) (not (empty .Values.image.tag)) -}}
    {{- printf "%s:%s" .Values.image.repository .Values.image.tag -}}
  {{- else -}}
    {{- printf "ghcr.io/netcracker/qubership-kube-events-generator:main" -}}
  {{- end -}}
{{- end -}}