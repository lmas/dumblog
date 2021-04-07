# dumblog
[![PkgGoDev](https://pkg.go.dev/badge/github.com/lmas/dumblog)](https://pkg.go.dev/github.com/lmas/dumblog)

Yet another minimal, dumb, static blog generator.

## Features

- Standard Go templates
- Easy to write blog posts:
  * Frontmatter meta data using yaml (`title`, `published` time, `short` description and list of `tags`)
  * Main body written using markdown ([commonmark](https://commonmark.org/help/) flavour)
- Optional site-wide meta data loaded from a `.meta.yaml` file

## Status

Beta testing.

## Installation

        go install github.com/lmas/dumblog

## Usage

Create an example template to start working from (default output dir is `./example`):

        dumblog init

Make any edits to the example templates or add new posts.

Then you can generate the final, static site (default output dir is `./public`):

        dumblog update ./example

**Optionally** run a local demo server to inspect the generated site:

        dumblog web

Finally, you can upload the static dir to your web host.

## Options

```
dumblog v0.1

Flags:
  -addr string
    	Local IP address for hosting the demo web server (default "127.0.0.1:8080")
  -out string
    	Output dir for generated site (default "./public")

Commands:
  init
  	Writes an example template (default output dir is `./example`)
  update
  	Regenerate the static site
  web
  	Run a demo web server
  version
  	Print version and exit
  help
  	Print this help message and exit
```

## TODO

- Unit tests
- Fuzzy testing inputs
- Check files' modtimes and avoid regenerating unchanged files

## License

GPL, See the [LICENSE](LICENSE) file for details.
