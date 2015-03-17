package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const strTemplate = `
{{ define "Arguments" }}{{ range $i, $e := .Params }}{{ if $i }}, {{ end }} {{ $e.name }} {{ $e.type | checkElement }}{{ end }}{{ end }}
{{ $name := .Name}}

{{/* Generate interface then struct */}}
{{ if ne .Name "MediaObject" }}
type I{{ .Name }} interface {
	{{ range .Methods }}{{.Name | title }}({{ template "Arguments" .}})({{ if .Return.type}}{{.Return.type}},{{end}} error)
	{{end}}
}
{{ end }}

{{ .Doc }}
type {{ .Name }} struct {
	{{if eq .Name "MediaObject"}}connection *Connection{{else}}{{ .Extends }}{{end}}
	{{ range .Properties }}
	{{ .doc }}
	{{ .name | title }} {{ .type }}
	{{ end }}
}


// Return contructor params to be called by "Create".
func (elem *{{.Name}}) getConstructorParams(from IMediaObject, options map[string]interface{}) map[string]interface{} {
	{{ if len .Constructor.Params }}
	// Create basic constructor params
	ret := map[string]interface{} {
		{{ range .Constructor.Params }}{{ if eq .type "string" "float64" "boolean" "int" }}"{{ .name }}" : {{ .defaultValue }},
		{{ else }} "{{ .name }}" : fmt.Sprintf("%s", from),
		{{ end }}{{ end }}
	}

	// then merge options
	mergeOptions(ret, options)

	return ret
	{{ else }}return options
	{{ end }}
}

{{ range .Methods }}
{{ .Doc }}{{ if .Return.doc }}
// Returns: 
{{ .Return.doc }}{{ end }}
func (elem *{{$name}}) {{ .Name | title }}({{ template "Arguments" .}}) ({{if .Return.type }}{{ .Return.type }}, {{ end }} error) {
	req := elem.getInvokeRequest()
	{{ if .Params }}
	params := make(map[string]interface{})
	{{ range .Params }}
	setIfNotEmpty(params, "{{.name}}", {{.name}}){{ end }}
	{{ end }}

	req["params"] = map[string]interface{}{
		"operation" : "{{ .Name }}",
		"object"	: elem.Id,{{ if .Params }}
		"operationParams" : params,
		{{ end }}
	}

	// Call server and wait response
	response := <- elem.connection.Request(req)
	{{ if .Return}}
	{{ .Return.doc }}
		{{ if eq .Return.type "string" "int" "float64" "boolean" }}
	return response.Result["value"], response.Error
		{{ else }}{{/* More complicated but... let's go */}}
	ret := {{ .Return.type }}{}
	return ret, response.Error
		{{ end }}
	{{ else }}
	// Returns error or nil
	return response.Error
	{{end}}

}
{{ end }}
`

const complexTypeTemplate = `
{{ if eq .TypeFormat "ENUM" }}
{{ $name := .Name }}
{{ .Doc }}
type {{.Name}} string

// Implement fmt.Stringer interface
func (t {{.Name}}) String() string {
	return string(t)
}

const (
	{{ range .Values }}{{ $name | uppercase }}_{{ . }} {{ $name }} = "{{ . }}" 
	{{ end}}
)
{{ else }}

type {{ .Name }} struct {
	{{ range .Properties}}{{ .name | title }} {{ .type }}
	{{ end }}
}
{{ end }}
`

const packageTemplate = `package kurento
{{ .Content }}
`

const DOCLINELENGTH = 79

var re = regexp.MustCompile(`(.+)\[\]`)

var CPXTYPES = make([]string, 0)

type Return struct {
	Doc  string
	Type string
}

type constructor struct {
	Name   string
	Doc    string
	Params []map[string]interface{}
}

type method struct {
	constructor
	Return map[string]interface{}
}

type class struct {
	Name        string
	Extends     string
	Doc         string
	Abstract    bool
	Properties  []map[string]interface{}
	Events      []string
	Constructor constructor
	Methods     []method
}

type core struct {
	RemoteClasses []class
	ComplexTypes  []ComplexType
}

type ComplexType struct {
	TypeFormat string
	Doc        string
	Values     []string
	Name       string
	Properties []map[string]interface{}
}

const (
	CORE     = "kms-core/src/server/interface/core.kmd.json"
	ELEMENTS = "kms-elements/src/server/interface/"
)

// template func that change MediaXXX to IMediaXXX to
// be sure to work with interface.
// Set it global to be used by funcMap["paramValue"] above.
func tplCheckElement(p string) string {
	if len(p) > 5 && p[:5] == "Media" {
		if p[len(p)-4:] != "Type" {
			return "IMedia" + p[5:]
		}
	}
	return p
}

func isComplexType(t string) bool {
	for _, c := range CPXTYPES {
		if c == t {
			return true
		}
	}
	return false
}

var funcMap = template.FuncMap{
	"title":        strings.Title,
	"uppercase":    strings.ToUpper,
	"checkElement": tplCheckElement,
	"paramValue": func(p map[string]interface{}) string {
		name := p["name"].(string)
		t := p["type"].(string)
		t = tplCheckElement(t)

		ctype := isComplexType(t)
		switch t {
		case "float64", "int":
			return fmt.Sprintf("\"%s\" = %s", name, name)
		case "string", "boolean":
			return fmt.Sprintf("\"%s\" = %s", name, name)
		default:
			// If param is not complexType, we have Id from String() method
			if !ctype && t[0] == 'I' { /* TODO: fix isInterface */
				return fmt.Sprintf("\"%s\" = fmt.Sprintf(\"%%s\", %s)", name, name)
			}
		}
		// Default is to set value to param
		return fmt.Sprintf("\"%s\" = %s", name, name)
	},
}

