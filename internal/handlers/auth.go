package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/devops-360-online/go-with-me/internal/models"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm" // Import gorm
    "net/http"
)

func RegisterHandler(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }
    user.Password = string(hashedPassword)
    db := c.MustGet("db").(*gorm.DB)
    if err := db.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}
