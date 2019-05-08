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
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/secsy/goftp"
)

var (
	debug    = flag.Bool("debug", false, "output protocol debug to stderr")
	user     = flag.String("user", "anonymous", "ftp user name")
	password = flag.String("password", "anonymous", "ftp password")
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

func main() {
	flag.Parse()
	server, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatalln("Invalid URL:", err)
	}
	if server.Scheme != "ftp" && server.Scheme != "ftps" {
		log.Fatalln("Invalid URL: missing ftp:// or ftps:// prefix:", server)
	}
	if server.Hostname() == "" {
		log.Fatalln("Invalid URL: missing host name:", server)
	}

	config := goftp.Config{
		User:     *user,
		Password: *password,
	}
	if server.Scheme == "ftps" {
		config.TLSConfig = &tls.Config{
			//ServerName: server.Host,
			InsecureSkipVerify: true,
		}
	}
	if *debug {
		config.Logger = os.Stderr
		log.Println("Connecting to", server.Host)
	}

	client, err := goftp.DialConfig(config, server.Host)
	if err != nil {
		log.Fatalln("Client failed:", err)
	}
	defer client.Close()

	var walk func(parent string) (files []*File)
	walk = func(parent string) (files []*File) {
		entries, err := client.ReadDir(parent)
		if err != nil {
			log.Fatalln("List error:", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				if entry.Name() != "." && entry.Name() != ".." {
					files = append(files, walk(path.Join(parent, entry.Name()))...)
				}
			} else {
				files = append(files, &File{entry, parent})
			}
		}

		return files
	}

	files := walk(server.Path)
	for _, file := range files {
		fmt.Println(*file)
	}
}
