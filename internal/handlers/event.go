package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/devops-360-online/go-with-me/config"
	"github.com/devops-360-online/go-with-me/internal/middlewares"
	"github.com/devops-360-online/go-with-me/internal/models"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ListEventsHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Get pagination parameters from query
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var events []models.Event
	if err := db.Preload("Users").Offset((page - 1) * pageSize).Limit(pageSize).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

func GetEventHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	eventIDParam := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := db.Preload("Users").First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

// UploadFileToS3 uploads a file to S3 (or LocalStack in your case)
func UploadFileToS3(ctx context.Context, cfg *config.Config, file multipart.File, fileHeader *multipart.FileHeader, eventID uint, eventName string) (string, error) {
	tracer := otel.Tracer("event-service") // Get the tracer
	ctx, span := tracer.Start(ctx, "UploadFileToS3") // Start a new span for the S3 upload process
	defer span.End() // End the span when the function completes

	// Log the event and file details for tracing
	span.SetAttributes(
		attribute.String("event.name", eventName),
		attribute.String("file.original_name", fileHeader.Filename),
		attribute.Int("event.id", int(eventID)),
	)

	// Get the S3 endpoint from the environment (default to LocalStack if available)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		span.RecordError(fmt.Errorf("file type not allowed")) // Record error in the trace
		return "", errors.New("file type not allowed: only PNG, JPEG files are accepted")
	}

	// Trace session creation
	_, sessSpan := tracer.Start(ctx, "CreateS3Session") // Start a span for session creation
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.S3Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AwsAccessKeyID, cfg.AwsSecretAccessKey, ""),
		Endpoint:         aws.String(cfg.S3Endpoint), // Use the custom endpoint for LocalStack or AWS
		S3ForcePathStyle: aws.Bool(true),             // Force path-style URLs for LocalStack compatibility
	})
	if err != nil {
		sessSpan.RecordError(err) // Record error in trace
		sessSpan.End()            // End span
		return "", err
	}
	sessSpan.End() // End the span for session creation

	// Create a new S3 client
	s3Client := s3.New(sess)

	// Clean the event name by removing spaces or special characters for use in the S3 key
	cleanedEventName := strings.ReplaceAll(strings.ToLower(eventName), " ", "-")

	// Generate a random string for uniqueness
	randomStr, err := generateRandomString(8) // 8-byte random string (16 characters hex)
	if err != nil {
		span.RecordError(err) // Record error in the trace
		return "", err
	}

	// Generate a unique S3 key (path) using the event name, event ID, and random string
	fileName := fmt.Sprintf("events/%s-%d/%d-%s-%s%s", cleanedEventName, eventID, time.Now().Unix(), randomStr, fileHeader.Filename, ext)
	span.SetAttributes(attribute.String("file.s3_key", fileName)) // Add S3 key to the trace

	// Upload file to S3 and trace the process
	_, uploadSpan := tracer.Start(ctx, "UploadFileToS3") // Start a span for file upload
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(cfg.S3BucketNameEvents),
		Key:    aws.String(fileName), // The key includes the folder path for the event
		Body:   file,
		ACL:    aws.String("public-read"), // Make the file publicly readable
	})
	if err != nil {
		uploadSpan.RecordError(err) // Record error in the trace
		uploadSpan.End()            // End the upload span
		return "", err
	}
	uploadSpan.End() // End the span for the upload operation

	// Generate dynamic file URL based on whether you're using LocalStack or AWS
	var fileURL string
	if cfg.S3Endpoint != "" {
		// For LocalStack, generate the URL using localhost and bucket name in the path
		fileURL = fmt.Sprintf("http://localhost:4566/%s/%s", cfg.S3BucketNameEvents, fileName)
		span.SetAttributes(attribute.String("file.localstack_url", fileURL)) // Add the file URL to the trace
	} else {
		// For AWS, use the standard S3 URL format
		fileURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.S3BucketNameEvents, cfg.S3Region, fileName)
		span.SetAttributes(attribute.String("file.aws_url", fileURL)) // Add the file URL to the trace
	}

	return fileURL, nil
}

