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
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	filePerm fs.FileMode = 0644
	dirPerm  fs.FileMode = 0755
)

func trimDir(path, dir string) string {
	return strings.TrimPrefix(path, dir)
}

func containsDot(path string) bool {
	for _, p := range strings.Split(path, string(os.PathSeparator)) {
		if strings.HasPrefix(p, ".") {
			return true
		}
	}
	return false
}

func firstDir(path string) string {
	for i := 0; i < len(path); i++ {
		if os.IsPathSeparator(path[i]) {
			return path[:i]
		}
	}
	return ""
}

func createDir(path string) error {
	d := filepath.Dir(path)
	return os.MkdirAll(d, dirPerm)
}

func writeFile(path string, data []byte) error {
	if err := createDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, filePerm)
}

func copyFile(r, w string) error {
	src, err := os.Open(r) // #nosec G304
	if err != nil {
		return err
	}
	// Think Close() errors on read only files can be safely ignored
	defer src.Close() // #nosec G307

	if err := createDir(w); err != nil {
		return err
	}
	dst, err := os.OpenFile(w, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filePerm) // #nosec G304
	if err != nil {
		return err
	}
	// Hopefully any errors will be already caught by Sync()
	defer dst.Close() // #nosec G307

	// TODO: make sure byte count is the same as the src?
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	// Hopefully catches any file write errors before dst.Close(), see:
	// https://www.joeshaw.org/dont-defer-close-on-writable-files/
	return dst.Sync()
}
