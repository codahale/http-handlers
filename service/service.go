// Package service combines the other various packages in http-handlers to
// provide an operations-friendly http.Handler for your application.
package service

import (
	"net/http"
	"os"

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
