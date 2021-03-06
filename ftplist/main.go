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
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/marineam/experiments/network/ftputil"
	"github.com/secsy/goftp"
)

var (
	dummy    = flag.Bool("dummy", false, "launch our own ftp server for testing")
	debug    = flag.Bool("debug", false, "output protocol debug to stderr")
	insecure = flag.Bool("insecure", false, "disable TLS server name verification")
	timeout  = flag.Duration("timeout", 5*time.Second, "timeout for all operations")
	user     = flag.String("user", "anonymous", "ftp user name")
	password = flag.String("password", "anonymous", "ftp password")
)

func main() {
	flag.Parse()

	var server *url.URL
	var client *goftp.Client
	if *dummy {
		var logger io.Writer
		if *debug {
			logger = os.Stderr
		}
		tc, err := ftputil.NewTestClient(logger)
		if err != nil {
			log.Fatalln("Failed to setup test client:", err)
		}
		server = tc.URL()
		client = tc.Client
		defer tc.Close()
	} else {
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
			Timeout:  *timeout,
		}
		if server.Scheme == "ftps" {
			config.TLSConfig = &tls.Config{
				ServerName:         server.Hostname(),
				InsecureSkipVerify: *insecure,
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
	}

	files, err := ftputil.FindFiles(client, server.Path)
	if err != nil {
		log.Fatalln("Listing files failed:", err)
	}

	for name, file := range files {
		fmt.Println(name, file)
	}
}
