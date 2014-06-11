// Package recovery provides an HTTP handler which recovers panics in an
// underlying handler, logs debug information about the panic, and returns a 500
// Internal Server Error to the client.
package recovery

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
)

// Wrap returns an handler which proxies requests to the given handler, but
// handles panics by logging the stack trace and returning a 500 Internal Server
// Error to the client, if possible.
func Wrap(h http.Handler, l *log.Logger) http.Handler {
	return &recoveryHandler{
		h: h,
		l: l,
	}
}

type recoveryHandler struct {
	h http.Handler
	l *log.Logger
	r sync.Mutex
}

func (h *recoveryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		e := recover()
		if e != nil {
			h.r.Lock()
			defer h.r.Unlock()

			id := rand.Int63()

			body := fmt.Sprintf(
				"%s\n%016x",
				http.StatusText(http.StatusInternalServerError),
				id,
			)
			http.Error(w, body, http.StatusInternalServerError)

			h.l.Printf("panic=%016x message=%v\n", id, e)
			for skip := 1; ; skip++ {
				pc, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				if file[len(file)-1] == 'c' {
					continue
				}
				f := runtime.FuncForPC(pc)
				h.l.Printf("panic=%016x %s:%d %s()\n", id, file, line, f.Name())
			}
		}
	}()
	h.h.ServeHTTP(w, r)
}
