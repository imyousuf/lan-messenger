package application

import (
	"database/sql"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/imyousuf/lan-messenger/application/storage"
	"github.com/imyousuf/lan-messenger/profile"
	"github.com/imyousuf/lan-messenger/utils"
)

const (
	deleteUserModelsSQL          = "DELETE FROM user_models"
	deleteSessionModelsSQL       = "DELETE FROM session_models"
	countUserModelsByUsernameSQL = "SELECT count(*) FROM user_models WHERE username = ?"
)

var globalDbSetupForDomainTests = sync.Once{}

func setupCleanTestTables() {
	globalDbSetupForDomainTests.Do(func() {
		loadConfiguration = GetTestConfiguration()
		locationInitializer = sync.Once{}
		dbInitializer = sync.Once{}
		successful = false
	})
	GetDB().Exec(deleteUserModelsSQL)
	GetDB().Exec(deleteSessionModelsSQL)
}

func checkRowCount(row *sql.Row, expectedRowCount uint) bool {
	var actualCount uint
	row.Scan(&actualCount)
	return actualCount == expectedRowCount
}

func assertUserProfileData(userProfile profile.UserProfile, user *User, t *testing.T) {
	if userProfile.GetDisplayName() != user.GetUserProfile().GetDisplayName() ||
		userProfile.GetEmail() != user.GetUserProfile().GetEmail() ||
		userProfile.GetUsername() != user.GetUserProfile().GetUsername() {
		t.Error("User profile data does not match")
	}
}

func TestNewUser(t *testing.T) {
	setupCleanTestTables()
	uProfile := profile.NewUserProfile(GetUserProfile())
	user := NewUser(uProfile)
	if GetDB().NewRecord(&user.userModel) || user.userProfile == nil {
		t.Error("Could not persist new user")
	}
	assertUserProfileData(uProfile, user, t)
	modelID := user.userModel.ID
	if !checkRowCount(GetDB().Raw(countUserModelsByUsernameSQL, uProfile.GetUsername()).Row(), 1) {
		t.Error("Could not match records in DB")
	}
	uProfile = profile.NewUserProfile(GetUserProfile())
	if !checkRowCount(GetDB().Raw(countUserModelsByUsernameSQL, uProfile.GetUsername()).Row(), 1) {
		t.Error("Could not match records in DB")
	}
	if modelID != user.userModel.ID {
		t.Error("ID did not match for the first created user")
	}
}

func TestGetUserByUsername(t *testing.T) {
	setupCleanTestTables()
	uProfile := profile.NewUserProfile(GetUserProfile())
	user := GetUserByUsername(uProfile.GetUsername())
	if !GetDB().NewRecord(user.userModel) {
		t.Error("Should not have found any record!")
	}
	NewUser(uProfile)
	user = GetUserByUsername(uProfile.GetUsername())
	assertUserProfileData(uProfile, user, t)
	if GetDB().NewRecord(user.userModel) {
		t.Error("Should have found record!")
	}
}

func TestUser_IsPersisted(t *testing.T) {
	setupCleanTestTables()
	uProfile := profile.NewUserProfile(GetUserProfile())
	user := GetUserByUsername(uProfile.GetUsername())
	if user.IsPersisted() {
		t.Error("Should not have found any record!")
	}
	for i := 0; i < 3; i++ {
		pUser := NewUser(uProfile)
		if !pUser.IsPersisted() {
			t.Error("Should have found record!")
		}
	}
	user = GetUserByUsername(uProfile.GetUsername())
	if !user.IsPersisted() {
		t.Error("Should have found record!")
	}
}

func TestUser_GetUserProfile(t *testing.T) {
	setupCleanTestTables()
	uProfile := profile.NewUserProfile(GetUserProfile())
	user := NewUser(uProfile)
	if user.GetUserProfile() == nil || uProfile != user.GetUserProfile() {
		t.Error("Correct instace of user profile not set")
	}
}

func cloneSession(session Session) *Session {
	session.sessionModel = &storage.SessionModel{}
	return &session
}

