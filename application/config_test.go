package application

import (
	"fmt"
	"testing"

	"github.com/go-ini/ini"
)

const (
	port         = 34534
	iface        = "eth0"
	username     = "nicename"
	displayName  = "What to Show"
	email        = "user@email.co"
	deviceIndex  = 100
	iniConfigFmt = `[network]
	port=%d
	interface=%s
	
	[profile]
	username=%s
	displayname=%s
	email=%s
	
	[device]
	deviceindex=%d
	`
)

var (
	iniConfig = []byte(fmt.Sprintf(iniConfigFmt, port, iface, username, displayName, email,
		deviceIndex))
	mockLoadFunc = func() (*ini.File, error) {
		return ini.InsensitiveLoad(iniConfig)
	}
)

func TestGetDeviceConfig(t *testing.T) {
	loadConfiguration = mockLoadFunc
	if GetDeviceConfig() != deviceIndex {
		t.Error("Device config not returned correctly!")
	}
}

func TestGetNetworkConfig(t *testing.T) {
	loadConfiguration = mockLoadFunc
	if cPort, cIface := GetNetworkConfig(); cPort != port || cIface != iface {
		t.Error("Network config not returned correctly!")
	}
}

func TestGetUserProfile(t *testing.T) {
	loadConfiguration = mockLoadFunc
	if cUsername, cDisplayName, cEmail := GetUserProfile(); cUsername != username ||
		cDisplayName != displayName || cEmail != email {
		t.Error("User profile config not returned correctly!", cUsername, cDisplayName, cEmail,
			username, displayName, email)
	}
}
