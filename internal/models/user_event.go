package models

type UserEvent struct {
    UserID  uint `gorm:"primaryKey"`
    EventID uint `gorm:"primaryKey"`
}
