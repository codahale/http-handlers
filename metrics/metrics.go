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

	"github.com/bmizerany/perks/quantile"
)

// Wrap returns a handler which records the number of requests received and
// responses sent to the given handler, as well as latency quantiles for
// responses which are reset every minute.
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
	latency             *quantile.Stream
	latencyMutex        sync.Mutex
)

func recordLatency(start time.Time) {
	latencyMutex.Lock()
	defer latencyMutex.Unlock()

	latency.Insert(time.Now().Sub(start).Seconds() * 1000.0)
}

func resetLatency() {
	latencyMutex.Lock()
	defer latencyMutex.Unlock()

	latency.Reset()
}

func getStats() httpStats {
	req, res := atomic.LoadUint64(&requests), atomic.LoadUint64(&responses)

	latencyMutex.Lock()
	defer latencyMutex.Unlock()

	return httpStats{
		Requests:  req,
		Responses: res,
		Latency: latencyStats{
			P50:  latency.Query(0.50),
			P75:  latency.Query(0.75),
			P90:  latency.Query(0.90),
			P95:  latency.Query(0.95),
			P99:  latency.Query(0.99),
			P999: latency.Query(0.999),
		},
	}
}

func init() {
	latency = quantile.NewTargeted(0.50, 0.75, 0.90, 0.95, 0.99, 0.999)

	go func() {
		reset := time.NewTicker(1 * time.Minute)
		for _ = range reset.C {
			resetLatency()
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
	P50, P75, P90, P95, P99, P999 float64
}
