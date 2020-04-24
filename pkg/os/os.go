/*
 Copyright © 2020 Travis Clarke <travis.m.clarke@gmail.com>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in
 all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 THE SOFTWARE.
*/

package os

import (
	"os"
	"path/filepath"
	"regexp"
)

// RenameFile renames a file based on a specified regular expression pattern.
func RenameFile(pat string, src string, repl string) string {
	s := regexp.MustCompile(pat).ReplaceAllString(src, repl)
	return regexp.MustCompile(`^[^\w\d]`).ReplaceAllString(s, "")
}

// HasExtension checks if a file's extension matches a pattern.
func HasExtension(path string, pat string) bool {
	return regexp.MustCompile(pat).MatchString(filepath.Ext(path))
}

// Exists checks if a path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsFile checks if a path exists and is a file.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsRegular()
}

// IsDirectory checks if a path exists and is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsDir()
}
