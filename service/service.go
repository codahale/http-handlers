// Package service combines the other various packages in http-handlers to
// provide an operations-friendly http.Handler for your application. Including
// this package will also allow you to dump a full stack trace to stderr by
// sending your application the SIGUSR1 signal.
package service

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/codahale/http-handlers/debug"
	"github.com/codahale/http-handlers/logging"
	"github.com/codahale/http-handlers/metrics"
	"github.com/codahale/http-handlers/recovery"
)

// Wrap returns a new service-ready handler given an application handler.
//
// This stack application-level metrics, debug endpoints, panic recovery, and
// request logging, in that order.
func Wrap(h http.Handler) http.Handler {
	return logging.Wrap(
		recovery.Wrap(
			debug.Wrap(
				metrics.Wrap(
					h,
				),
			),
		),
		os.Stdout,
	)
}

func init() {
	dump := make(chan os.Signal)
	go func() {
		stack := make([]byte, 16*1024)
		for _ = range dump {
			n := runtime.Stack(stack, true)
			fmt.Fprintf(os.Stderr, "==== %s\n%s\n====\n", time.Now(), stack[0:n])
		}
	}()
	signal.Notify(dump, syscall.SIGUSR1)
}
