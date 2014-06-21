// Package metrics provides an HTTP handler which registers atomic expvars for
// the number of requests received and responses sent.
package metrics

import (
	"expvar"
	"net/http"
	"sync/atomic"
)

// Wrap returns a handler which records the number of requests received and
// responses sent to the given handler.
//
// These counters are published as http.Requests and http.Responses in expvars.
//
// By tracking incoming requests and outgoing responses, one can monitor not
// only the requests per second, but also the number of requests being processed
// at any given point in time.
func Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&requests, 1)
		defer atomic.AddUint64(&responses, 1)

		h.ServeHTTP(w, r)
	})
}

var (
	requests, responses uint64
)

func init() {
	expvar.Publish("http", expvar.Func(func() interface{} {
		return map[string]uint64{
			"Requests":  atomic.LoadUint64(&requests),
			"Responses": atomic.LoadUint64(&responses),
		}
	}))
}
