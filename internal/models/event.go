package models

import "time"

type Event struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string    `gorm:"size:255;not null"`
    Location    string    `gorm:"size:255;not null"`
    Date        time.Time `gorm:"not null"`
    Description string    `gorm:"type:text"`
    CreatorID   uint      `gorm:"not null"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    // Associations
    Users []User `gorm:"many2many:user_events;"`
}
