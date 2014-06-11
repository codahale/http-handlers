package metrics

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCountingHandler(t *testing.T) {
	requestCounter.clear()
	responseCounter.clear()

	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, world")
	}))
	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := http.Get(s.URL + "/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Status code was %d, but expected 200", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	actual := string(b)
	expected := "hello, world\n"
	if actual != expected {
		t.Errorf("Response was %q, but expected %q", actual, expected)
	}

	actual = requestCounter.String()
	expected = "1"
	if actual != expected {
		t.Errorf("Request counter was %q, but expected %s", actual, expected)
	}

	actual = responseCounter.String()
	expected = "1"
	if actual != expected {
		t.Errorf("Response counter was %q, but expected %s", actual, expected)
	}
}
