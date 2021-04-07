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
	"errors"
	"fmt"
	html "html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	text "text/template"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	// Name is the application name shown for "dumblog version"
	Name string = "dumblog"
	// Version is the current version shown for "dumblog version"
	Version string = "0.1"

	metaSource string = ".meta.yaml"
	postSource string = "post.md"
	postTmpl   string = "_post.html"
	postDest   string = "index.html"
)

type filePath struct {
	source string // source path
	rel    string // relative destination path
}

// Generator is loads & parses templates and then execs & writes them to a directory.
type Generator struct {
	meta     Meta
	postTmpl *html.Template
	tmplHTML []*html.Template
	tmplText []*text.Template
	posts    []Post
	files    []filePath
}

// New returns a new *Generator instance.
func New() *Generator {
	return &Generator{}
}

func loadMeta(path string) (Meta, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = nil // Ignore it
		}
		return nil, err
	}

	var meta Meta
	if err := yaml.UnmarshalStrict(b, &meta); err != nil {
		return nil, err
	}

	// yaml decodes the keys to lower case, make them upper case instead so it's nicer for the templates
	var keys []string
	for k := range meta {
		keys = append(keys, k)
	}
	for _, k := range keys {
		meta[strings.Title(k)] = meta[k]
		delete(meta, k) // the lower case version
	}
	return meta, nil
}

// ReadTemplate loads and parses the template files from `dir`.
// Optionally tries to load a `.meta.yaml` file, used for providing global meta data to the templates.
func (g *Generator) ReadTemplate(dir string) error {
	var err error
	g.meta, err = loadMeta(filepath.Join(dir, metaSource))
	if err != nil {
		return err
	}

	return filepath.WalkDir(dir, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if de.IsDir() {
			return nil
		} else if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		rel := trimDir(path, dir)
		ext := filepath.Ext(path)

		switch {
		case rel == postTmpl: // post template, must be caught before the other html loader
			g.postTmpl, err = loadHTML(path, rel)
			if err != nil {
				return err
			}

		case ext == ".html": // HTML templates
			t, err := loadHTML(path, rel)
			if err != nil {
				return err
			}
			g.tmplHTML = append(g.tmplHTML, t)

		case ext == ".xml", ext == ".txt": // Special text templates
			t, err := loadText(path, rel)
			if err != nil {
				return err
			}
			g.tmplText = append(g.tmplText, t)

		case ext == ".md": // Posts
			post, err := readPost(path, filepath.Join(filepath.Dir(rel), postDest))
			if err != nil {
				return fmt.Errorf("read %q: %s", path, err)
			}
			g.posts = append(g.posts, post)

		default: // Other static files
			g.files = append(g.files, filePath{
				source: path,
				rel:    rel,
			})
		}
		return nil
	})
}

func (g *Generator) loadParams() TemplateParams {
	params := TemplateParams{
		Time:  time.Now(),
		Meta:  g.meta,
		Posts: g.posts,
		Tags:  readTags(g.posts),
	}

	for _, t := range g.tmplHTML {
		url := path.Join("/", filepath.ToSlash(t.Name()))
		params.Pages = append(params.Pages, url)
	}
	for _, p := range g.posts {
		url := path.Join("/", filepath.ToSlash(p.rel))
		params.Pages = append(params.Pages, url)
	}

	sortPosts(params.Posts)
	sortTags(params.Tags)
	sort.Strings(params.Pages)
	return params
}

// ExecuteTemplate executes the templates and write the resulting files to dir. It also copy over any other plain files.
// ReadTemplate must have been called before.
func (g *Generator) ExecuteTemplate(dir string) error {
	params := g.loadParams()

	for _, p := range g.posts {
		path := filepath.Join(dir, p.rel)
		params.Current = p
		if err := executeTemplate(path, g.postTmpl, params); err != nil {
			return err
		}
	}
	if len(g.posts) > 0 {
		params.Current = g.posts[0]
	}

	for _, t := range g.tmplHTML {
		path := filepath.Join(dir, t.Name())
		if err := executeTemplate(path, t, params); err != nil {
			return err
		}
	}

	for _, t := range g.tmplText {
		path := filepath.Join(dir, t.Name())
		if err := executeTemplate(path, t, params); err != nil {
			return err
		}
	}

	for _, f := range g.files {
		path := filepath.Join(dir, f.rel)
		if err := copyFile(f.source, path); err != nil {
			return err
		}
	}
	return nil
}
