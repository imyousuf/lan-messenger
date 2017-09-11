package application

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-ini/ini"
	"github.com/imyousuf/lan-messenger/utils"
)

var loadConfiguration = func() (*ini.File, error) {
	return ini.InsensitiveLoad("lamess.cfg")
}

func getSection(sectionName string, loadFunc func() (*ini.File, error)) *ini.Section {
	cfg, err := loadFunc()
	if err != nil {
		log.Panic(err)
	}
	section, sErr := cfg.GetSection(sectionName)
	if sErr != nil {
		log.Panic(sErr)
	}
	return section
}

func getOptionalSection(sectionName string, loadFunc func() (*ini.File, error)) *ini.Section {
	var section *ini.Section
	utils.PanicableInvocation(func() {
		section = getSection(sectionName, loadFunc)
	}, func(err interface{}) {})
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
	section := getOptionalSection("device", loadConfiguration)
	var index uint
	if section != nil {
		sIndex, pErr := section.GetKey("deviceindex")
		if pErr == nil {
			index, _ = sIndex.Uint()
		}
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
			panic("Missing user profile config key " + key)
		}
	}
	return result[keys[0]], result[keys[1]], result[keys[2]]
}

var locationInitializer sync.Once

func createStorageLocationIfNotExists(location string) error {
	var actualError error
	locationInitializer.Do(func() {
		log.Println("Creating directories for", location)
		if _, err := os.Stat(location); os.IsNotExist(err) {
			err := os.MkdirAll(location, os.ModePerm)
			if err != nil {
				log.Println(err)
				actualError = err
			}
		}
	})
	return actualError
}

// GetStorageLocation returns the location where application may store data
func GetStorageLocation() string {
	section := getOptionalSection("storage", loadConfiguration)
	var storageLocation string
	if section != nil {
		if location, err := section.GetKey("location"); err == nil {
			storageLocation = location.String()
		}
	}
	if utils.IsStringBlank(storageLocation) {
		defaultLocation := filepath.Join(os.TempDir(), "lamess")
		storageLocation = defaultLocation
	}
	err := createStorageLocationIfNotExists(storageLocation)
	if err != nil {
		panic(err)
	}
	return storageLocation
}
