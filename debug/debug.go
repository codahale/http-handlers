// Package debug provides a handler which adds additional debug endpoints,
// including profiling and expvars.
package debug

// BUG(coda): Figure out how to test /debug/pprof/profile

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
)

// Wrap returns a handler which adds the following URLs as special cases:
//
//     /debug/pprof/        -- an HTML index of pprof endpoints
//     /debug/pprof/cmdline -- the running process's command line
//     /debug/pprof/profile -- pprof profiling endpoint
//     /debug/pprof/symbol  -- pprof debugging symbols
//     /debug/vars          -- JSON-formatted expvars
func Wrap(handler http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/vars", expvarHandler)
	mux.HandleFunc("/debug/gc", performGC)
	mux.Handle("/", handler)
	return mux
}

// Lifted entirely from expvar.go, which is a shame. This manually generates the
// JSON response in part because string representations of custom expvars are
// intended to be JSON.
func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func performGC(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "running GC...")
	runtime.GC()
	fmt.Fprintln(w, "done!")
}
