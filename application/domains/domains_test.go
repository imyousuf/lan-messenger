package domains

import (
	"database/sql"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/imyousuf/lan-messenger/application/conf"
	s "github.com/imyousuf/lan-messenger/application/storage"
	"github.com/imyousuf/lan-messenger/application/testutils"
	"github.com/imyousuf/lan-messenger/packet"
	"github.com/imyousuf/lan-messenger/profile"
	"github.com/imyousuf/lan-messenger/utils"
)

const (
	deleteUserModelsSQL          = testutils.DeleteUserModelsSQL
	deleteSessionModelsSQL       = testutils.DeleteSessionModelsSQL
	countUserModelsByUsernameSQL = "SELECT count(*) FROM user_models WHERE username = ?"
)

var globalDbSetupForDomainTests = sync.Once{}

func setupCleanTestTablesForDomainTests() {
	globalDbSetupForDomainTests.Do(func() {
		conf.SetupNewConfiguration(testutils.MockLoadFunc)
		s.ReInitDBConnection()
	})
	s.GetDB().Exec(deleteUserModelsSQL)
	s.GetDB().Exec(deleteSessionModelsSQL)
}

func checkRowCount(row *sql.Row, expectedRowCount uint) bool {
	var actualCount uint
	row.Scan(&actualCount)
	return actualCount == expectedRowCount
}

// **************** User ****************

func assertUserProfileData(userProfile profile.UserProfile, user *User, t *testing.T) {
	if userProfile.GetDisplayName() != user.GetUserProfile().GetDisplayName() ||
		userProfile.GetEmail() != user.GetUserProfile().GetEmail() ||
		userProfile.GetUsername() != user.GetUserProfile().GetUsername() {
		t.Error("User profile data does not match")
	}
}

func TestNewUser(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	uProfile := profile.NewUserProfile(conf.GetUserProfile())
	user := NewUser(uProfile)
	if s.GetDB().NewRecord(&user.userModel) || user.userProfile == nil {
		t.Error("Could not persist new user")
	}
	assertUserProfileData(uProfile, user, t)
	modelID := user.userModel.ID
	if !checkRowCount(s.GetDB().Raw(countUserModelsByUsernameSQL, uProfile.GetUsername()).Row(), 1) {
		t.Error("Could not match records in DB")
	}
	uProfile = profile.NewUserProfile(conf.GetUserProfile())
	if !checkRowCount(s.GetDB().Raw(countUserModelsByUsernameSQL, uProfile.GetUsername()).Row(), 1) {
		t.Error("Could not match records in DB")
	}
	if modelID != user.userModel.ID {
		t.Error("ID did not match for the first created user")
	}
}

func TestGetUserByUsername(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	uProfile := profile.NewUserProfile(conf.GetUserProfile())
	user, found := GetUserByUsername(uProfile.GetUsername())
	if found || !s.GetDB().NewRecord(user.userModel) {
		t.Error("Should not have found any record!")
	}
	NewUser(uProfile)
	user, found = GetUserByUsername(uProfile.GetUsername())
	assertUserProfileData(uProfile, user, t)
	if !found || s.GetDB().NewRecord(user.userModel) {
		t.Error("Should have found record!")
	}
}

func TestUser_IsPersisted(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	uProfile := profile.NewUserProfile(conf.GetUserProfile())
	user, found := GetUserByUsername(uProfile.GetUsername())
	if found || user.IsPersisted() {
		t.Error("Should not have found any record!")
	}
	for i := 0; i < 3; i++ {
		pUser := NewUser(uProfile)
		if !pUser.IsPersisted() {
			t.Error("Should have found record!")
		}
	}
	user, found = GetUserByUsername(uProfile.GetUsername())
	if !found || !user.IsPersisted() {
		t.Error("Should have found record!")
	}
}

func TestUser_GetUserProfile(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	uProfile := profile.NewUserProfile(conf.GetUserProfile())
	user := NewUser(uProfile)
	if user.GetUserProfile() == nil || uProfile != user.GetUserProfile() {
		t.Error("Correct instace of user profile not set")
	}
}

// **************** Session ****************

func cloneSession(session Session) *Session {
	session.sessionModel = &s.SessionModel{}
	return &session
}

func TestUser_AddSession(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(0)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
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
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	expiredSessionID := "A3"
	devicePrefIndex := uint8(0)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex+1, expiryTime, replyToStr)
	expiredSession := NewSession(expiredSessionID, devicePrefIndex+2, expiryTime.Add(-1*time.Hour),
		replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	persistedUser.AddSession(mainSession)
	persistedUser.AddSession(secondSession)
	persistedUser.AddSession(expiredSession)
	t.Run("Only Active Sessions", func(t *testing.T) {
		sessions := persistedUser.GetActiveSessions()
		if len(sessions) != 2 {
			t.Error("Did not return 2 sessions as expected")
		}
		for _, session := range sessions {
			if session.sessionID == "A3" {
				t.Error("Expired session returned as active sessions!")
			}
		}
	})
	t.Run("All Sessions From Private Function", func(t *testing.T) {
		allSessions := getSessionsForUser(persistedUser)
		if len(allSessions) != 3 {
			t.Error("Did not return 2 sessions as expected")
		}
	})
}

