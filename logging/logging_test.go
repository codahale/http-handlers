package logging

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func mockClock() clock {
	times := []time.Time{
		time.Date(2014, 6, 3, 16, 45, 22, 36e6, time.UTC),
		time.Date(2014, 6, 3, 16, 45, 23, 43e6, time.UTC),
	}
	return func() time.Time {
		t := times[0]
		times = times[1:]
		return t
	}
}

func TestLoggingHandler(t *testing.T) {
	out := bytes.NewBuffer(nil)
	logger := Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello, world!")
		}),
		out,
	)
	logger.clock = mockClock()
	logger.Start()

	server := httptest.NewServer(logger)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	actual := string(b)
	expected := "Hello, world!"
	if actual != expected {
		t.Errorf("Response was %#v, but expected %#v", actual, expected)
	}

	logger.Stop()

	actual = out.String()
	expected = `127.0.0.1 - - [03/Jun/2014:16:45:22 +0000] "GET / HTTP/1.1" 200 0 "-" "Go 1.1 package http" 1007` + "\n"
	if actual != expected {
		t.Errorf("Log output was `%s`, but expected `%s`", actual, expected)
	}
}
