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
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// htmlText strips down the body of an HTML document to a single line of text.
// charset may be declared in content type header or document or both.
func htmlText(b []byte, charset string) string {
	var reader io.Reader
	switch strings.ToLower(charset) {
	case "", "utf-8", "us-ascii":
		reader = bytes.NewReader(b)
	case "iso-8859-1":
		reader = newLatin1Reader(bytes.NewReader(b))
	default:
		// give up
		return ""
	}

	dec := xml.NewDecoder(reader)
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = func(enc string, r io.Reader) (io.Reader, error) {
		switch strings.ToLower(enc) {
		case "us-ascii":
			return r, nil
		case "iso-8859-1":
			if charset == "iso-8859-1" {
				// charset was already declared and translated
				return r, nil
			}
			return newLatin1Reader(r), nil
		default:
			return nil, errors.New("unsupported encoding")
		}
	}

	var depth int
	var raw []byte
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch tok := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(tok.Name.Local)
			if depth == 0 && name != "html" {
				// Not an html document
				break
			}

			if depth == 1 && name != "body" {
				if err := dec.Skip(); err != nil {
					break
				}
				continue
			}

			if name == "script" || name == "style" {
				if err := dec.Skip(); err != nil {
					break
				}
				continue
			}

			depth++

		case xml.EndElement:
			// Ensure block elements always break up text.
			switch strings.ToLower(tok.Name.Local) {
			case "h1", "h2", "h3", "h4", "h5", "h6",
				"br", "blockquote", "div", "p":
				raw = append(raw, ' ')
			}

			depth--

		case xml.CharData:
			raw = append(raw, tok...)

		case xml.ProcInst:
		case xml.Directive:
		}
	}

	// Chomp all consecutive spaces, constructing a single line of text.
	text := make([]byte, 0, len(raw))
	wasSpace := true
	for len(raw) > 0 {
		r, l := utf8.DecodeRune(raw)
		if unicode.IsSpace(r) {
			if !wasSpace {
				text = append(text, ' ')
				wasSpace = true
			}
		} else {
			text = append(text, raw[:l]...)
			wasSpace = false
		}
		raw = raw[l:]
	}
	// Trim off trailing space
	if len(text) > 0 && wasSpace {
		text = text[:len(text)-1]
	}

	return string(text)
}

type latin1Reader struct {
	buf *bufio.Reader
}

func newLatin1Reader(r io.Reader) io.Reader {
	return &latin1Reader{buf: bufio.NewReader(r)}
}

func (l *latin1Reader) Read(b []byte) (n int, err error) {
	for {
		c, err := l.buf.ReadByte()
		if err != nil {
			return n, err
		}
		if n+utf8.RuneLen(rune(c)) > len(b) {
			l.buf.UnreadByte()
			return n, nil
		}
		n += utf8.EncodeRune(b[n:], rune(c))
	}
}
