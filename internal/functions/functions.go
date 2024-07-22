package functions

import (
	"bytes"
	"fmt"
	// "os"
	// "path/filepath"
	"text/template"
)

func must(v any) (any, error) {
	if v == nil {
		return nil, fmt.Errorf("missing")
	}
	if s, ok := v.(string); ok {
		if s == "" {
			return nil, fmt.Errorf("missing")
		}
	}
	return v, nil
}

func FuncMap() template.FuncMap {
	markdownFuncs := NewMarkdownFuncs()
	return template.FuncMap{
		"must":           must,
		"parseUrl":       parseUrlInfo,
		"markdownToHTML": markdownFuncs.MarkdownToHTML,
	}
}

func TemplateFuncMap(t *template.Template) template.FuncMap {
	return template.FuncMap{
		"eval": func(name string, arg interface{}) (string, error) {
			var buf bytes.Buffer
			err := t.ExecuteTemplate(&buf, name, arg)
			return buf.String(), err
		},
	}
}

// func RelativeTemplateFuncMap(t *template.Template, dir string) template.FuncMap {
// 	return template.FuncMap{
// 		"readFile": func(filename string) (string, error) {
// 			data, err := os.ReadFile(filepath.Join(dir, filename))
// 			return string(data), err
// 		},
// 		"abspath": func(filename string) (string, error) {
// 			return filepath.Abs(filepath.Join(dir, filename))
// 		},
// 		"partial": func(filename string, data any) (string, error) {
// 			nt, err := t.Clone()
// 			if err != nil {
// 				return "", err
// 			}
// 			nt.Funcs(RelativeTemplateFuncMap(t, filepath.Dir(filename)))

// 			srcData, err := os.ReadFile(filepath.Join(dir, filename))
// 			if err != nil {
// 				return "", err
// 			}
// 			if _, err := nt.Parse(string(srcData)); err != nil {
// 				return "", err
// 			}

// 			var b bytes.Buffer
// 			if err := nt.Execute(&b, data); err != nil {
// 				return "", err
// 			}
// 			return b.String(), nil
// 		},
// 	}
// }
