package metrics

import (
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMetrics(t *testing.T) {
	requests = 0
	responses = 0

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

	a := string(b)
	e := "hello, world\n"
	if a != e {
		t.Errorf("Response was %q, but expected %q", a, e)
	}

	// jiggle the handle to avoid timing-dependent tests
	latency.Reset()
	latency.Insert(1)

	var actual httpStats
	if err := json.Unmarshal([]byte(expvar.Get("http").String()), &actual); err != nil {
		t.Fatal(err)
	}

	expected := httpStats{
		Requests:  1,
		Responses: 1,
		Latency: latencyStats{
			P50:  1,
			P75:  1,
			P90:  1,
			P95:  1,
			P99:  1,
			P999: 1,
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Was %#v, but expected %#v", actual, expected)
	}
}

func BenchmarkMetrics(b *testing.B) {
	var (
		r *http.Request
		w http.ResponseWriter
		h = Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.ServeHTTP(w, r)
		}
	})
}
