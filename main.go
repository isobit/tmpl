package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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
	DataFilenames []string `cli:"name=data-file,short=d,append,placeholder=FILENAME,nodefault,help=file to load data from (can be specified multiple times)"`
	Data          []string `cli:"name=data,short=D,append,placeholder=KEY=VAL,nodefault,help=set top-level data keys (can be specified multiple times)"`
	NoEnv         bool     `cli:"hidden,help=disable including environment variables as data at .Env"`
	InPlace       bool     `cli:"short=i,help=write to output files instead of writing to stdout"`
	Files         []string `cli:"args"`
}

func (cmd *Cmd) Run() error {
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
	t.Funcs(template.FuncMap{
		"must": func(v any) (any, error) {
			if v == nil {
				return nil, fmt.Errorf("missing")
			}
			if s, ok := v.(string); ok {
				if s == "" {
					return nil, fmt.Errorf("missing")
				}
			}
			return v, nil
		},
		"parseUrl": parseUrlInfo,
	})
	return t
}

type urlInfo struct {
	Scheme   string
	Username string
	Password string
	Hostname string
	Port     string
	Path     string
	Query    map[string][]string
	Fragment string
}

func parseUrlInfo(s string) (urlInfo, error) {
	u, err := url.Parse(s)
	if err != nil {
		return urlInfo{}, err
	}

	password, _ := u.User.Password()

	return urlInfo{
		Scheme:   u.Scheme,
		Username: u.User.Username(),
		Password: password,
		Hostname: u.Hostname(),
		Port:     u.Port(),
		Path:     u.Path,
		Query:    u.Query(),
		Fragment: u.Fragment,
	}, nil
}
