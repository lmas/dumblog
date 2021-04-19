// Copyright © 2021 Alex
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package internal

import (
	"embed"
	"fmt"
	html "html/template"
	"io"
	"io/fs"
	"net/url"
	"path/filepath"
	"strings"
	text "text/template"
	"time"
)

// TemplateParams is a struct holding all available meta data you can use in a template.
type TemplateParams struct {
	// Time is the current date and time
	Time time.Time
	// Meta contains user defined meta data, see Meta
	Meta Meta
	// Posts is a list of parsed Post
	Posts []Post
	// Tags is a list of parsed Tag
	Tags []Tag
	// Pages is a list of all html pages that will be written
	Pages []string
	// Current is the latest post published (or the active post while writing each individual post)
	Current Post
}

// TemplateFuncs contains helper functions for the templates
var TemplateFuncs = text.FuncMap{
	"atomdate": func(t time.Time) string {
		return t.UTC().Format(time.RFC3339)
	},
	"shortdate": func(t time.Time) string {
		return t.UTC().Format("2006-01-02")
	},
	"prettydate": func(t time.Time) string {
		return t.UTC().Format("Monday, 02 January 2006")
	},
	"prettyduration": func(d time.Duration) string {
		s, m, h := int(d.Seconds()), int(d.Minutes()), int(d.Hours())
		if s < 1 {
			return "instant"
		} else if s == 1 {
			return "1 second"
		} else if s < 60 {
			return fmt.Sprintf("%d seconds", s)
		} else if m == 1 {
			return "1 minute"
		} else if m < 60 {
			return fmt.Sprintf("%d minutes", m)
		} else if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	},
	"safehtml": func(s string) html.HTML {
		return html.HTML(s) // #nosec G203
	},
	"maxposts": func(max int, posts []Post) []Post {
		l := len(posts)
		if l > max {
			l = max
		}
		return posts[:l]
	},
	"slugify": func(s string) string {
		return url.PathEscape(strings.ReplaceAll(strings.ToLower(s), " ", "_"))
	},
}

var (
	//go:embed example/*
	content    embed.FS
	exampleDir string = "example"
)

// CreateTemplate creates an example dir with some template files you can use.
func CreateTemplate(dir string) error {
	return fs.WalkDir(content, exampleDir, func(src string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if de.IsDir() {
			return nil // Skip
		}
		dst := filepath.Join(dir, trimDir(src, exampleDir))
		if err := writeTemplate(src, dst); err != nil {
			return err
		}
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func loadHTML(path string) (*html.Template, error) {
	name := filepath.Base(path)
	return html.New(name).Funcs(html.FuncMap(TemplateFuncs)).ParseFiles(path)
}

func cloneHTML(orig *html.Template, path string) (*html.Template, error) {
	t, err := orig.Clone()
	if err != nil {
		return nil, err
	}
	return t.ParseFiles(path)
}

func loadText(path string) (*text.Template, error) {
	name := filepath.Base(path)
	return text.New(name).Funcs(TemplateFuncs).ParseFiles(path)
}

// Let's us accept both `text/template` and `html/template` in executeTemplate()
type templateExecuter interface {
	Execute(io.Writer, interface{}) error
}

func executeTemplate(file string, tmpl templateExecuter, params TemplateParams) error {
	f, err := createFile(file)
	if err != nil {
		return err
	}
	defer f.Close() // #nosec G307
	if err := tmpl.Execute(f, params); err != nil {
		return err
	}
	return f.Sync()
}

func writeTemplate(src, dst string) error {
	r, err := content.Open(src)
	if err != nil {
		return err
	}
	// Assume it's safe to ignore Close() errors on files just being read
	defer r.Close() // #nosec G307
	w, err := createFile(dst)
	if err != nil {
		return err
	}
	defer w.Close() // #nosec G307
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	// Hopefully catches any file write errors before dst.Close(), see:
	// https://www.joeshaw.org/dont-defer-close-on-writable-files/
	return w.Sync()
}
