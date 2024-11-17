package handlers

import (
	"net/http"

	"errors"

	"github.com/devops-360-online/go-with-me/internal/models"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes" // Import codes package for status codes
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm" // Import gorm
)

func RegisterHandler(c *gin.Context) {
	tracer := otel.Tracer("event-service")                            // Initialize the tracer for the service
	ctx, span := tracer.Start(c.Request.Context(), "RegisterHandler") // Start the main span for RegisterHandler
	defer span.End()                                                  // Ensure the span ends when the function exits

	var user models.User
	_, bindSpan := tracer.Start(ctx, "BindJSON")
	if err := c.ShouldBindJSON(&user); err != nil {
		bindSpan.RecordError(err) // Record the error in tracing
		bindSpan.SetStatus(codes.Error, err.Error())
		bindSpan.SetAttributes(attribute.String("error.message", err.Error()))
		bindSpan.End()
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bindSpan.End()

	// Database Context Retrieval
	dbInterface, exists := c.Get("db")
	if !exists {
		err := errors.New("database not found in context")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.message", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not found"})
		return
	}
	db := dbInterface.(*gorm.DB)

	// Check if email already exists
	_, checkEmailSpan := tracer.Start(ctx, "CheckEmailExists")
	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		errMsg := "Email already registered"
		checkEmailSpan.RecordError(errors.New(errMsg))
		checkEmailSpan.SetStatus(codes.Error, errMsg)
		checkEmailSpan.SetAttributes(attribute.String("error.message", errMsg))
		checkEmailSpan.End()
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// An unexpected error occurred while checking the email
		checkEmailSpan.RecordError(err)
		checkEmailSpan.SetStatus(codes.Error, err.Error())
		checkEmailSpan.SetAttributes(attribute.String("error.message", err.Error()))
		checkEmailSpan.End()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email"})
		return
	}
	checkEmailSpan.End()

	// Password Hashing Span
	_, hashSpan := tracer.Start(ctx, "HashPassword")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		hashSpan.RecordError(err)
		hashSpan.SetStatus(codes.Error, err.Error())
		hashSpan.SetAttributes(attribute.String("error.message", err.Error()))
		hashSpan.End()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)
	hashSpan.End()

	_, dbSpan := tracer.Start(ctx, "CreateUserInDB")
	if err := db.Create(&user).Error; err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.SetAttributes(attribute.String("error.message", err.Error()))
		dbSpan.End()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	dbSpan.End()

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}
