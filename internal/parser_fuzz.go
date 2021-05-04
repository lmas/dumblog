// +build gofuzz

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
)

func Fuzz(data []byte) int {
	// I'm running both scanners in the same func like this, as it makes is easier for now to run the fuzzer and
	// friends as a single instance.
	body := bytes.NewBuffer(data)

	var post Post
	err := scanHeader(body, &post)
	if err != nil ||
		len(post.Meta.Title) < 1 ||
		post.Meta.Published.IsZero() ||
		len(post.Meta.Short) < 1 ||
		len(post.Meta.Tags) < 1 {
		return 0
	}

	var buf bytes.Buffer
	if err := scanBody(body, &buf); err != nil {
		return 0
	}

	return 1
}
