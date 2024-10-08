package middlewares

import (
    "bytes"
    "context"
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "net/http"
)

type responseBodyWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

func CacheMiddleware(rdb *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := context.Background()
        cacheKey := c.Request.RequestURI
        val, err := rdb.Get(ctx, cacheKey).Result()
        if err == nil {
            c.Data(http.StatusOK, "application/json", []byte(val))
            c.Abort()
            return
        }

        // Use the custom response writer
        blw := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
        c.Writer = blw

        c.Next()

        // After handler
        if c.Writer.Status() == http.StatusOK {
            rdb.Set(ctx, cacheKey, blw.body.String(), 0)
        }
    }
}
