package middlewares

import (
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/devops-360-online/go-with-me/internal/logger"
	"github.com/devops-360-online/go-with-me/internal/models"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

var IdentityKey = "id"

func AuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	tracer := otel.Tracer("event-service") // Initialize tracer for the auth service

	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "go-with-me",
		Key:         []byte("secret key"), // Replace with a secure key
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: IdentityKey,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			ctx, span := tracer.Start(c.Request.Context(), "Authenticator") // Start main Authenticator span
			defer span.End()

			var loginVals struct {
				Email    string `form:"email" json:"email" binding:"required"`
				Password string `form:"password" json:"password" binding:"required"`
			}

			// Retrieve the Trace ID from the span
			traceID := span.SpanContext().TraceID().String()

			// JSON Binding Span
			_, bindSpan := tracer.Start(ctx, "BindLoginValues")
			if err := c.ShouldBind(&loginVals); err != nil {
				bindSpan.RecordError(err)
				bindSpan.SetStatus(codes.Error, err.Error())
				bindSpan.SetAttributes(
					attribute.String("error.message", err.Error()),
					attribute.String("exception.type", fmt.Sprintf("%T", err)),
				)
				bindSpan.End()

				// Log the error
				logger.LogMessage("error", "Missing login values", traceID, map[string]interface{}{
					"error": err.Error(),
				})

				return nil, jwt.ErrMissingLoginValues
			}
			bindSpan.End()

			// Normalize email by trimming whitespace and converting to lowercase
			email := strings.ToLower(strings.TrimSpace(loginVals.Email))
			password := loginVals.Password
			span.SetAttributes(attribute.String("user.email", email)) // Track email as an attribute

			// Log the login attempt
			logger.LogMessage("info", fmt.Sprintf("Received login attempt for email: %s", email), traceID, map[string]interface{}{
				"user": email,
			})

			// Database Span to retrieve user
			dbInterface, exists := c.Get("db")
			if !exists {
				err := errors.New("database not found in context")
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.SetAttributes(
					attribute.String("error.message", err.Error()),
					attribute.String("exception.type", fmt.Sprintf("%T", err)),
				)

				// Log the error
				logger.LogMessage("error", "Database not found in context", traceID, map[string]interface{}{
					"error": err.Error(),
				})

				return nil, err
			}
			db := dbInterface.(*gorm.DB)

			_, dbSpan := tracer.Start(ctx, "GetUserByEmail") // Start a span for database retrieval
			user, err := models.GetUserByEmail(db, email)
			if err != nil {
				dbSpan.RecordError(err)
				dbSpan.SetStatus(codes.Error, "Email not found")
				dbSpan.SetAttributes(
					attribute.String("error.message", "Email not found"),
					attribute.String("exception.type", fmt.Sprintf("%T", err)),
				)
				dbSpan.End()

				// Log the error
				logger.LogMessage("error", fmt.Sprintf("User not found: %s", email), traceID, map[string]interface{}{
					"user":  email,
					"error": err.Error(),
				})

				// Return a custom authentication error
				authErr := &AuthenticationError{Message: "Incorrect email or password"}
				return nil, authErr
			}
			dbSpan.End()

			// Log that user was found
			logger.LogMessage("info", fmt.Sprintf("User found: %s", user.Email), traceID, map[string]interface{}{
				"user": user.Email,
			})

			// Password Verification Span
			_, verifySpan := tracer.Start(ctx, "VerifyPassword")

			if !models.CheckPasswordHash(password, user.Password) {
				err := &AuthenticationError{Message: "Incorrect email or password"}
				verifySpan.RecordError(err)
				verifySpan.SetStatus(codes.Error, err.Error())
				verifySpan.SetAttributes(
					attribute.String("error.message", "Password verification failed"),
					attribute.String("exception.type", fmt.Sprintf("%T", err)),
				)
				verifySpan.End()

				// Log the error
				logger.LogMessage("error", fmt.Sprintf("Password verification failed for user: %s", user.Email), traceID, map[string]interface{}{
					"user":  user.Email,
					"error": err.Error(),
				})

				return nil, err
			}
			verifySpan.End()

			// Log successful authentication
			logger.LogMessage("info", fmt.Sprintf("Password verified for user: %s", user.Email), traceID, map[string]interface{}{
				"user": user.Email,
			})

			return user, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true // You can add authorization logic and trace it if needed
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					IdentityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			_, span := tracer.Start(c.Request.Context(), "IdentityHandler") // Span for IdentityHandler
			defer span.End()

			// Retrieve the Trace ID
			traceID := span.SpanContext().TraceID().String()

			claims := jwt.ExtractClaims(c)
			userID := uint(claims[IdentityKey].(float64))
			span.SetAttributes(attribute.Int64("user.id", int64(userID))) // Track user ID

			// Log identity handling
			logger.LogMessage("info", fmt.Sprintf("IdentityHandler invoked for user ID: %d", userID), traceID, map[string]interface{}{
				"user_id": userID,
			})

			return &models.User{
				ID: userID,
			}
		},
	})
}
