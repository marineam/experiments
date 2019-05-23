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
	"testing"
)

func TestFindFiles(t *testing.T) {
	client, err := NewTestClient(nil)
	if err != nil {
		t.Fatal("Test client failed:", err)
	}
	defer client.Close()

	files, err := FindFiles(client.Client, client.URL().Path)
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
