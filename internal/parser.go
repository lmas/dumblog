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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v2"
)

var markdownParser = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		//emoji.New(
		//emoji.WithRenderingMethod(emoji.Entity),
		//),
		highlighting.NewHighlighting(
			// All available styles can be found at: https://github.com/alecthomas/chroma/tree/master/styles
			// (and an outdated gallery at: https://xyproto.github.io/splash/docs/all.html)
			highlighting.WithStyle("monokai"),
			highlighting.WithGuessLanguage(true),
			highlighting.WithFormatOptions(
				chromahtml.WithLineNumbers(true),
				chromahtml.LineNumbersInTable(true), // Copy-friendly lines
				chromahtml.TabWidth(8),
				// Enable this to get rid of inline css styles and add classes you can style yourself
				//chromahtml.WithClasses(true),
			),
		),
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	maxHeaderSize int64 = 1000    // 1kb, gives plenty of bytes for long titles, shorts and a lot of tags.
	maxFileSize   int64 = 5000000 // 5mb, barely enough for shakespear a la 80 hours of text (with 200wpm)
)

var headerSeparator = []byte("---") // map, slice and array can't be const

func isSeparator(line []byte) bool {
	return bytes.HasPrefix(line, headerSeparator)
}

type scanCheck func([]byte) ([]byte, bool) // Return true when you want to stop the scanning

func scan(f io.Reader, maxSize int64, check scanCheck) ([]byte, error) {
	var buf bytes.Buffer
	s := bufio.NewScanner(io.LimitReader(f, maxSize))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		b, stop := check(s.Bytes())
		if stop {
			break
		}
		if b == nil {
			continue // ignore empty lines
		}
		buf.Write(b) // Don't trim any whitespace or ur gonna mess up the markdown parsing later on
		buf.WriteRune('\n')
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func scanHeader(r io.Reader, v interface{}) error {
	separators := 0
	b, err := scan(r, maxHeaderSize, func(line []byte) ([]byte, bool) {
		if isSeparator(line) {
			separators++
			if separators == 2 {
				return nil, true
			}
			return nil, false
		}
		return line, false
	})
	switch {
	case err != nil:
		return err
	case len(b) < 1:
		return fmt.Errorf("missing header")
	case separators != 2:
		return fmt.Errorf("missing separator lines")
	}
	return yaml.UnmarshalStrict(b, v)
}

func scanBody(r io.Reader) (string, error) {
	separators := 0
	b, err := scan(r, maxFileSize, func(line []byte) ([]byte, bool) {
		if separators == 2 {
			return line, false
		}
		if isSeparator(line) {
			separators++
		}
		return nil, false
	})
	switch {
	case err != nil:
		return "", err
	case len(b) < 1:
		return "", fmt.Errorf("missing body")
	case separators != 2:
		return "", fmt.Errorf("missing separator lines")
	}
	return string(b), nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////

func readPost(path, rel string) (Post, error) {
	f, err := os.Open(path) // #nosec G304
	if err != nil {
		return Post{}, err
	}
	// It's only being read, should be safe to ignore Close() errors
	defer f.Close() // #nosec G307

	post := Post{
		source: path,
		rel:    rel,
	}
	if err := scanHeader(f, &post.Meta); err != nil {
		return Post{}, fmt.Errorf("read head: %s", err)
	}

	// TODO: verify the yaml parser trims whitespace properly
	// TODO: also do fuzz testing

	switch {
	case len(post.Meta.Title) < 1:
		return Post{}, fmt.Errorf("header is missing the title field")
	case post.Meta.Published.IsZero():
		return Post{}, fmt.Errorf("header is missing the published field")
	case len(post.Meta.Short) < 1:
		return Post{}, fmt.Errorf("header is missing the short field")
	case len(post.Meta.Tags) < 1:
		return Post{}, fmt.Errorf("header is missing the tags field")
	}

	post.Meta.Title = strings.Title(post.Meta.Title)
	for i := range post.Meta.Tags {
		post.Meta.Tags[i] = strings.ToLower(post.Meta.Tags[i])
	}
	sort.Strings(post.Meta.Tags)
	return post, nil
}
