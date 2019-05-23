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
	"io"
	"net/url"
	"path/filepath"
	"runtime"

	"github.com/marineam/experiments/network/inetd"
	"github.com/secsy/goftp"
)

type TestClient struct {
	*goftp.Client
	inetd *inetd.Inetd
}

func testDataPath(elem ...string) string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		filename = filepath.Join(".", "nope.go")
	}
	elem = append([]string{filepath.Dir(filename), "testdata"}, elem...)
	return filepath.Join(elem...)
}

func NewTestClient(log io.Writer) (*TestClient, error) {
	inetd, err := inetd.Listen("tcp", "localhost:0", testDataPath("ftpd.sh"), "--tls=3")
	if err != nil {
		return nil, err
	}

	config := goftp.Config{
		Logger: log,
		TLSConfig: &tls.Config{
			ServerName:         "localhost",
			InsecureSkipVerify: true,
		},
	}

	client, err := goftp.DialConfig(config, inetd.Addr().String())
	if err != nil {
		inetd.Close()
		return nil, err
	}

	return &TestClient{
		Client: client,
		inetd:  inetd,
	}, nil
}

func (tc *TestClient) URL() *url.URL {
	return &url.URL{
		Scheme: "ftps",
		Host:   tc.inetd.Addr().String(),
		Path:   "/",
	}
}

func (tc *TestClient) Close() error {
	ierr := tc.inetd.Close()
	cerr := tc.Client.Close()
	if cerr != nil {
		return cerr
	} else if ierr != nil {
		return ierr
	}
	return nil
}
