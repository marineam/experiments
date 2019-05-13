// Copyright 2019 Michael Marineau
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Test program for accessing FTP servers.
package main

import (
	"fmt"
	"os"
	"path"

	"github.com/secsy/goftp"
)

type File struct {
	os.FileInfo
	Parent string
}

func (f File) Path() string {
	return path.Join(f.Parent, f.Name())
}

func (f File) String() string {
	return fmt.Sprint(f.Path(), f.FileInfo)
}

// Find all files under a given path on an FTP server.  Aborts on any error.
// May want to skip inaccessable directories and similar things in the future.
func FindFiles(client *goftp.Client, root string) (files []*File, err error) {
	entries, err := client.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "." || entry.Name() == ".." {
				continue
			}
			subfiles, err := FindFiles(client, path.Join(root, entry.Name()))
			if err != nil {
				return nil, err
			}
			files = append(files, subfiles...)
		} else {
			files = append(files, &File{entry, root})
		}
	}

	return files, nil
}
