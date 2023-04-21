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
)

func main() {
	cmd := cli.New("tmpl", &Cmd{})
	cmd.SetDescription(`
		Render STDIN and/or template file args using Go's text/template engine.

		Data can be specified as explicit top-level keys as flags, or as a
		JSON/TOML/YAML data file.

		When -i/--in-place is specified, output is written to the same file
		name as each template file, but with the ".tmpl" extension trimmed.
	`)
	cmd.Parse().RunFatal()
}

type Cmd struct {
	ErrMissingKey bool     `cli:"short=e,help=error for missing keys"`
	DataFilenames []string `cli:"name=data-file,short=d,append,placeholder=DATAFILE,nodefault,help=file to load data from (can be specified multiple times)"`
	Data          []string `cli:"name=data,short=D,append,placeholder=KEY=VAL,nodefault,help=set top-level data keys (can be specified multiple times)"`
	InPlace       bool     `cli:"short=i,help=write to output files instead of writing to stdout"`
	Files         []string `cli:"args"`
}

func (cmd *Cmd) Run() error {
	data := map[string]any{}
	for _, filename := range cmd.DataFilenames {
		dataData, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		switch filepath.Ext(filename) {
		case ".json":
			if err := json.Unmarshal(dataData, &data); err != nil {
				return err
			}
		case ".yaml":
			if err := yaml.Unmarshal(dataData, &data); err != nil {
				return err
			}
		case ".toml":
			if err := toml.Unmarshal(dataData, &data); err != nil {
				return err
			}
		default:
			return fmt.Errorf("data file has unsupported format: %s", filename)
		}
	}
	for _, s := range cmd.Data {
		key, value, _ := strings.Cut(s, "=")
		data[key] = value
	}

	stdinStat, _ := os.Stdin.Stat()
	if stdinStat.Mode()&os.ModeCharDevice == 0 {
		text, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading template from stdin: %w", err)
		}

		tmpl, err := cmd.newTemplate().Parse(string(text))
		if err != nil {
			return fmt.Errorf("error parsing template from stdin: %w", err)
		}

		if err := tmpl.Execute(os.Stdout, data); err != nil {
			return fmt.Errorf("error executing template from stdin: %w", err)
		}
	}

	for _, filename := range cmd.Files {
		text, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("error reading template file %s: %w", filename, err)
		}

		tmpl, err := cmd.newTemplate().Parse(string(text))
		if err != nil {
			return fmt.Errorf("error parsing template file %s: %w", filename, err)
		}

		var out io.Writer
		if cmd.InPlace {
			outFilename := strings.TrimSuffix(filename, ".tmpl")
			fmt.Fprintf(os.Stderr, "writing %s\n", outFilename)
			inInfo, err := os.Stat(filename)
			if err != nil {
				return fmt.Errorf("error getting mode of template file %s: %w", filename, err)
			}
			outFile, err := os.OpenFile(outFilename, os.O_RDWR|os.O_CREATE, inInfo.Mode())
			defer outFile.Close()
			out = outFile
		} else {
			out = os.Stdout
		}

		if err := tmpl.Execute(out, data); err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
	}

	return nil
}

func (cmd *Cmd) newTemplate() *template.Template {
	t := template.New("")
	if cmd.ErrMissingKey {
		t.Option("missingkey=error")
	}
	t.Funcs(sprig.TxtFuncMap())
	return t
}
