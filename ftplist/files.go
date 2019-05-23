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
	"os"
	"path"

	"github.com/secsy/goftp"
)

// Find all files under a given path on an FTP server.  Aborts on any error.
// May want to skip inaccessable directories and similar things in the future.
func FindFiles(client *goftp.Client, root string) (map[string]os.FileInfo, error) {
	entries, err := client.ReadDir(root)
	if err != nil {
		return nil, err
	}

	files := make(map[string]os.FileInfo)
	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == "." || entry.Name() == ".." {
				continue
			}
			subfiles, err := FindFiles(client, path.Join(root, entry.Name()))
			if err != nil {
				return nil, err
			}
			for subpath, subentry := range subfiles {
				files[subpath] = subentry
			}
		} else {
			files[path.Join(root, entry.Name())] = entry
		}
	}

	return files, nil
}
