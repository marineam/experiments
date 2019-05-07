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
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/jlaffaye/ftp"
)

var (
	debug    = flag.Bool("debug", false, "output protocol debug to stderr")
	user     = flag.String("user", "anonymous", "ftp user name")
	password = flag.String("password", "anonymous", "ftp password")
)

type File struct {
	Path string
	Size uint64
	Time time.Time
}

func newFile(parent string, entry *ftp.Entry) *File {
	return &File{
		Path: path.Join(parent, entry.Name),
		Size: entry.Size,
		Time: entry.Time,
	}
}

func main() {
	flag.Parse()
	server, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatalln("Invalid URL:", err)
	}
	if server.Hostname() == "" {
		log.Fatalln("Invalid URL: missing host name:", server)
	}

	options := []ftp.DialOption{
		ftp.DialWithTimeout(5 * time.Second),
	}

	if *debug {
		options = append(options, ftp.DialWithDebugOutput(os.Stderr))
		log.Println("Connecting to", server.Host)
	}

	conn, err := ftp.Dial(server.Host, options...)
	if err != nil {
		log.Fatalln("Connection failed:", err)
	}
	defer conn.Quit()

	if err := conn.Login(*user, *password); err != nil {
		log.Fatalln("Login failed:", err)
	}

	var walk func(parent string) (files []*File)
	walk = func(parent string) (files []*File) {
		entries, err := conn.List(parent)
		if err != nil {
			log.Fatalln("List error:", err)
		}

		for _, entry := range entries {
			switch entry.Type {
			case ftp.EntryTypeFile:
				files = append(files, newFile(parent, entry))
			case ftp.EntryTypeFolder:
				if entry.Name != "." && entry.Name != ".." {
					files = append(files, walk(path.Join(parent, entry.Name))...)
				}
			case ftp.EntryTypeLink:
				// ignore for now
			}
		}

		return files
	}

	files := walk(server.Path)
	for _, file := range files {
		fmt.Println(*file)
	}
}