func TestUser_AddSession(t *testing.T) {
	setupCleanTestTables()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(0)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	// Test adding session to a user not persisted
	t.Run("Add Session to non-persisted User", func(t *testing.T) {
		nonPersistedUser := User{}
		utils.PanicableInvocation(func() {
			nonPersistedUser.AddSession(mainSession)
			t.Error("Should have paniced because of non persisted user")
		}, func(err interface{}) {
			if stateError, ok := err.(InvalidStateError); !ok {
				t.Error("Did not receive InvalidStateError")
			} else {
				log.Println(stateError)
			}
		})
	})
	//Test adding a good session to a user
	t.Run("Adding a good session", func(t *testing.T) {
		persistableSession := cloneSession(*mainSession)
		if !persistedUser.AddSession(persistableSession) {
			t.Error("AddSession should have returned true")
		}
		if !persistableSession.IsPersisted() {
			t.Error("The self session instance should have been persisted")
		}
		if persistedUser.AddSession(persistableSession) {
			t.Error("AddSession should have returned false for already added session")
		}
	})
	// Test adding session of one user to the other user
	t.Run("Add another user's session", func(t *testing.T) {
		nextProfile := profile.NewUserProfile(defaultProfile.GetUsername()+"a",
			defaultProfile.GetDisplayName(), defaultProfile.GetEmail()+"m")
		nextUser := NewUser(nextProfile)
		time.Sleep(time.Millisecond)
		persistableSecondSession := cloneSession(*secondSession)
		if !nextUser.AddSession(persistableSecondSession) {
			t.Error("Second session should have been persistable")
		}
		if !persistableSecondSession.IsPersisted() {
			t.Error("Second session instance should be self persisted")
		}
		utils.PanicableInvocation(func() {
			persistedUser.AddSession(persistableSecondSession)
			t.Error("Should have paniced because of another user's session")
		}, func(err interface{}) {
			if stateError, ok := err.(InvalidStateError); !ok {
				t.Error("Did not receive InvalidStateError")
			} else {
				log.Println(stateError)
			}
		})
	})
	// Test adding duplicate session
	t.Run("Add duplicate session", func(t *testing.T) {
		time.Sleep(time.Millisecond)
		utils.PanicableInvocation(func() {
			duplicateSecondSession := cloneSession(*secondSession)
			persistedUser.AddSession(duplicateSecondSession)
			t.Error("Should have paniced because of duplicate session id")
		}, func(err interface{}) {
			if stateError, ok := err.(SaveOperationFailedError); !ok {
				log.Println(err)
				t.Error("Did not receive SaveOperationFailedError")
			} else {
				log.Println(stateError)
			}
		})
	})
}

func TestUser_GetActiveSessions(t *testing.T) {
	setupCleanTestTables()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(0)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex+1, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	persistedUser.AddSession(mainSession)
	persistedUser.AddSession(secondSession)
	sessions := persistedUser.GetActiveSessions()
	if len(sessions) != 2 {
		t.Error("Did not return 2 sessions as expected")
	}
}

func TestUser_GetMainSession(t *testing.T) {
	setupCleanTestTables()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	session, found := persistedUser.GetMainSession()
	if found || session.IsPersisted() {
		t.Error("Should not have found any trace of persistence!")
	}
	persistedUser.AddSession(cloneSession(*mainSession))
	persistedUser.AddSession(cloneSession(*secondSession))
	session, found = persistedUser.GetMainSession()
	if !found || session.sessionID != secondSessionID {
		t.Error("Did not return correct main session")
	}
}

func TestLoadSessionFromID(t *testing.T) {
	setupCleanTestTables()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	log.Println("Created A1", persistedUser.AddSession(cloneSession(*mainSession)))
	log.Println("Created A2", persistedUser.AddSession(cloneSession(*secondSession)))
	t.Run("Load A1", func(t *testing.T) {
		a1Session := LoadSessionFromID(sessionID)
		if !a1Session.expiryTime.Equal(expiryTime) {
			t.Error("A1 Expiry time does not match", expiryTime, a1Session.expiryTime)
		}
		if a1Session.devicePreferenceIndex != devicePrefIndex {
			t.Error("A1 device index did not match")
		}
		if a1Session.sessionID != sessionID {
			t.Error("A1 session id did not match")
		}
	})
	t.Run("Load A2", func(t *testing.T) {
		a2Session := LoadSessionFromID(secondSessionID)
		if !a2Session.expiryTime.Equal(expiryTime) {
			t.Error("A2 Expiry time does not match", expiryTime, a2Session.expiryTime)
		}
		if a2Session.devicePreferenceIndex != devicePrefIndex-1 {
			t.Error("A2 device index did not match")
		}
		if a2Session.sessionID != secondSessionID {
			t.Error("A1 session id did not match")
		}
	})
	t.Run("Load A3", func(t *testing.T) {
		a3Session := LoadSessionFromID("A3")
		if a3Session.IsPersisted() {
			t.Error("A3 should have been found")
		}
	})
}

func TestSession_IsExpired(t *testing.T) {
	setupCleanTestTables()
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, time.Now().Add(4*time.Minute), replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1,
		time.Now().Add(-4*time.Minute), replyToStr)
	if mainSession.IsExpired() {
		t.Error("A1 should not have been expired")
	}
	if !secondSession.IsExpired() {
		t.Error("A2 should have been expired")
	}
}

func TestSession_GetReplyToConnectionString(t *testing.T) {
	setupCleanTestTables()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	log.Println("Created A1", persistedUser.AddSession(cloneSession(*mainSession)))
	if LoadSessionFromID(sessionID).GetReplyToConnectionString() != replyToStr {
		t.Error("Reply to string did not match")
	}
}
