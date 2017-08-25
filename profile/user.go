package profile

// UserProfile represents the user that this instance of the application is for
type UserProfile interface {
	GetUsername() string
	GetDisplayName() string
	GetEmail() string
}

type _UserProfile struct {
	username    string
	displayName string
	email       string
}

func (profile _UserProfile) GetUsername() string {
	return profile.username
}

func (profile _UserProfile) GetDisplayName() string {
	return profile.displayName
}
func (profile _UserProfile) GetEmail() string {
	return profile.email
}

// NewUserProfile creates a UserProfile from the information passed as parameters
func NewUserProfile(username string, displayName string, email string) UserProfile {
	profile := _UserProfile{}
	profile.username, profile.displayName, profile.email = username, displayName, email
	return &profile
}
