package printer

const listTemplate = `{{- $length := len .List.Items -}}{{- if eq $length 0 -}}
No {{ .List.Kind }} found
{{ else -}}
{{- if not $.NoHeaders -}}
{{- if $.AllNamespaces -}}
NAMESPACE	NAME	UID	STARTED	DURATION	STATUS
{{ else -}}
NAME	UID	STARTED	DURATION	STATUS
{{ end -}}
{{- end -}}
{{- range $_, $item := .List.Items }}{{- if $item }}{{- if $.AllNamespaces -}}
{{ $item.Namespace }}	{{ $item.Name }}	{{ $item.UID }}	{{ formatAge $item.Status.StartTime $.Time }}	{{ formatDuration $item.Status.StartTime $item.Status.CompletionTime }}	{{ formatCondition $item.Status.Conditions }}	{{ $item.UID }}
{{ else -}}
{{ $item.Name }}	{{ $item.UID }}	{{ formatAge $item.Status.StartTime $.Time }}	{{ formatDuration $item.Status.StartTime $item.Status.CompletionTime }}	{{ formatCondition $item.Status.Conditions }}
{{ end -}}{{- end -}}{{- end -}}
{{- end -}}`
