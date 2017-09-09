package application

import (
	"log"

	"github.com/go-ini/ini"
)

var loadConfiguration = func() (*ini.File, error) {
	return ini.InsensitiveLoad("lamess.cfg")
}

func getSection(sectionName string, loadFunc func() (*ini.File, error)) *ini.Section {
	cfg, err := loadFunc()
	if err != nil {
		log.Fatal(err)
	}
	section, sErr := cfg.GetSection(sectionName)
	if sErr != nil {
		log.Fatal(sErr)
	}
	return section
}

// GetNetworkConfig returns the port to listen to and the interface to listen to. Though we take a
// single port in, the configuration represents a sequential 3 port config - listening for incoming
// msg, listen for broadcasting message and listening for transmitted message response respectively.
func GetNetworkConfig() (int, string) {
	section := getSection("network", loadConfiguration)
	sPort, pErr := section.GetKey("port")
	port := 0
	if pErr == nil {
		port, _ = sPort.Int()
	}
	if port <= 0 {
		port = 30000
	}
	sInterfaceName, iErr := section.GetKey("interface")
	interfaceName := ""
	if iErr == nil {
		interfaceName = sInterfaceName.String()
	}
	if len(interfaceName) <= 0 {
		interfaceName = "wlan0"
	}
	return port, interfaceName
}

// GetDeviceConfig returns the index of important for the current device for the specified user
// profile
func GetDeviceConfig() uint8 {
	section := getSection("device", loadConfiguration)
	sIndex, pErr := section.GetKey("deviceindex")
	var index uint
	if pErr == nil {
		index, _ = sIndex.Uint()
	}
	if index <= 0 {
		index = 1
	}
	return uint8(index)
}

// GetUserProfile returns the username, displayname, email of the current app instance it basically
// represents the profile.UserProfile
func GetUserProfile() (string, string, string) {
	section := getSection("profile", loadConfiguration)
	keys := []string{"username", "displayname", "email"}
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		sString, err := section.GetKey(key)
		if err == nil {
			result[key] = sString.String()
		} else {
			result[key] = ""
		}
	}
	return result[keys[0]], result[keys[1]], result[keys[2]]
}
