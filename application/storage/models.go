package storage

import "github.com/jinzhu/gorm"

// UserModel represents the UserProfile in storage
type UserModel struct {
	gorm.Model
	Username    string `gorm:"not null;unique"`
	DisplayName string
	Email       string `gorm:"not null;unique"`
}
