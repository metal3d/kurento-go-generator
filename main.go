package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

const strTemplate = `
{{ $name := .Name}}
{{ .Doc }}
type {{ .Name }} struct {
	{{ .Extends }}
	{{ range .Properties }}
	{{ .doc }}
	{{ .name }} {{ .type }}
	{{ end }}
}


// Return contructor params to be called by "Create"
func (elem *{{.Name}}) getConstructorParams(from IMediaObject, options map[string]interface{}) map[string]interface{} {
	{{ if len .Constructor.Params }}
	// Create basic constructor params
	ret := map[string]interface{} {
		{{ range .Constructor.Params }}{{ if eq .type "string" "boolean" "int" }}"{{ .name }}" : {{ .defaultValue }},
			{{ else }}"{{ .name }}" : fmt.Sprintf("%s", from), //elem.getField("{{ .name }}"),
			{{ end }}{{ end }}
	}

	// then merge options
	mergeOptions(ret, options)

	return ret
	{{ else }}return options
	{{ end }}
}

{{ range .Methods }}
{{ .Doc }}
func (elem *{{$name}}) {{ .Name | title }}({{ range $i, $e := .Params }}{{ if $i }}, {{ end }} {{ $e.name }} {{ $e.type}}{{ end }}) {
	req := elem.getInvokeRequest()
	req["params"] = map[string]interface{}{
		"operation" : "{{ .Name }}",
		"object"	: elem.id,{{ if .Params }}
		"operationParams" : map[string]string{
			{{ range .Params }} "{{.name }}" : fmt.Sprintf("%s", {{.name}} ),
			{{ end }}
		},
		{{ end }}
	}

}
{{ end }}
`

const complexTypeTemplate = `
{{ if eq .TypeFormat "ENUM" }}
{{ $name := .Name }}
{{ .Doc }}
type {{.Name}} string
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

const packageTempalte = `package kurento
{{ .Content }}
`

var re = regexp.MustCompile(`(.+)\[\]`)

type method struct {
	Name   string
	Doc    string
	Params []map[string]interface{}
}

type class struct {
	Name        string
	Extends     string
	Doc         string
	Abstract    bool
	Properties  []map[string]interface{}
	Events      []string
	Constructor method
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

var funcMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"title":     strings.Title,
	"uppercase": strings.ToUpper,
}

func formatTypes(p map[string]interface{}) map[string]interface{} {
	doc := strings.Split(p["doc"].(string), "\n")
	for i, d := range doc {
		doc[i] = "//" + d
	}
	p["doc"] = strings.Join(doc, "\n")

	if p["type"] == "String[]" {
		p["type"] = "[]string"
	}

	if p["type"] == "String" {
		p["type"] = "string"
	}

	if p["type"] == "float" {
		p["type"] = "float64"
	}

	if re.MatchString(p["type"].(string)) {
		found := re.FindAllStringSubmatch(p["type"].(string), -1)
		p["type"] = "[]" + found[0][1]
	}

	if p["defaultValue"] == "" || p["defaultValue"] == nil {
		switch p["type"] {
		case "string":
			p["defaultValue"] = `""`
		case "boolean":
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
						p["type"] = "[]*" + p["type"].(string)[2:]
					} else {
						p["type"] = "*" + p["type"].(string)
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
			doc := strings.Split(m.Doc, "\n")
			for i, d := range doc {
				doc[i] = "//" + d
			}
			m.Doc = strings.Join(doc, "\n")

			cl.Methods[j] = m

		}
		for j, p := range cl.Constructor.Params {
			p := formatTypes(p)
			//log.Println(p)
			cl.Constructor.Params[j] = p
		}

		tpl, _ := template.New("structure").Funcs(funcMap).Parse(strTemplate)
		buff := bytes.NewBufferString("")
		doc := strings.Split(cl.Doc, "\n")
		for i, d := range doc {
			doc[i] = "//" + d
		}
		cl.Doc = strings.Join(doc, "\n")

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

			doc := strings.Split(ctype.Doc, "\n")
			for i, d := range doc {
				doc[i] = "//" + d
			}
			ctype.Doc = strings.Join(doc, "\n")

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
	tpl, _ := template.New("package").Parse(packageTempalte)
	buff := bytes.NewBufferString("")
	tpl.Execute(buff, map[string]string{
		"Content": content,
	})
	ioutil.WriteFile(path, buff.Bytes(), os.ModePerm)
}

func main() {
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
	/*/
	//*/
	parseComplexTypes()

}
