// Code generated by microgen 0.9.0. DO NOT EDIT.

package service

import (
	"context"
	log "github.com/go-kit/kit/log"
	service "github.com/dreamsxin/go-kitcli/examples/addsvc/addsvc"
)

// ErrorLoggingMiddleware writes to logger any error, if it is not nil.
func ErrorLoggingMiddleware(logger log.Logger) Middleware {
	return func(next service.Service) service.Service {
		return &errorLoggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type errorLoggingMiddleware struct {
	logger log.Logger
	next   service.Service
}

func (M errorLoggingMiddleware) Sum(ctx context.Context, a int, b int) (result int, err error) {
	defer func() {
		if err != nil {
			M.logger.Log("method", "Sum", "message", err)
		}
	}()
	return M.next.Sum(ctx, a, b)
}

func (M errorLoggingMiddleware) Concat(ctx context.Context, a string, b string) (result string, err error) {
	defer func() {
		if err != nil {
			M.logger.Log("method", "Concat", "message", err)
		}
	}()
	return M.next.Concat(ctx, a, b)
}
