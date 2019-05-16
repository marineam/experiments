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

package inetd

import (
	"bytes"
	"io/ioutil"
	"net"
	"testing"
)

func TestListen(t *testing.T) {
	i, err := Listen("tcp", "localhost:0", "false")
	if err != nil {
		t.Fatal(err)
	}

	if err := i.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestDial(t *testing.T) {
	i, err := Listen("tcp", "localhost:0", "cat")
	if err != nil {
		t.Fatal(err)
	}
	defer i.Close()

	c, err := i.Dial()
	if err != nil {
		t.Fatal(err)
	}

	writedata := []byte("test")
	if _, err := c.Write(writedata); err != nil {
		t.Fatal(err)
	}
	if err := c.(*net.TCPConn).CloseWrite(); err != nil {
		t.Fatal(err)
	}

	readdata, err := ioutil.ReadAll(c)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(writedata, readdata) {
		t.Errorf("%s != %s", string(writedata), string(readdata))
	}
}
