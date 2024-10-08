package handlers

import (
	"time"
    "net/http"
    "strconv"

    jwt "github.com/appleboy/gin-jwt/v2"
    "github.com/gin-gonic/gin"
    "github.com/devops-360-online/go-with-me/internal/middlewares"
    "github.com/devops-360-online/go-with-me/internal/models"
    "gorm.io/gorm"
)


func ListEventsHandler(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)

    var events []models.Event
    if err := db.Find(&events).Error; err != nil {
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
    if err := db.First(&event, eventID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
        }
        return
    }

    c.JSON(http.StatusOK, event)
}

func CreateEventHandler(c *gin.Context) {
    var event models.Event
    if err := c.ShouldBindJSON(&event); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // Get user ID from JWT
    claims := jwt.ExtractClaims(c)
    userID := uint(claims[middlewares.IdentityKey].(float64))
    event.CreatorID = userID
    event.CreatedAt = time.Now()
    event.UpdatedAt = time.Now()

    db := c.MustGet("db").(*gorm.DB)
    if err := db.Create(&event).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
        return
    }
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
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve event"})
        }
        return
    }

    // Create the association between user and event
    userEvent := models.UserEvent{
        UserID:  userID,
        EventID: uint(eventID),
    }

    if err := db.Create(&userEvent).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join event"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the event"})
}
