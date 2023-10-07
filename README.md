# tmpl

`tmpl` is a very simple tool which renders text file templates using Go's
`text/template` engine.

```
USAGE:
    tmpl [OPTIONS] [ARGS]

OPTIONS:
    -h, --help                  show usage help
    -e, --err-missing-key       error for missing keys
    -d, --data-file <FILENAME>  file to load data from (can be specified multiple times)
    -D, --data <KEY=VAL>        set top-level data keys (can be specified multiple times)
    -i, --in-place              write to output files instead of writing to stdout

DESCRIPTION:
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
```
