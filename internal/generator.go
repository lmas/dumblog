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
	Version string = "0.1.5"

	postOrig     string = "post.md"
	postDest     string = "index.html"
	metaSource   string = ".dumblog/meta.yaml"
	layoutSource string = ".dumblog/layout.html"
	postSource   string = ".dumblog/post.html"
)

type filePath struct {
	source string // source path
	rel    string // relative destination path
}

// Generator is loads & parses templates and then execs & writes them to a directory.
type Generator struct {
	meta       Meta
	tmplLayout *text.Template
	tmplPost   *text.Template
	posts      []Post
	tmpls      []filePath
	files      []filePath
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
	g.tmplLayout, err = loadTemplate(filepath.Join(dir, layoutSource))
	if err != nil {
		return err
	}
	g.tmplPost, err = cloneTemplate(g.tmplLayout, filepath.Join(dir, postSource))
	if err != nil {
		return err
	}

	return filepath.WalkDir(dir, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if de.IsDir() {
			return nil
		}

		rel := trimDir(path, dir)
		if containsDot(rel) {
			return nil
		}
		ext := filepath.Ext(rel)

		switch {
		case filepath.Base(rel) == postOrig: // Posts
			post, err := readPost(path, filepath.Join(filepath.Dir(rel), postDest))
			if err != nil {
				return fmt.Errorf("read %q: %s", path, err)
			}
			g.posts = append(g.posts, post)

		case ext == ".html", ext == ".xml", ext == ".txt": // Text templates
			g.tmpls = append(g.tmpls, filePath{
				source: path,
				rel:    rel,
			})

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

	for _, f := range g.tmpls {
		if filepath.Ext(f.rel) != ".html" {
			continue
		}
		url := path.Join("/", filepath.ToSlash(f.rel))
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
		params.Current = p
		path := filepath.Join(dir, p.rel)
		if err := executeTemplate(path, g.tmplPost, params); err != nil {
			return err
		}

	}
	if len(g.posts) > 0 {
		params.Current = g.posts[0]
	}

	for _, f := range g.tmpls {
		tmpl, err := cloneTemplate(g.tmplLayout, f.source)
		if err != nil {
			return err
		}
		path := filepath.Join(dir, f.rel)
		if err := executeTemplate(path, tmpl, params); err != nil {
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
