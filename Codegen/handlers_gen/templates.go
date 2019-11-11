package main

var prerequsites = `
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type srvRes struct {
	Error    string      ` + "`" + `json:"error"` + "`" + `
	Response interface{} ` + "`" + `json:"response,omitempty"` + "`" + `
}
`

var validatorTemplate = `
func (params *{{.Name}}) ParseAndValidate(reqValues url.Values) error {
	{{range .FieldsData -}}

	{{- if eq .Type "int" -}}
	var err error
	params.{{.Name}}, err = strconv.Atoi(reqValues.Get("
		{{- if keyInMap .Vals "paramname"}}{{.Vals.paramname}}{{else}}{{toLower .Name}}{{end -}}
	"))
	if err != nil {
		return fmt.Errorf("{{toLower .Name}} must be int")
	}
	
	{{if keyInMap .Vals "min" -}}
	if params.{{.Name}} < {{toInt .Vals.min}} {
		return fmt.Errorf("{{toLower .Name}} must be >= {{.Vals.min}}")
	}
	{{end}}

	{{- if keyInMap .Vals "max" -}}
	if params.{{.Name}} > {{toInt .Vals.max}} {
		return fmt.Errorf("{{toLower .Name}} must be <= {{.Vals.max}}")
	}
	{{end -}}

	{{- end -}}

	{{if eq .Type "string" -}}

	params.{{.Name}} = reqValues.Get("
		{{- if keyInMap .Vals "paramname"}}{{.Vals.paramname}}{{else}}{{toLower .Name}}{{end -}}
	")
	{{if keyInMap .Vals "default" -}}
	if params.{{.Name}} == "" {
		params.{{.Name}} = "{{.Vals.default}}"
	}
	{{end -}}

	{{- if keyInMap .Vals "required"}}
	if params.{{.Name}} == "" {
		return fmt.Errorf("{{toLower .Name}} must me not empty")
	}
	{{end -}}

	{{- if keyInMap .Vals "enum"}} {{$name := .Name}} {{$validVals := (split .Vals.enum "|")}}
	if{{range $i, $val := $validVals}} params.{{$name}} != "{{$val}}" {{if lt (add $i 1) (len $validVals)}}&&{{end}}{{end}} {
		return fmt.Errorf("{{toLower .Name}} must be one of [{{join $validVals ", "}}]")
	}
	{{end -}}

	{{- if keyInMap .Vals "min"}}
	if len(params.{{.Name}}) < {{toInt .Vals.min}} {
		return fmt.Errorf("{{toLower .Name}} len must be >= {{.Vals.min}}")
	}
	{{end -}}

	{{- if keyInMap .Vals "max"}}
	if len(params.{{.Name}}) > {{toInt .Vals.max}} {
		return fmt.Errorf("{{toLower .Name}} len must be <= {{.Vals.max}}")
	}
	{{end -}}
	{{end}}

	{{end -}}
	return nil
}
`

var sendErrorF = `
func SendError(w http.ResponseWriter, status int, mes string) {
	w.WriteHeader(status)
	paramErr, _ := json.Marshal(srvRes{Error: mes})
	w.Write(paramErr)
}
`

var handlerTemplate = `
func (srv *{{.ApiName}}) Handler{{.Name}}(w http.ResponseWriter, req *http.Request) {
	{{- if eq .Method "POST" }}
	if req.Method != http.MethodPost {
		SendError(w, http.StatusNotAcceptable, "bad method")
		return
	}
	{{end}}

	{{- if eq .Auth true}}
	if req.Header.Get("X-AUTH") != "100500" {
		SendError(w, http.StatusForbidden, "unauthorized")
		return
	}
	{{- end}}
	
	req.ParseForm()
	params := {{.ParamsName}}{}

	err := params.ParseAndValidate(req.Form)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := srv.{{.Name}}(req.Context(), params)

	if err == nil {
		w.WriteHeader(http.StatusOK)

		body, _ := json.Marshal(
			srvRes{Response: user},
		)
		w.Write(body)
	} else if e, ok := err.(ApiError); ok {
		SendError(w, e.HTTPStatus, e.Error())
	} else {
		SendError(w, http.StatusInternalServerError, err.Error())
	}
}
`

var serveTemplate = `
func (srv *{{.ApiName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
	{{- range .Cases }}
	case "{{index . 0}}":
		srv.Handler{{index . 1}}(w, r)
	{{- end}}
	default:
		SendError(w, http.StatusNotFound, "unknown method")
	}
}
`