func TestUser_GetMainSession(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
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

func TestGetSessionBySessionID(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	log.Println("Created A1", persistedUser.AddSession(cloneSession(*mainSession)))
	log.Println("Created A2", persistedUser.AddSession(cloneSession(*secondSession)))
	t.Run("Load A1", func(t *testing.T) {
		a1Session, _ := GetSessionBySessionID(sessionID)
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
		a2Session, found := GetSessionBySessionID(secondSessionID)
		if !found || !a2Session.expiryTime.Equal(expiryTime) {
			t.Error("A2 Expiry time does not match", expiryTime, a2Session.expiryTime)
		}
		if !found || a2Session.devicePreferenceIndex != devicePrefIndex-1 {
			t.Error("A2 device index did not match")
		}
		if !found || a2Session.sessionID != secondSessionID {
			t.Error("A1 session id did not match")
		}
	})
	t.Run("Load A3", func(t *testing.T) {
		a3Session, found := GetSessionBySessionID("A3")
		if found || a3Session.IsPersisted() {
			t.Error("A3 should have been found")
		}
	})
}

func TestSession_IsExpired(t *testing.T) {
	setupCleanTestTablesForDomainTests()
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
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	log.Println("Created A1", persistedUser.AddSession(cloneSession(*mainSession)))
	if session, found := GetSessionBySessionID(sessionID); found && session.GetReplyToConnectionString() != replyToStr {
		t.Error("Reply to string did not match")
	}
}

func TestSession_IsSelf(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	sessionID := "A1"
	secondSessionID := packet.GetCurrentSessionID()
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, time.Now().Add(4*time.Minute), replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1,
		time.Now().Add(-4*time.Minute), replyToStr)
	if mainSession.IsSelf() {
		t.Error("A1 Can not be current session!")
	}
	if !secondSession.IsSelf() {
		t.Error("Second one should have been current session", packet.GetCurrentSessionID(),
			secondSession.sessionID)
	}
}

func TestSession_Renew(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	secondSessionID := "A2"
	devicePrefIndex := uint8(1)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSession := NewSession(secondSessionID, devicePrefIndex-1, expiryTime, replyToStr)
	thirdSessionID := "A3"
	thirdSession := NewSession(thirdSessionID, devicePrefIndex-2, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	persistedUser.AddSession(thirdSession)
	t.Run("Test Invalid renewal time", func(t *testing.T) {
		firstSession := cloneSession(*mainSession)
		persistedUser.AddSession(firstSession)
		err := firstSession.Renew(time.Now().Add(-1 * time.Hour))
		if err == nil || err.Error() != InvalidRenewTimeErrorMsg {
			t.Error("Did not receive invalid time error")
		}
	})
	t.Run("Test Invalid renewal time", func(t *testing.T) {
		unsavedSession := cloneSession(*mainSession)
		err := unsavedSession.Renew(time.Now().Add(1 * time.Hour))
		if err == nil {
			t.Error("Did not receive error")
		}
	})
	t.Run("Test valid renewal time", func(t *testing.T) {
		secondSession := cloneSession(*secondSession)
		persistedUser.AddSession(secondSession)
		newExpiryTime := time.Now().Add(time.Hour)
		err := secondSession.Renew(newExpiryTime)
		if err != nil {
			t.Error("Renew was not successful!")
		}
		secondSession, found := GetSessionBySessionID(secondSessionID)
		if !found || !secondSession.expiryTime.Equal(newExpiryTime) {
			t.Error("New Expiry time value not set correctly")
		}
		if !found || !secondSession.sessionModel.ExpiryTime.Equal(newExpiryTime) {
			t.Error("New Expiry time value in model not set correctly")
		}
	})
}

func TestSession_SignOff(t *testing.T) {
	setupCleanTestTablesForDomainTests()
	expiryTime := time.Now().Add(4 * time.Minute)
	sessionID := "A1"
	devicePrefIndex := uint8(5)
	replyToStr := "127.0.0.1:4000"
	mainSession := NewSession(sessionID, devicePrefIndex, expiryTime, replyToStr)
	secondSessionID := "A2"
	secondSession := NewSession(secondSessionID, devicePrefIndex-1, expiryTime, replyToStr)
	thirdSessionID := "A3"
	thirdSession := NewSession(thirdSessionID, devicePrefIndex-2, expiryTime, replyToStr)
	defaultProfile := profile.NewUserProfile(conf.GetUserProfile())
	persistedUser := NewUser(defaultProfile)
	persistedUser.AddSession(thirdSession)
	t.Run("sign off persisted session", func(t *testing.T) {
		pMainSession := cloneSession(*mainSession)
		persistedUser.AddSession(pMainSession)
		if err := pMainSession.SignOff(); err != nil {
			t.Error("Should not have been any error signing off")
		}
		if !pMainSession.IsExpired() {
			t.Error("Session should have been expired")
		}
	})
	t.Run("sign off non-persisted session", func(t *testing.T) {
		pMainSession := cloneSession(*secondSession)
		if err := pMainSession.SignOff(); err == nil {
			t.Error("Should have been error signing off")
		}
		if pMainSession.IsExpired() {
			t.Error("Session should not have been expired")
		}
	})
}
