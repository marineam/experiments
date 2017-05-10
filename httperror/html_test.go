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
	"testing"
)

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
	apacheTrivial404Text = `Not Found The requested URL /404 was not found on this server. Apache Server at localhost Port 80`
	apacheFancy404Type   = `text/html; charset=iso-8859-1`
	apacheFancy404Body   = `<?xml version="1.0" encoding="UTF-8"?>
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
	apacheFancy404Text = `Object not found! The requested URL was not found on this server. If you entered the URL manually please check your spelling and try again. If you think this is a server error, please contact the webmaster. Error 404 localhost Apache`
)

func TestHTMLTextTrivial(t *testing.T) {
	const expect = apacheTrivial404Text
	text, err := htmlText([]byte(apacheTrivial404Body))
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != expect {
		t.Fatalf("text != expect:\n%s\n%s", text, expect)
	}
}

func TestHTMLTextFancy(t *testing.T) {
	const expect = apacheFancy404Text
	text, err := htmlText([]byte(apacheFancy404Body))
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != expect {
		t.Fatalf("text != expect:\n%s\n%s", text, expect)
	}
}
