package storage

import (
	"time"

	"github.com/jinzhu/gorm"
)

// UserModel represents the UserProfile in storage
type UserModel struct {
	gorm.Model
	Username    string `gorm:"not null;unique"`
	DisplayName string
	Email       string `gorm:"not null;unique"`
}

// SessionModel represents a Session of a User
type SessionModel struct {
	gorm.Model
	UserModelID             uint      // Foreign Key to UserModel
	UserModel               UserModel // Convenient Method for load related user from model directly
	SessionID               string    `gorm:"not null;unique"`
	DevicePreferenceIndex   uint8
	ExpiryTime              time.Time
	ReplyToConnectionString string
}
