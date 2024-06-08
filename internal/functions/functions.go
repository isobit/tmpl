package functions

import (
	"text/template"
)

var FuncMap template.FuncMap = template.FuncMap{
	"must":     must,
	"parseUrl": parseUrlInfo,
}
