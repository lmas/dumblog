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
	"bytes"
	"embed"
	"fmt"
	html "html/template"
	"io/fs"
	"net/url"
	"os"
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
	"postslimit": func(max int, posts []Post) []Post {
		l := len(posts)
		if l > max {
			l = max
		}
		return posts[:l]
	},
	"postsbydir": func(dir string, posts []Post) []Post {
		var list []Post
		for _, p := range posts {
			if firstDir(p.rel) == dir {
				list = append(list, p)
			}
		}
		return list
	},
	"slugify": func(s string) string {
		return url.PathEscape(strings.ReplaceAll(strings.ToLower(s), " ", "_"))
	},
}

// CreateTemplate creates an example dir with some template files you can use.
func CreateTemplate(dst, src string, content embed.FS) error {
	return fs.WalkDir(content, src, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if de.IsDir() {
			return nil // Skip
		}
		b, err := content.ReadFile(path)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, trimDir(path, src))
		return writeFile(out, bytes.TrimSpace(b))
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func loadTemplate(path string) (*text.Template, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	return text.New(name).Funcs(TemplateFuncs).Parse(string(b))
}

func cloneTemplate(base *text.Template, path string) (*text.Template, error) {
	t, err := base.Clone()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	return t.New(name).Parse(string(b))
}

func executeTemplate(file string, tmpl *text.Template, params TemplateParams) error {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, tmpl.Name(), params); err != nil {
		return err
	}
	return writeFile(file, bytes.TrimSpace(buf.Bytes()))
}
