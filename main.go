package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/isobit/cli"
	"gopkg.in/yaml.v3"

	internalFunctions "github.com/isobit/tmpl/internal/functions"
)

func main() {
	cmd := cli.New("tmpl", &Cmd{})
	cmd.SetDescription(`
		Render STDIN and/or template file args using Go's text/template engine.
		See https://pkg.go.dev/text/template for more documentation on the
		templating language. Note that Sprig functions are available by
		default, see http://masterminds.github.io/sprig/ for more
		documentation.

		Data can be specified as explicit top-level keys as flags, or as a
		JSON/TOML/YAML data file. All environment variables are available by
		default on the ".Env" data key.

		When -i/--in-place is specified, output is written to the same file
		name as each template file, but with the ".tmpl" extension trimmed.
	`)
	cmd.Parse().RunFatal()
}

type Cmd struct {
	ErrMissingKey bool     `cli:"short=e,help=error for missing keys"`
	Data          []string `cli:"name=data,short=d,append,placeholder=KEY=VAL,nodefault,help=set top-level data keys (can be specified multiple times)"`
	DataFilenames []string `cli:"name=datafile,short=D,append,placeholder=FILENAME,nodefault,help=file to load data from (can be specified multiple times)"`
	NoEnv         bool     `cli:"hidden,help=disable including environment variables as data at .Env"`
	TemplateName  string   `cli:"short=t"`
	Files         []string `cli:"args"`
}

func (cmd *Cmd) data() (map[string]any, error) {
	data := map[string]any{}

	if !cmd.NoEnv {
		env := map[string]string{}
		for _, keyval := range os.Environ() {
			key, val, _ := strings.Cut(keyval, "=")
			env[key] = val
		}
		data["Env"] = env
	}

	for _, filename := range cmd.DataFilenames {
		dataData, err := os.ReadFile(filename)
		if err != nil {
			return data, err
		}

		switch filepath.Ext(filename) {
		case ".json":
			if err := json.Unmarshal(dataData, &data); err != nil {
				return data, err
			}
		case ".yaml":
			if err := yaml.Unmarshal(dataData, &data); err != nil {
				return data, err
			}
		case ".toml":
			if err := toml.Unmarshal(dataData, &data); err != nil {
				return data, err
			}
		default:
			return data, fmt.Errorf("data file has unsupported format: %s", filename)
		}
	}
	for _, s := range cmd.Data {
		key, value, _ := strings.Cut(s, "=")
		data[key] = value
	}
	return data, nil
}

func templateName(filename string) string {
	return filepath.Base(filename)
}

func (cmd *Cmd) Run() error {
	if len(cmd.Files) < 1 {
		return cli.UsageErrorf("at least one template file is required")
	}

	var outputTemplateName string
	if cmd.TemplateName != "" {
		outputTemplateName = cmd.TemplateName
	} else {
		outputTemplateName = templateName(cmd.Files[len(cmd.Files)-1])
	}

	t := template.New(outputTemplateName)
	if cmd.ErrMissingKey {
		t.Option("missingkey=error")
	}
	t.Funcs(sprig.TxtFuncMap())
	t.Funcs(internalFunctions.FuncMap)
	markdownFuncs := internalFunctions.NewMarkdownFuncs()
	t.Funcs(template.FuncMap{
		"markdownToHTML": markdownFuncs.MarkdownToHTML,
		"eval": func(name string, arg interface{}) (string, error) {
			var buf bytes.Buffer
			err := t.ExecuteTemplate(&buf, name, arg)
			return buf.String(), err
		},
		"readFile": func(filename string) (string, error) {
			data, err := os.ReadFile(filename)
			return string(data), err
		},
	})

	for _, filename := range cmd.Files {
		var text string
		if filename == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read template from stdin: %w", err)
			}
			text = string(data)
		} else {
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read template file %s: %w", filename, err)
			}
			text = string(data)
		}

		// fmt.Printf("filename=%s name=%s\n", filename, templateName(filename))
		t = t.New(templateName(filename))
		// fmt.Printf("defined %s\n", t.Name())
		if _, err := t.Parse(text); err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", filename, err)
		}
	}

	data, err := cmd.data()
	if err != nil {
		return err
	}

	if err := t.Execute(os.Stdout, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
