package main

import (
	"github.com/marineam/experiments/httperror"
	"io"
	"log"
	"net/http"
	"os"
)

// get returns an error for non-200 responses.
func get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = httperror.New(resp) // closes original resp.Body
	}
	return resp, err
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s http://example.com", os.Args[0])
	}

	resp, err := get(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatal(err)
	}
}
