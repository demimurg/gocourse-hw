package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type apiReq struct {
	Name       string
	ParamsName string
	ApiName    string
	URL        string `json:"url"`
	Auth       bool   `json:"auth"`
	Method     string `json:"method"`
}

type paramInfo struct {
	Name       string
	FieldsData []fieldData
}

type fieldData struct {
	Name, Type string
	Vals       map[string]string
}

var pms = []paramInfo{}
var reqs = []apiReq{}

func init() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	getType := func(el *ast.Field) string {
		expr := el.Type
		if starEx, ok := expr.(*ast.StarExpr); ok {
			expr = starEx.X
		}
		return expr.(*ast.Ident).Name
	}

	for _, decl := range f.Decls {
		f, ok := decl.(*ast.FuncDecl)
		if ok && strings.Contains(f.Doc.Text(), "apigen:api") {
			method := apiReq{
				Name:       f.Name.Name,
				ParamsName: getType(f.Type.Params.List[1]),
				ApiName:    getType(f.Recv.List[0]),
			}

			doc := f.Doc.Text()
			apiParams := doc[strings.Index(doc, "{"):]
			json.Unmarshal([]byte(apiParams), &method)

			reqs = append(reqs, method)
			continue
		}

		g, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structName := currType.Name.Name
			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok || !strings.Contains(structName, "Params") {
				continue
			}

			param := paramInfo{Name: structName}

			for _, field := range currStruct.Fields.List {
				if field.Tag == nil || !strings.Contains(field.Tag.Value, "apivalidator") {
					continue
				}
				fData := fieldData{
					Name: field.Names[0].Name,
					Type: getType(field),
					Vals: map[string]string{},
				}

				tag := field.Tag.Value
				valsRaw := tag[strings.Index(tag, `"`)+1 : strings.LastIndex(tag, `"`)]

				for _, val := range strings.Split(valsRaw, ",") {
					if i := strings.Index(val, "="); i == -1 {
						fData.Vals[val] = ""
					} else {
						fData.Vals[val[:i]] = val[i+1:]
					}
				}

				param.FieldsData = append(param.FieldsData, fData)
			}

			pms = append(pms, param)
		}
	}
}

type router struct {
	ApiName string
	Cases   [][]string
}

var serveData = []router{}

func main() {
	var (
		req  apiReq
		rout router
		j    int
	)
	for _, req = range reqs {
		apiExist := false
		for j, rout = range serveData {
			if req.ApiName == rout.ApiName {
				apiExist = true
				break
			}
		}

		if apiExist {
			serveData[j].Cases = append(
				serveData[j].Cases,
				[]string{req.URL, req.Name},
			)
		} else {
			serveData = append(serveData, router{
				ApiName: req.ApiName,
				Cases: [][]string{
					[]string{req.URL, req.Name},
				},
			})
		}
	}

	file, err := os.Create(os.Args[2])
	defer file.Close()
	if err != nil {
		log.Fatal(fmt.Sprintln("File can't be created, error: ", err))
	}

	fmt.Fprintln(file, prerequsites, sendErrorF)

	tmpl := template.Must(template.New("validate").Funcs(
		template.FuncMap{
			"split":   strings.Split,
			"join":    strings.Join,
			"toLower": strings.ToLower,
			"toStr":   strconv.Itoa,
			"add":     func(n, m int) int { return n + m },
			"toInt": func(s string) int {
				digit, _ := strconv.Atoi(s)
				return digit
			},
			"keyInMap": func(store map[string]string, key string) bool {
				if _, ok := store[key]; ok {
					return true
				}
				return false
			},
		},
	).Parse(validatorTemplate))
	for _, param := range pms {
		tmpl.Execute(file, param)
	}

	tmpl = template.Must(template.New("handler").Funcs(
		template.FuncMap{
			"replaceAll": strings.ReplaceAll,
		},
	).Parse(handlerTemplate))
	for _, reqData := range reqs {
		tmpl.Execute(file, reqData)
	}

	tmpl = template.Must(template.New("serveHTTP").Parse(serveTemplate))
	for _, rout := range serveData {
		tmpl.Execute(file, rout)
	}
}
