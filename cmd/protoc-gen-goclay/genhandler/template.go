package genhandler

import (
	"bytes"
	"text/template"

	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/pkg/errors"
)

var (
	errNoTargetService = errors.New("no target service defined in the file")
)

type param struct {
	*descriptor.File
	Imports          []descriptor.GoPackage
	SwagBuffer       []byte
	EmitJSONDefaults bool
}

func applyTemplate(p param) (string, error) {
	w := bytes.NewBuffer(nil)
	if err := headerTemplate.Execute(w, p); err != nil {
		return "", err
	}

	if err := regTemplate.Execute(w, p); err != nil {
		return "", err
	}

	if err := footerTemplate.Execute(w, string(p.SwagBuffer)); err != nil {
		return "", err
	}

	if err := patternsTemplate.Execute(w, p); err != nil {
		return "", err
	}

	return w.String(), nil
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-goclay
// source: {{.GetName}}
// DO NOT EDIT!

/*
Package {{.GoPkg.Name}} is a self-registering gRPC and JSON+Swagger service definition.

It conforms to the github.com/utrack/clay Service interface.
*/
package {{.GoPkg.Name}}
import (
	{{range $i := .Imports}}{{if $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}

	{{range $i := .Imports}}{{if not $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}
)

`))
	regTemplate = template.Must(template.New("svc-reg").Parse(`
{{range $svc := .Services}}
// {{$svc.GetName}}Desc is a descriptor/registrator for the {{$svc.GetName}}Server.
type {{$svc.GetName}}Desc struct {
      svc {{$svc.GetName}}Server
}

// New{{$svc.GetName}}ServiceDesc creates new registrator for the {{$svc.GetName}}Server.
func New{{$svc.GetName}}ServiceDesc(svc {{$svc.GetName}}Server) *{{$svc.GetName}}Desc {
      return &{{$svc.GetName}}Desc{svc:svc}
}

// RegisterGRPC implements service registrator interface.
func (d *{{$svc.GetName}}Desc) RegisterGRPC(s *grpc.Server) {
      Register{{$svc.GetName}}Server(s,d.svc)
}

// SwaggerDef returns this file's Swagger definition.
func (d *{{$svc.GetName}}Desc) SwaggerDef() []byte {
      return _swaggerDef
}

// RegisterHTTP registers this service's HTTP handlers/bindings.
func (d *{{$svc.GetName}}Desc) RegisterHTTP(mux transport.Router) {
	{{range $m := $svc.Methods}}
	// Handlers for {{$m.GetName}}
	{{range $b := $m.Bindings}}
	mux.HandleFunc("/"+pattern_goclay_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, func(w http.ResponseWriter, r *http.Request) {
	  //TODO only POST is supported atm
	  var req {{$m.RequestType.GetName}}
	  err := jsonpb.Unmarshal(r.Body, &req)
	  if err != nil {
	    httpruntime.SetError(r.Context(),r,w,errors.Wrap(err,"couldn't read request JSON"),nil)
	    return
	  }
	  ret,err := d.svc.{{$m.GetName}}(r.Context(),&req)
	  if err != nil {
	    httpruntime.SetError(r.Context(),r,w,errors.Wrap(err,"returned from handler"),nil)
	    return
	  }

	  err = _{{$svc.GetName}}_pbMarshaler.Marshal(w, ret)
	  if err != nil {
	    httpruntime.SetError(r.Context(),r,w,errors.Wrap(err,"couldn't write response"),nil)
	    return
	  }
      })
      {{end}}
      {{end}}
}
var _{{$svc.GetName}}_pbMarshaler = &jsonpb.Marshaler{
      EmitDefaults: {{$.EmitJSONDefaults}},
}
{{end}}
`))
	footerTemplate = template.Must(template.New("footer").Parse(`
var _swaggerDef = []byte(` + "`" + `{{.}}` + `
` + "`)" + `
`))
	patternsTemplate = template.Must(template.New("patterns").Parse(`
var (
{{range $svc := .Services}}
{{range $m := $svc.Methods}}
{{range $b := $m.Bindings}}
	pattern_goclay_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}} = strings.Join({{$b.PathTmpl.Pool | printf "%#v"}},"/")
{{end}}
{{end}}
{{end}}
)
`))
)
