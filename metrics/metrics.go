// Package metrics provides an HTTP handler which registers expvar counters for
// the number of requests received and responses sent as well as quantiles of
// the latency of responses.
package metrics

import (
	"expvar"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codahale/hdrhistogram/hdr"
)

// Wrap returns a handler which records the number of requests received and
// responses sent to the given handler, as well as latency quantiles for
// responses over a five-minute window.
//
// These counters are published as the "http" object in expvars.
//
// By tracking incoming requests and outgoing responses, one can monitor not
// only the requests per second, but also the number of requests being processed
// at any given point in time.
func Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&requests, 1)        // inc requests
		defer atomic.AddUint64(&responses, 1) // inc responses when we're done
		defer recordLatency(time.Now())       // record latency when we're done

		h.ServeHTTP(w, r)
	})
}

var (
	requests, responses uint64
	m                   sync.Mutex

	// a five-minute window tracking 1ms-3min
	latency = hdr.NewWindowedHistogram(5, 1, 1000*60*3, 3)
)

func recordLatency(start time.Time) {
	m.Lock()
	defer m.Unlock()

	elapsedMS := time.Now().Sub(start).Seconds() * 1000.0
	_ = latency.Current.RecordValue(int64(elapsedMS))
}

func rotateLatency() {
	m.Lock()
	defer m.Unlock()

	latency.Rotate()
}

func getStats() httpStats {
	req, res := atomic.LoadUint64(&requests), atomic.LoadUint64(&responses)

	m.Lock()
	defer m.Unlock()

	m := latency.Merge()

	return httpStats{
		Requests:  req,
		Responses: res,
		Latency: latencyStats{
			P50:  m.ValueAtQuantile(50),
			P75:  m.ValueAtQuantile(75),
			P90:  m.ValueAtQuantile(90),
			P95:  m.ValueAtQuantile(95),
			P99:  m.ValueAtQuantile(99),
			P999: m.ValueAtQuantile(99.9),
		},
	}
}

func init() {
	go func() {
		reset := time.NewTicker(1 * time.Minute)
		for _ = range reset.C {
			rotateLatency()
		}
	}()

	expvar.Publish("http", expvar.Func(func() interface{} {
		return getStats()
	}))
}

type httpStats struct {
	Requests, Responses uint64
	Latency             latencyStats
}

type latencyStats struct {
	P50, P75, P90, P95, P99, P999 int64
}
