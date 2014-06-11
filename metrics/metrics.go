// Package metrics provides an HTTP handler which registers atomic expvars for
// the number of requests received and responses sent.
package metrics

import (
	"expvar"
	"fmt"
	"net/http"
	"sync/atomic"
)

// Wrap returns a handler which records the number of requests received and
// responses sent to the given handler.
//
// These counters are published as http_requests and http_responses in expvars.
//
// By tracking incoming requests and outgoing responses, one can monitor not
// only the requests per second, but also the number of requests being processed
// at any given point in time.
func Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter.inc()
		defer responseCounter.inc()

		h.ServeHTTP(w, r)
	})
}

type atomicCounter struct {
	n uint64
}

func (a *atomicCounter) inc() {
	atomic.AddUint64(&a.n, 1)
}

func (a *atomicCounter) clear() {
	atomic.StoreUint64(&a.n, 0)
}

func (a atomicCounter) String() string {
	n := atomic.LoadUint64(&a.n)
	return fmt.Sprintf("%d", n)
}

var (
	requestCounter  = atomicCounter{}
	responseCounter = atomicCounter{}
)

func init() {
	expvar.Publish("http_requests", requestCounter)
	expvar.Publish("http_responses", responseCounter)
}
