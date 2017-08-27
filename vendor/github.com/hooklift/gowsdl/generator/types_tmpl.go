// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package generator

var typesTmpl = `
{{define "SimpleType"}}
	{{$type := replaceReservedWords .Name}}
	type {{$type}} {{toGoType .Restriction.Base}}
	const (
		{{with .Restriction}}
			{{range .Enumeration}}
				{{if .Doc}} {{.Doc | comment}} {{end}}
				{{$type}}_{{$value := replaceReservedWords .Value}}{{$value | makePublic}} {{$type}} = "{{$value}}" {{end}}
		{{end}}
	)
{{end}}

{{define "ComplexContent"}}
	{{$baseType := toGoType .Extension.Base}}
	{{ if $baseType }}
		{{$baseType}}
	{{end}}

	{{template "Elements" .Extension.Sequence}}
	{{template "Attributes" .Extension.Attributes}}
{{end}}

{{define "Attributes"}}
	{{range .}}
		{{if .Doc}} {{.Doc | comment}} {{end}} {{if not .Type}}
			{{ .Name | makePublic}} {{toGoType .SimpleType.Restriction.Base}} ` + "`" + `xml:"{{.Name}},attr,omitempty"` + "`" + `
		{{else}}
			{{ .Name | makePublic}} {{toGoType .Type}} ` + "`" + `xml:"{{.Name}},attr,omitempty"` + "`" + `
		{{end}}
	{{end}}
{{end}}

{{define "SimpleContent"}}
	Value {{toGoType .Extension.Base}}{{template "Attributes" .Extension.Attributes}}
{{end}}

{{define "ComplexTypeGlobal"}}
	{{$name := replaceReservedWords .Name}}
	type {{$name}} struct {
		XMLName xml.Name ` + "`xml:\"{{targetNamespace}} {{$name}}\"`" + `
		{{if ne .ComplexContent.Extension.Base ""}}
			{{template "ComplexContent" .ComplexContent}}
		{{else if ne .SimpleContent.Extension.Base ""}}
			{{template "SimpleContent" .SimpleContent}}
		{{else}}
			{{template "Elements" .Sequence.Elements}}
			{{template "Choices" .Sequence.Choices}}
			{{template "Groups" .Sequence.Groups}}
			{{template "Elements" .Sequence.Sequences}}
			{{template "Elements" .Choice.Elements}}
			{{template "Elements" .All}}
			{{template "Attributes" .Attributes}}
		{{end}}
	}
{{end}}

{{define "ComplexTypeLocal"}}
	{{$name := replaceReservedWords .Name}}
	{{with .ComplexType}}
		type {{$name}} struct {
			XMLName xml.Name ` + "`xml:\"{{targetNamespace}} {{$name}}\"`" + `
			{{if ne .ComplexContent.Extension.Base ""}}
				{{template "ComplexContent" .ComplexContent}}
			{{else if ne .SimpleContent.Extension.Base ""}}
				{{template "SimpleContent" .SimpleContent}}
			{{else}}
				{{template "Elements" .Sequence}}
				{{template "Elements" .Choice}}
				{{template "Elements" .All}}
				{{template "Attributes" .Attributes}}
			{{end}}
		}
	{{end}}
{{end}}

{{define "ComplexTypeInline"}}
	{{with .ComplexType}}
		{{if ne .ComplexContent.Extension.Base ""}}
			{{template "ComplexContent" .ComplexContent}}
		{{else if ne .SimpleContent.Extension.Base ""}}
			{{template "SimpleContent" .SimpleContent}}
		{{else}}
			{{template "Elements" .Sequence}}
			{{template "Elements" .Choice}}
			{{template "Elements" .All}}
			{{template "Attributes" .Attributes}}
		{{end}}
	{{end}}
{{end}}

{{define "Elements"}}
	{{range .}}
		{{if not .Type}}
			{{template "ComplexTypeInline" .}}
		{{else}}
			{{if .Doc}} {{.Doc | comment}} {{end}}
			{{replaceReservedWords .Name | makePublic}} {{if eq .MaxOccurs "unbounded"}}[]{{end}}{{.Type | toGoType}} ` + "`" + `xml:"{{.Name}},omitempty"` + "`" + `
		{{end}}
	{{end}}
{{end}}

{{range .Schemas}}
	{{range .SimpleType}}
		{{template "SimpleType" .}}
	{{end}}
	{{range .Elements}}
		{{if not .Type}}
			{{template "ComplexTypeLocal" .}}
		{{end}}
	{{end}}
	{{range .ComplexTypes}}
		{{template "ComplexTypeGlobal" .}}
	{{end}}
{{end}}
`