func formatDoc(doc string) string {

	doc = strings.Replace(doc, ":rom:cls:", "", -1)
	doc = strings.Replace(doc, ":term:", "", -1)
	doc = strings.Replace(doc, "``", `"`, -1)

	lines := strings.Split(doc, "\n")
	part := make([]string, 0)
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		// if line is too long, cut !
		if len(line) > DOCLINELENGTH {
			pos := DOCLINELENGTH
			for len(line) > DOCLINELENGTH {
				// find previous space
				for i := pos; line[pos] != ' '; i-- {
					pos = i
				}
				part = append(part, line[:pos])
				line = line[pos:]
			}
		}
		// then append remaining line
		part = append(part, line)
	}

	for i, p := range part {
		part[i] = "// " + strings.TrimSpace(p)
	}
	ret := strings.Join(part, "\n")
	return ret
}

func formatTypes(p map[string]interface{}) map[string]interface{} {
	p["doc"] = formatDoc(p["doc"].(string))
	if p["type"] == "String[]" {
		p["type"] = "[]string"
	}

	if p["type"] == "String" {
		p["type"] = "string"
	}

	if p["type"] == "float" {
		p["type"] = "float64"
	}
	if p["type"] == "boolean" {
		p["type"] = "bool"
	}

	if re.MatchString(p["type"].(string)) {
		found := re.FindAllStringSubmatch(p["type"].(string), -1)
		p["type"] = "[]" + found[0][1]
	}

	if p["defaultValue"] == "" || p["defaultValue"] == nil {
		switch p["type"] {
		case "string":
			p["defaultValue"] = `""`
		case "bool":
			p["defaultValue"] = "false"
		case "int", "float64":
			p["defaultValue"] = "0"
		}
	}

	return p
}

func getModel(path string) core {

	i := core{}
	data, _ := ioutil.ReadFile(path)
	err := json.Unmarshal(data, &i)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func getInterfaces() {
	paths, _ := filepath.Glob(ELEMENTS + "elements.*.kmd.json")
	for _, p := range paths {
		r := getModel(p).RemoteClasses
		classes := parse(r)
		base := filepath.Base(p)
		base = strings.Replace(base, "elements.", "", -1)
		base = strings.Replace(base, ".kmd.json", "", -1)
		base = "kurento/" + base + ".go"
		writeFile(base, classes)
	}
}

func parse(c []class) []string {
	ret := make([]string, len(c))
	for idx, cl := range c {

		log.Println("Generating ", cl.Name)
		// rewrite types
		for j, p := range cl.Properties {
			p = formatTypes(p)
			switch p["type"] {
			case "string", "float64", "int", "bool", "[]string":
			default:
				if _, ok := p["type"].(string); ok {
					if p["type"].(string)[:2] == "[]" {
						t := p["type"].(string)[2:]
						if isComplexType(t) {
							p["type"] = "[]*" + t
						} else {
							p["type"] = "[]I" + t
						}
					} else {
						if isComplexType(p["type"].(string)) {
							p["type"] = "*" + p["type"].(string)
						} else {
							p["type"] = "I" + p["type"].(string)
						}
					}
				}
			}
			cl.Properties[j] = p
		}

		for j, m := range cl.Methods {
			for i, p := range m.Params {
				p := formatTypes(p)
				m.Params[i] = p
			}
			m.Doc = formatDoc(m.Doc)

			if m.Return["type"] != nil {
				m.Return = formatTypes(m.Return)
				m.Return["doc"] = formatDoc(m.Return["doc"].(string))
			}

			cl.Methods[j] = m

		}
		for j, p := range cl.Constructor.Params {
			p := formatTypes(p)
			cl.Constructor.Params[j] = p
		}

		tpl, _ := template.New("structure").Funcs(funcMap).Parse(strTemplate)
		buff := bytes.NewBufferString("")
		cl.Doc = formatDoc(cl.Doc)

		tpl.Execute(buff, cl)
		ret[idx] = buff.String()
	}
	return ret
}

func parseComplexTypes() {
	paths, _ := filepath.Glob("kms-elements/src/server/interface/elements.*.kmd.json")
	paths = append(paths, CORE)
	ret := make([]string, 0)
	for _, path := range paths {
		ctypes := getModel(path).ComplexTypes
		for _, ctype := range ctypes {

			// Add in list
			CPXTYPES = append(CPXTYPES, ctype.Name)

			ctype.Doc = formatDoc(ctype.Doc)

			for i, p := range ctype.Properties {
				ctype.Properties[i] = formatTypes(p)
			}

			buff := bytes.NewBufferString("")
			tpl, _ := template.New("complexttypes").Funcs(funcMap).Parse(complexTypeTemplate)
			tpl.Execute(buff, ctype)
			ret = append(ret, buff.String())
		}
	}
	writeFile("kurento/complexTypes.go", ret)

}

func writeFile(path string, classes []string) {
	content := strings.Join(classes, "\n")
	tpl, _ := template.New("package").Parse(packageTemplate)
	buff := bytes.NewBufferString("")
	tpl.Execute(buff, map[string]string{
		"Content": content,
	})
	ioutil.WriteFile(path, buff.Bytes(), os.ModePerm)
}

func main() {
	// Perpare complexTypes to get the list
	parseComplexTypes()

	// create base
	c := getModel(CORE).RemoteClasses
	coreclasses := parse(c)
	writeFile("kurento/core.go", coreclasses)

	// make same for each interfaces
	getInterfaces()

	// finish by putting base.go
	data, _ := ioutil.ReadFile("kurento_go_base/base.go")
	// Write data to dst
	ioutil.WriteFile("kurento/base.go", data, os.ModePerm)

}
