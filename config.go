package main

import (
	"log"

	"github.com/go-ini/ini"
)

func getSection(sectionName string) *ini.Section {
	cfg, err := ini.InsensitiveLoad("lamess.cfg")
	if err != nil {
		log.Fatal(err)
	}
	section, sErr := cfg.GetSection(sectionName)
	if sErr != nil {
		log.Fatal(sErr)
	}
	return section
}

func getNetworkConfig() (int, string) {
	section := getSection("network")
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

func getDeviceConfig() uint8 {
	section := getSection("device")
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

func getUserProfile() (string, string, string) {
	section := getSection("profile")
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
