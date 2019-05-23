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

package main

import (
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/marineam/experiments/network/inetd"
	"github.com/secsy/goftp"
)

func TestFindFiles(t *testing.T) {
	inetd, err := inetd.Listen("tcp", "localhost:0", "./testdata/ftpd.sh", "--tls=3")
	if err != nil {
		t.Fatal("Test ftpd failed:", err)
	}
	defer inetd.Close()

	server := url.URL{
		Scheme: "ftps",
		Host:   inetd.Addr().String(),
		Path:   "/",
	}
	config := goftp.Config{
		TLSConfig: &tls.Config{
			ServerName:         server.Hostname(),
			InsecureSkipVerify: true,
		},
	}
	//config.Logger = os.Stderr

	client, err := goftp.DialConfig(config, server.Host)
	if err != nil {
		t.Fatal("Client failed:", err)
	}
	defer client.Close()

	files, err := FindFiles(client, server.Path)
	if err != nil {
		t.Fatal("Listing files failed:", err)
	}

	pem, ok := files["/ftpd.pem"]
	if !ok {
		t.Fatal("ftpd.pem missing from testdata listing")
	}
	if !pem.Mode().IsRegular() {
		t.Errorf("ftpd.pem has unexpected mode: %s", pem.Mode())
	}
	if pem.Size() != 2803 {
		t.Errorf("ftpd.pem has size %d, expected 2803", pem.Size())
	}
}
