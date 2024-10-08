package models

import (
    "errors"
    "time"

    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string    `gorm:"size:255;not null"`
    Email     string    `gorm:"uniqueIndex;size:255;not null"`
    Password  string    `gorm:"size:255;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func GetUserByEmail(db *gorm.DB, email string) (*User, error) {
    var user User
    result := db.Where("email = ?", email).First(&user)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, result.Error
    }
    return &user, result.Error
}

func CheckPasswordHash(password, hashedPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
