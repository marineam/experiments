// Copyright 2015 CoreOS, Inc.
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

// Test program for checking the behavior of thread-local attributes.
//
// We have been using LockOSThread heavily to ensure a thread attribute
// such as a namespace or uid/gid remains the same through a goroutine.
// However we have not been sure how worried we need to be about the
// attribute leaking into other goroutines. Turns out we need to worry:
//
//  $ go build
//  $ sudo ./go-danger
//  2015/06/03 10:25:04 goroutine  0 as egid  0 on thread 23331
//  2015/06/03 10:25:04 goroutine  1 as egid  1 on thread 23331
//  2015/06/03 10:25:04 goroutine  2 as egid  2 on thread 23333
//  2015/06/03 10:25:04 goroutine  0 as egid  2 on thread 23335
//  2015/06/03 10:25:04 mismatch!
//
// This will happen regardless of the value of GOMAXPROCS which limits
// the number of currently running threads, not the number of threads
// as I had previously assumed. (That being the case not sure how I
// hit deadlocks using LockOSThread in the past but I'll test that some
// other day...)
//
// So the take away here is: the moment you lock to a thread in order
// to modify a thread-local attribute, from then on in the Go program
// any goroutine that needs the original attribute value *must* also
// lock to a thread and explicitly set the attribute back. Ick.
//
// Note: this program performs this check using egid which is thread-
// local in the Linux kernel. This may seem confusing to a C programmer
// because POSIX states otherwise. Well it turns out that NPTL uses
// signals to stop all threads and set egid when you call seteuid(2).
// However, in Go this doesn't happen because Go avoids using libc.
package main

import (
	"log"
	"runtime"
	"syscall"
	"time"
)

const (
	threadCount    = 1
	goroutineCount = 2
)

type info struct {
	goid int
	gid  int
	tid  int
}

func locked(goid int, infoc chan<- info) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := syscall.Setregid(-1, goid)
	if err != nil {
		panic(err)
	}

	for {
		infoc <- info{goid, syscall.Getegid(), syscall.Gettid()}
		time.Sleep(time.Millisecond)
	}
}

func leaked(infoc chan<- info) {
	for {
		// the lock is only here to ensure Getegid and Gettid
		// are called from within the same thread.
		runtime.LockOSThread()
		i := info{0, syscall.Getegid(), syscall.Gettid()}
		runtime.UnlockOSThread()

		infoc <- i
		time.Sleep(time.Millisecond)
	}
}

func main() {
	runtime.GOMAXPROCS(threadCount)

	// we abuse egid so must start off as root
	if syscall.Getegid() != 0 {
		log.Fatal("Must run as root!")
	}

	infoc := make(chan info)
	go leaked(infoc)
	for goid := 1; goid <= goroutineCount; goid++ {
		go locked(goid, infoc)
	}

	for i := range infoc {
		log.Printf("goroutine %2d as egid %2d on thread %d", i.goid, i.gid, i.tid)
		if i.goid != i.gid {
			log.Fatalf("mismatch!")
		}
	}
}
