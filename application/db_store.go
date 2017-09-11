package application

import (
	"path/filepath"
	"sync"

	"github.com/imyousuf/lan-messenger/application/storage"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	dbName                    = "lamess.db"
	connectionAttemptMaxTries = 5
)

var (
	db            *gorm.DB
	dbInitializer sync.Once
	successful    = false
)

// openDBConnection opens the DB connection pool and should called from application
func openDBConnection() bool {
	if !successful {
		var err error
		dbInitializer.Do(func() {
			db, err = gorm.Open("sqlite3", filepath.Join(GetStorageLocation(), dbName))
			if err == nil {
				successful = true
				db.AutoMigrate(&storage.UserModel{}, &storage.SessionModel{})
			}
		})
	}
	return successful
}

// IsDBConnectionAvailable checks if DB connection is available.
// Returns true if connection is available else false
func IsDBConnectionAvailable() bool {
	return openDBConnection()
}

// GetDB retrieve the DB connection pool
func GetDB() *gorm.DB {
	if ok := openDBConnection(); !ok {
		panic("DB Connection could not be retrieved!!")
	}
	return db
}

// CloseDB closes the current DB connection. Returns true if closed successfully.
func CloseDB() bool {
	if IsDBConnectionAvailable() {
		err := db.Close()
		return err == nil
	}
	return false
}
