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

// Copy from FTP to Google Cloud Storage.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/marineam/experiments/network/ftputil"
	"github.com/secsy/goftp"
	"google.golang.org/api/iterator"
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
	ctx := context.Background()
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

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln("GCS client failed:", err)
	}

	target, err := url.Parse(flag.Arg(1))
	if err != nil {
		log.Fatalln("Invalid GS URL:", err)

	}
	if target.Scheme != "gs" {
		log.Fatalln("Invalid GS URL: missing gs:// prefix:", target)
	}
	if target.Host == "" {
		log.Fatalln("Invalid GS URL: missing bucket:", target)
	}

	objit := gcs.Bucket(target.Host).Objects(ctx,
		&storage.Query{Prefix: FixPrefix(target.Path)})
	objs := make(map[string]*storage.ObjectAttrs)
	for {
		obj, err := objit.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			log.Fatal("Error from goog:", err)
		}
		objs["/"+obj.Name] = obj
	}

	files, err := ftputil.FindFiles(client, server.Path)
	if err != nil {
		log.Fatalln("Listing files failed:", err)
	}

	for name, file := range files {
		if obj, ok := objs[name]; ok && QuickCheck(obj, file) {
			continue
		}

		obj := gcs.Bucket(target.Host).Object(strings.TrimPrefix(name, "/"))
		w := obj.NewWriter(ctx)
		SetModTime(w, file.ModTime())

		if err := client.Retrieve(name, w); err != nil {
			log.Fatalln("Retrieve failed:", err)
		}
		if err := w.Close(); err != nil {
			log.Fatalln("Write/Close failed:", err)
		}
		log.Println(name)
	}
}

// Metadata keys used by Google's gsutil rsync
const (
	GoogMtime = "goog-reserved-file-mtime"
	GoogGID   = "goog-reserved-posix-gid"
	GoogUID   = "goog-reserved-posix-uid"
	GoogAtime = "goog-reserved-file-atime"
	GoogMode  = "goog-reserved-posix-mode"
)

var (
	MissingModTime = errors.New("Object is missing mtime metadata")
	InvalidModTime = errors.New("Object mtime metadata is invalid")
)

// May return MissingModTime or InvalidModTime
func ObjModTime(obj *storage.ObjectAttrs) (time.Time, error) {
	mtimestr, ok := obj.Metadata[GoogMtime]
	if !ok {
		return time.Time{}, MissingModTime
	}

	mtimeint, err := strconv.ParseInt(mtimestr, 10, 64)
	if err != nil {
		return time.Time{}, InvalidModTime
	}

	// gsutil internally uses -1 to represent no mtime so
	// we always consider negative values as invalid too.
	if mtimeint <= -1 {
		return time.Time{}, InvalidModTime
	}

	return time.Unix(mtimeint, 0), nil
}

// May return InvalidModTime if mtime is a negative Unix timestamp.
func SetModTime(w *storage.Writer, mtime time.Time) error {
	mtimeint := mtime.Unix()

	// gsutil internally uses -1 to represent no mtime so
	// we always consider negative values as invalid too.
	if mtimeint <= -1 {
		return InvalidModTime
	}

	if w.Metadata == nil {
		w.Metadata = make(map[string]string)
	}

	w.Metadata[GoogMtime] = strconv.FormatInt(mtimeint, 10)
	return nil
}

// QuickCheck compares Object and FileInfo size and mtime. Returns true on match.
func QuickCheck(obj *storage.ObjectAttrs, fi os.FileInfo) bool {
	if obj == nil || fi == nil {
		return false
	}

	if !fi.Mode().IsRegular() {
		return false
	}

	if obj.Size != fi.Size() {
		return false
	}

	// Use mtime if available and valid, otherwise just skip.
	if mtime, err := ObjModTime(obj); err != nil || !mtime.Equal(fi.ModTime()) {
		return false
	}

	return true
}

// FixPrefix ensures non-empty paths end in a slash but never start with one.
func FixPrefix(p string) string {
	if p != "" && !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return strings.TrimPrefix(p, "/")
}
