// Copyright 2017 CoreOS, Inc.
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

package httperror

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoTextError(t *testing.T) {
	for _, tt := range []struct {
		code      int
		timeout   bool
		temporary bool
	}{{
		code: http.StatusTeapot,
	}, {
		code:      http.StatusGatewayTimeout,
		timeout:   true,
		temporary: true,
	}, {
		code:      http.StatusInternalServerError,
		temporary: true,
	}} {
		rec := httptest.NewRecorder()
		http.Error(rec, "something failed", tt.code)
		err := New(rec.Result())
		if !strings.HasSuffix(err.Error(), ": something failed") {
			t.Errorf("missing error message: %v", err)
		}
		if err.Timeout() != tt.timeout {
			t.Errorf("Timeout() was %s, expected %s",
				err.Timeout(), tt.timeout)
		}
		if err.Temporary() != tt.temporary {
			t.Errorf("Temporary() was %s, expected %s",
				err.Timeout(), tt.temporary)
		}
	}
}

func TestApacheTrivialHTML(t *testing.T) {
	const expect = `http error: Not Found: Not Found The requested URL /404 was not found on this server. A`
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", apacheTrivial404Type)
	rec.WriteHeader(http.StatusNotFound)
	rec.Write([]byte(apacheTrivial404Body))
	err := New(rec.Result())
	if err.Error() != expect {
		t.Errorf("err != expect:\n%s\n%s", err.Error(), expect)
	}
}

func TestApacheFancyHTML(t *testing.T) {
	const expect = `http error: Not Found: Object not found! The requested URL was not found on this server`
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", apacheFancy404Type)
	rec.WriteHeader(http.StatusNotFound)
	rec.Write([]byte(apacheFancy404Body))
	err := New(rec.Result())
	if err.Error() != expect {
		t.Errorf("err != expect:\n%s\n%s", err.Error(), expect)
	}
}

func ExampleHTTPError() {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	go (&http.Server{}).Serve(l)
	defer l.Close()

	// get returns an error for non-200 responses.
	get := func(url string) (*http.Response, error) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode != http.StatusOK {
			err = New(resp) // closes original resp.Body
		}
		return resp, err
	}

	url := fmt.Sprintf("http://%s/404", l.Addr())
	resp, err := get(url)
	for i := 0; i < 3; i++ {
		// retry after temporary network/http errors.
		if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
			resp, err = get(url)
		} else {
			break
		}
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	// use response

	// Output:
	// http error: 404 Not Found: 404 page not found
}
