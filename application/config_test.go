package application

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-ini/ini"
	"github.com/imyousuf/lan-messenger/utils"
)

const (
	port            = 34534
	iface           = "eth0"
	username        = "nicename"
	displayName     = "What to Show"
	email           = "user@email.co"
	deviceIndex     = 100
	storageLocation = "/tmp/test-lamess/"
	iniConfigFmt    = `[network]
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
)

var mockLoadFunc = func() (*ini.File, error) {
	return ini.InsensitiveLoad([]byte(fmt.Sprintf(iniConfigFmt, port, iface, username, displayName,
		email, deviceIndex, storageLocation)))
}

func TestMissingMandatoryConfigs(t *testing.T) {
	loadConfiguration = func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[null]
		deviceindex1=0`))
	}
	utils.PanicableInvocation(func() {
		GetDeviceConfig()
	}, func(err interface{}) {
		t.Error("Should not have paniced for lack of device config")
	})
	utils.PanicableInvocation(func() {
		GetStorageLocation()
	}, func(err interface{}) {
		t.Error("Should not have paniced for lack of storage config")
	})
	utils.PanicableInvocation(func() {
		GetNetworkConfig()
		t.Error("Should have paniced for lack of network config")
	}, func(err interface{}) {
		// expected
	})
	utils.PanicableInvocation(func() {
		GetUserProfile()
		t.Error("Should have paniced for lack of user profile config")
	}, func(err interface{}) {
		// expected
	})
}

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
	locationInitializer = sync.Once{}
	if cUsername, cDisplayName, cEmail := GetUserProfile(); cUsername != username ||
		cDisplayName != displayName || cEmail != email {
		t.Error("User profile config not returned correctly!", cUsername, cDisplayName, cEmail,
			username, displayName, email)
	}
}

func TestGetStorageLocation(t *testing.T) {
	loadConfiguration = mockLoadFunc
	locationInitializer = sync.Once{}
	if storageLocation != GetStorageLocation() {
		t.Error("Expected storage location not returned")
	}
	if _, err := os.Stat(storageLocation); os.IsNotExist(err) {
		t.Error("Storage location does not exist")
	}
}

func TestDefaultGetDeviceConfig(t *testing.T) {
	loadConfiguration = func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[device]
		deviceindex1=0`))
	}
	if GetDeviceConfig() != 1 {
		t.Error("Default device config not returned correctly!")
	}
	loadConfiguration = func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[device]
		deviceindex=a`))
	}
	if GetDeviceConfig() != 1 {
		t.Error("Default device config not returned correctly!")
	}

}

func TestDefaultStorageLocation(t *testing.T) {
	loadConfiguration = func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[device]
		deviceindex1=0`))
	}
	locationInitializer = sync.Once{}
	defaultLocation := GetStorageLocation()
	if _, err := os.Stat(defaultLocation); os.IsNotExist(err) {
		t.Error("Default location does not exist")
	}
}

func TestDefaultNetworkConfig(t *testing.T) {
	loadConfiguration = func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[network]
		deviceindex1=0`))
	}
	port, interfaceName := GetNetworkConfig()
	if port != 30000 && interfaceName != "wlan0" {
		t.Error("Default values for network config does not match")
	}
}

func TestMissingUserProfileConfig(t *testing.T) {
	loadConfigurations := []func() (*ini.File, error){func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[profile]
		deviceindex1=0`))
	}, func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[profile]
		username=0
		displayname=0`))
	}, func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[profile]
		username=0`))
	}}
	keys := []string{"username", "email", "displayname"}
	for index := 0; index < len(loadConfigurations); index++ {
		loadConfiguration = loadConfigurations[index]
		utils.PanicableInvocation(func() {
			GetUserProfile()
		}, func(err interface{}) {
			if errStr, ok := err.(string); ok {
				if !strings.Contains(errStr, keys[index]) {
					t.Error("Unexpected error:", index, "::", errStr)
				}
			} else {
				t.Error("Unexpected error!")
			}
		})
	}
}

func TestPanicableGetStorageLocation(t *testing.T) {
	loadConfiguration = func() (*ini.File, error) {
		return ini.InsensitiveLoad([]byte(`
		[storage]
		location=/asd/0`))
	}
	locationInitializer = sync.Once{}
	utils.PanicableInvocation(func() {
		GetStorageLocation()
		t.Error("Should have panicked when creating storage location")
	}, func(r interface{}) {})
}

func GetTestConfiguration() func() (*ini.File, error) {
	return mockLoadFunc
}
