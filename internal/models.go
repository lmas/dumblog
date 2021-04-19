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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Meta is a map of strings that allows you to insert custom data into your templates, like links and titles.
// See example/.meta.yaml and the example html templates for usage.
type Meta map[string]string

////////////////////////////////////////////////////////////////////////////////////////////////////

// Post contains the meta data header from a `post.md`.
type Post struct {
	// filepaths
	source string
	rel    string

	Meta struct {
		// Title is the title of the post
		Title string
		// Published is the publishing date
		Published time.Time
		// Short is a short description for the post
		Short string
		// Tags is a list of optional string tags
		Tags []string
	}
}

// Body parses the post's body and parses it as commonmark.
func (p Post) Body() (string, error) {
	// f, err := os.Open(p.File()) // #nosec G304
	f, err := os.Open(p.source) // #nosec G304
	if err != nil {
		return "", err
	}
	// It's only being read, should be safe to ignore Close() errors
	defer f.Close() // #nosec G307

	var body bytes.Buffer
	if err := scanBody(f, &body); err != nil {
		return "", fmt.Errorf("read body: %s", err)
	}
	return body.String(), nil
}

// Link returns a relative http link to the post.
func (p Post) Link() string {
	return path.Join("/", filepath.ToSlash(p.rel))
}

func sortPosts(posts []Post) {
	const date string = "20060102"
	sort.Slice(posts, func(i, j int) bool {
		// SORT ORDER: latest date then by name
		p1, p2 := posts[i], posts[j]
		d1, d2 := p1.Meta.Published.Format(date), p2.Meta.Published.Format(date)
		if d1 != d2 {
			return d1 > d2
		}
		return p1.Meta.Title < p2.Meta.Title
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Tag contains a list of posts it was tagged in.
type Tag struct {
	Title string
	Posts []Post
}

func sortTags(tags []Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Title < tags[j].Title
	})
}

func readTags(posts []Post) []Tag {
	tagMap := make(map[string][]Post)
	for _, p := range posts {
		for _, t := range p.Meta.Tags {
			tagMap[t] = append(tagMap[t], p)
		}
	}

	var tags []Tag
	for t, ps := range tagMap {
		sortPosts(ps)
		tags = append(tags, Tag{
			Title: strings.Title(t),
			Posts: ps,
		})
	}
	return tags
}
