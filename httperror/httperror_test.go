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

// extracted from httpd-2.4.25
const (
	apacheTrivial404Type = `text/html; charset=iso-8859-1`
	apacheTrivial404Body = `<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN">
<html><head>
<title>404 Not Found</title>
</head><body>
<h1>Not Found</h1>
<p>The requested URL /404 was not found on this server.</p>
<hr>
<address>Apache Server at localhost Port 80</address>
</body></html>
`
	apacheFancy404Type = `text/html; charset=iso-8859-1`
	apacheFancy404Body = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
  "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en">
<head>
<title>Object not found!</title>
<link rev="made" href="mailto:root@localhost" />
<style type="text/css"><!--/*--><![CDATA[/*><!--*/ 
    body { color: #000000; background-color: #FFFFFF; }
    a:link { color: #0000CC; }
    p, address {margin-left: 3em;}
    span {font-size: smaller;}
/*]]>*/--></style>
</head>

<body>
<h1>Object not found!</h1>
<p>


    The requested URL was not found on this server.

  

    If you entered the URL manually please check your
    spelling and try again.

  

</p>
<p>
If you think this is a server error, please contact
the <a href="mailto:root@localhost">webmaster</a>.

</p>

<h2>Error 404</h2>
<address>
  <a href="/">localhost</a><br />
  <span>Apache</span>
</address>
</body>
</html>
`
)

func TestApacheTrivialHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", apacheTrivial404Type)
	rec.WriteHeader(http.StatusNotFound)
	rec.Write([]byte(apacheTrivial404Body))
	New(rec.Result())
	t.Skip("iso-8859-1 and html not implemented")
}

func TestApacheFancyHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", apacheFancy404Type)
	rec.WriteHeader(http.StatusNotFound)
	rec.Write([]byte(apacheFancy404Body))
	New(rec.Result())
	t.Skip("html not implemented")
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
