# Check My Web

Check your website using various sercices directly inside your CI.

## Available Services

{{ range $key, $service := . }}
### [{{$key}}](service/{{$key}})

{{$service.Desc}}

{{ end }}

_Each service serve it's own documentation/examples_
