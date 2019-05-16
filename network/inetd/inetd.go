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

// inetd implemented for Go
package inetd

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/marineam/experiments/network/neterror"
)

type Inetd struct {
	listener net.Listener
	program  string
	args     []string
}

func Listen(network, address, program string, arg ...string) (*Inetd, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	i := &Inetd{
		listener: listener,
		program:  program,
		args:     arg,
	}

	go func() {
		for {
			if err := i.accept(); err != nil {
				if !neterror.IsClosed(err) {
					fmt.Fprintf(os.Stderr, "inetd: accept failed: %s\n", err)
				}
				return
			}
		}
	}()

	return i, nil
}

func (i *Inetd) accept() error {
	conn, err := i.listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	// os.exec will reuse the fd from os.File but not net.Conn
	var fconn *os.File
	switch conn := conn.(type) {
	case *net.TCPConn:
		fconn, err = conn.File()
	case *net.UnixConn:
		fconn, err = conn.File()
	default:
		err = fmt.Errorf("unknown connection type: %T", conn)
	}
	if err != nil {
		return err
	}
	defer fconn.Close()

	cmd := exec.Command(i.program, i.args...)
	cmd.Stdin = fconn
	cmd.Stdout = fconn
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("inetd: %s", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "inetd: %s exited with %s\n", i.program, err)
		}
	}()

	return nil
}

func (i *Inetd) Dial() (net.Conn, error) {
	addr := i.listener.Addr()
	return net.Dial(addr.Network(), addr.String())
}

func (i *Inetd) Addr() net.Addr {
	return i.listener.Addr()
}

func (i *Inetd) Close() error {
	return i.listener.Close()
}
