package main

import (
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
	`)
	cmd.Parse().RunFatal()
}

type Cmd struct {
	Debug           bool     `cli:"hidden"`
	ErrMissingKey   bool     `cli:"short=e,help=error for missing keys"`
	Data            []string `cli:"name=data,short=d,append,placeholder=KEY=VAL,nodefault,help=set top-level data keys (can be specified multiple times)"`
	DataFilenames   []string `cli:"name=datafile,short=D,append,placeholder=FILENAME,nodefault,help=file to load data from (can be specified multiple times)"`
	ContentFilename string   `cli:"name=contentfile,short=C"`
	NoEnv           bool     `cli:"hidden,help=disable including environment variables as data at .Env"`
	TemplateName    string   `cli:"short=t"`
	Files           []string `cli:"args"`
}

func (cmd *Cmd) debugf(format string, a ...any) {
	if !cmd.Debug {
		return
	}
	if len(format) > 0 && format[len(format)-1] != '\n' {
		format = format + "\n"
	}
	fmt.Fprintf(os.Stderr, format, a...)
}

func (cmd *Cmd) data() (map[string]any, error) {
	data := map[string]any{}

	if !cmd.NoEnv {
		cmd.debugf("setting Env from env vars")
		env := map[string]string{}
		for _, keyval := range os.Environ() {
			key, val, _ := strings.Cut(keyval, "=")
			env[key] = val
		}
		data["Env"] = env
	}

	for _, filename := range cmd.DataFilenames {
		cmd.debugf("reading data file: %s", filename)
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
		cmd.debugf("adding data key from arguments: %s", key)
		data[key] = value
	}

	if cmd.ContentFilename != "" {
		contentData, err := os.ReadFile(cmd.ContentFilename)
		if err != nil {
			return data, fmt.Errorf("failed to read content from %s: %w", cmd.ContentFilename, err)
		}
		data["Content"] = string(contentData)
	}

	return data, nil
}

func (cmd *Cmd) Run() error {
	if len(cmd.Files) < 1 {
		return cli.UsageErrorf("at least one template file is required")
	}

	var t *template.Template

	for _, filename := range cmd.Files {
		var name, text string
		var err error
		if filename == "-" {
			name, text, err = cmd.readFileStdin()
		} else {
			name, text, err = cmd.readFileOS(filename)
		}
		if err != nil {
			return err
		}

		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
			if cmd.ErrMissingKey {
				t.Option("missingkey=error")
			}
			t.Funcs(sprig.TxtFuncMap())
			t.Funcs(internalFunctions.FuncMap())
			t.Funcs(internalFunctions.TemplateFuncMap(t))
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		cmd.debugf("new template: filename=%s name=%s\n", filename, name)
		if _, err := tmpl.Parse(text); err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", filename, err)
		}
	}

	data, err := cmd.data()
	if err != nil {
		return err
	}

	cmd.debugf(t.DefinedTemplates())
	cmd.debugf("executing template: %s", t.Name())
	if err := t.Execute(os.Stdout, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func (cmd *Cmd) readFileStdin() (string, string, error) {
	cmd.debugf("reading template from stdin")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("failed to read template from stdin: %w", err)
	}
	return "-", string(data), nil
}

func (cmd *Cmd) readFileOS(filename string) (string, string, error) {
	cmd.debugf("reading template from file: %s", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to read template file %s: %w", filename, err)
	}
	return filepath.Base(filename), string(data), nil
}
