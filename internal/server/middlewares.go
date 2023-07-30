package server

import (
	"github.com/mojixcoder/kid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// NewTraceMiddleware returns a new trace middleware.
func NewTraceMiddleware() kid.MiddlewareFunc {
	propagator := otel.GetTextMapPropagator()

	return func(next kid.HandlerFunc) kid.HandlerFunc {
		return func(c *kid.Context) {
			ctx := propagator.Extract(
				c.Request().Context(),
				propagation.HeaderCarrier(c.Request().Header),
			)

			c.Set("ctx", ctx)

			next(c)
		}
	}
}
