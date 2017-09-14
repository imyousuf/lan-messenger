package testutils

import (
	"fmt"

	"github.com/go-ini/ini"
)

const (
	// Port - Test port for the App
	Port = 34534
	// Iface -Test interface name
	Iface = "eth0"
	// Username - Test username for the mocking
	Username = "nicename"
	// DisplayName - Test name to display in the app
	DisplayName = "What to Show"
	// Email - Test email for the test user
	Email = "user@email.co"
	// DeviceIndex - Test device index
	DeviceIndex = 100
	// StorageLocation - Test location for storing information
	StorageLocation = "/tmp/test-lamess/"
	// IniConfigFmt - Configuration string representing mock configuration
	IniConfigFmt = `[network]
	port=%d
	interface=%s
	
	[profile]
	username=%s
	displayname=%s
	email=%s
	
	[device]
	deviceindex=%d

	[storage]
	location=%s
	`
	// DeleteUserModelsSQL - The SQL for deleting all user model rows
	DeleteUserModelsSQL = "DELETE FROM user_models"
	// DeleteSessionModelsSQL - The SQL for deleting all session
	DeleteSessionModelsSQL = "DELETE FROM session_models"
)

// MockLoadFunc for a test load func
var MockLoadFunc = func() (*ini.File, error) {
	return ini.InsensitiveLoad([]byte(fmt.Sprintf(IniConfigFmt, Port, Iface, Username, DisplayName,
		Email, DeviceIndex, StorageLocation)))
}
