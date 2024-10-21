package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

// TracingMiddleware creates a trace span for each incoming request
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the tracer
		tracer := otel.Tracer("event-service")

		// Start a new span for the incoming request
		ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
		defer span.End()

		// Attach the span context to the request
		c.Request = c.Request.WithContext(ctx)

		// Continue processing the request
		c.Next()
	}
}