func CreateEventHandler(c *gin.Context) {
    tracer := otel.Tracer("event-service") // Get the tracer
    ctx, span := tracer.Start(c.Request.Context(), "CreateEventHandler")
    defer span.End() // End the span when the function completes

    // Extract the form fields
    event := models.Event{
        Name:        c.PostForm("name"),
        Location:    c.PostForm("location"),
        Description: c.PostForm("description"),
    }

    // Parse the date field (check for empty and handle date-only formats)
    dateStr := c.PostForm("date")
    if dateStr == "" {
        span.RecordError(fmt.Errorf("Date field cannot be empty"))
        c.JSON(http.StatusBadRequest, gin.H{"error": "Date field cannot be empty"})
        return
    }

    // Try to parse the full RFC3339 format first, fallback to date-only format
    eventDate, err := time.Parse(time.RFC3339, dateStr)
    if err != nil {
        // Fallback to parsing just the date (YYYY-MM-DD)
        eventDate, err = time.Parse("2006-01-02", dateStr)
        if err != nil {
            span.RecordError(err) // Trace the error
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Expected formats: YYYY-MM-DD or RFC3339"})
            return
        }
    }
    event.Date = eventDate

    // Extract user ID from JWT claims
    claims := jwt.ExtractClaims(c)
    userID := uint(claims[middlewares.IdentityKey].(float64))
    event.CreatorID = userID
    event.CreatedAt = time.Now()
    event.UpdatedAt = time.Now()

    // Handle the file upload
    file, fileHeader, err := c.Request.FormFile("file")
    if err != nil && err != http.ErrMissingFile {
        span.RecordError(err) // Trace the error
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file: " + err.Error()})
        return
    }

    // If a file is uploaded, process it (e.g., upload it to S3)
    if file != nil {
        cfg := c.MustGet("config").(*config.Config)
        fileURL, err := UploadFileToS3(ctx, cfg, file, fileHeader, event.ID, event.Name) // Pass context for tracing
        if err != nil {
            span.RecordError(err) // Trace the error
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3: " + err.Error()})
            return
        }
        event.FileURL = fileURL
        // Add file-related attributes to the trace
        span.SetAttributes(
            attribute.String("file.name", fileHeader.Filename),
            attribute.String("file.url", fileURL),
        )
    }

    // Trace the database operation of saving the event
    db := c.MustGet("db").(*gorm.DB)
    _, dbSpan := tracer.Start(ctx, "DB_SaveEvent") // Start a span for the DB operation
    if err := db.Create(&event).Error; err != nil {
        dbSpan.RecordError(err) // Trace the error
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
        dbSpan.End() // End the DB span
        return
    }
    dbSpan.End() // End the DB span

    // Add event-related attributes to the span
    span.SetAttributes(
        attribute.String("event.name", event.Name),
        attribute.Int("event.id", int(event.ID)),
        attribute.String("event.location", event.Location),
    )

    // Respond with the created event
    c.JSON(http.StatusCreated, event)
}

func JoinEventHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	eventIDParam := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get user ID from JWT
	claims := jwt.ExtractClaims(c)
	userID := uint(claims[middlewares.IdentityKey].(float64))

	// Check if the event exists
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		}
		return
	}

	// Check if the user has already joined the event
	var userEvent models.UserEvent
	if err := db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&userEvent).Error; err == nil {
		// Record exists, user already joined
		c.JSON(http.StatusConflict, gin.H{"error": "User has already joined this event"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check event participation"})
		return
	}

	// Create the association between user and event
	userEvent = models.UserEvent{
		UserID:  userID,
		EventID: uint(eventID),
	}

	if err := db.Create(&userEvent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the event"})
}

func UnjoinEventHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	eventIDParam := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get user ID from JWT
	claims := jwt.ExtractClaims(c)
	userIDFloat, ok := claims[middlewares.IdentityKey].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID := uint(userIDFloat)

	// Check if the event exists
	var event models.Event
	if err := db.First(&event, eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
		}
		return
	}

	// Check if the user has joined the event
	var userEvent models.UserEvent
	if err := db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&userEvent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// User hasn't joined the event
			c.JSON(http.StatusConflict, gin.H{"error": "User has not joined this event"})
			return
		} else {
			// Some other error occurred
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check event participation"})
			return
		}
	}

	// Delete the association between user and event
	if err := db.Delete(&userEvent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully left the event"})
}
