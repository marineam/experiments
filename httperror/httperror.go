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
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

const (
	bodyLimit = 64 * 1024 // Limit error document body size to 64K
)

// HTTPError represents a HTTP error document as a Go error and implements net.Error
type HTTPError struct {
	*http.Response
	msg string
}

// New wraps a Response in an HTTPError. The response's Body is replaced
// by a small buffer so reading and closing it is purely optional.
func New(resp *http.Response) *HTTPError {
	var buf bytes.Buffer

	// ignore any errors, this is only a best-effort for error messages.
	io.CopyN(&buf, resp.Body, bodyLimit)
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(&buf)

	msg := "http error: " + resp.Status
	if buf.Len() > 0 {
		// TODO: extract <body> text from HTML documents.
		switch contentType(resp) {
		case "text/plain":
			msg += ": " + bodyText2Msg(buf.Bytes())
		}
	}

	return &HTTPError{
		Response: resp,
		msg:      msg,
	}
}

func (he *HTTPError) Error() string {
	return he.msg
}

// Timeout returns true for timeout status codes.
func (he *HTTPError) Timeout() bool {
	switch he.StatusCode {
	case http.StatusRequestTimeout: // 408
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	default:
		return false
	}
}

// Temporary returns true for potentially temporary status codes such
// as server and gateway errors.
func (he *HTTPError) Temporary() bool {
	if he.Timeout() {
		return true
	}
	switch he.StatusCode {
	case http.StatusTooManyRequests: // 429
		return true
	case http.StatusInternalServerError: // 500
		return true
	case http.StatusBadGateway: // 502
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	default:
		return false
	}
}

// Return no more than 64 characters from the first line of text.
// Sufficient for use with servers returning errors via http.Error.
func bodyText2Msg(b []byte) string {
	b = bytes.TrimLeft(b, "\r\n\t ")
	if len(b) > 64 {
		b = b[:64]
	}
	i := bytes.IndexAny(b, "\r\n")
	if i > 0 {
		b = b[:i]
	}
	return string(b)
}

func contentType(resp *http.Response) string {
	typ, par, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return ""
	}

	charset, _ := par["charset"]
	charset = strings.ToLower(charset)
	// Only allow UTF-8 and ASCII since Go assumes UTF-8.
	if charset != "" && charset != "utf-8" && charset != "us-ascii" {
		return ""
	}

	return typ
}
