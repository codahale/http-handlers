package recovery

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryHandler(t *testing.T) {
	out := bytes.NewBuffer(nil)
	l := log.New(out, "", log.LstdFlags)
	recovery := Wrap(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("aaaaugh")
		}),
		l,
	)

	server := httptest.NewServer(recovery)
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
	if !strings.HasPrefix(actual, "Internal Server Error") {
		t.Errorf("Unexpected response: %#v", actual)
	}

	actual = out.String()
	t.Log("\n" + actual)
	if !strings.Contains(actual, "aaaaugh") {
		t.Errorf("Unexpected error output: %#v", actual)
	}
}
