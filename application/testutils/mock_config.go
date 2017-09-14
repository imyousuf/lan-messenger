package testutils

import (
	"fmt"

	"github.com/go-ini/ini"
)

const (
	Port            = 34534
	Iface           = "eth0"
	Username        = "nicename"
	DisplayName     = "What to Show"
	Email           = "user@email.co"
	DeviceIndex     = 100
	StorageLocation = "/tmp/test-lamess/"
	IniConfigFmt    = `[network]
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
	DeleteUserModelsSQL    = "DELETE FROM user_models"
	DeleteSessionModelsSQL = "DELETE FROM session_models"
)

// MockLoadFunc for a test load func
var MockLoadFunc = func() (*ini.File, error) {
	return ini.InsensitiveLoad([]byte(fmt.Sprintf(IniConfigFmt, Port, Iface, Username, DisplayName,
		Email, DeviceIndex, StorageLocation)))
}
