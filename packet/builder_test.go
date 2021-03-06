package packet

import (
	"fmt"
	"testing"
	"time"

	"github.com/imyousuf/lan-messenger/profile"
	"github.com/imyousuf/lan-messenger/utils"
)

func ExampleNewBuilderFactory() {
	regPacket := NewBuilderFactory().CreateNewSession().CreateSession(5*time.Minute).
		CreateUserProfile(profile.NewUserProfile("a", "a", "a@a.com")).
		RegisterDevice("127.0.0.1:3000", 2).BuildRegisterPacket()
	buffer := regPacket.GetPacketID() - 1
	fmt.Println(regPacket.GetPacketID() - buffer)
	signal := make(chan int)
	const totalAsyncs = 3
	completed := 0
	for i := 0; i < totalAsyncs; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond)
			pingPacket := NewBuilderFactory().Ping().RenewSession(5 * time.Minute).BuildPingPacket()
			fmt.Println(pingPacket.GetPacketID() - buffer)
			signal <- 1
		}()
	}
	for inSignal := range signal {
		completed += inSignal
		if completed >= totalAsyncs {
			close(signal)
		}
	}
	logoffPacket := NewBuilderFactory().SignOff().BuildSignOffPacket()
	fmt.Println(logoffPacket.GetPacketID() - buffer)
	// Unordered output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

func unknownPacketBuild(packetBuilder func(), panicHandler func()) {
	defer panicHandler()
	packetBuilder()
}

func ExampleNewBuilderFactory_withPanic() {
	panicHandler := func(r interface{}) {
		fmt.Println("As expected panic handled:", r)
	}
	uProfile := profile.NewUserProfile("a", "a", "a@a.co")
	utils.PanicableInvocation(func() {
		NewBuilderFactory().CreateNewSession().CreateSession(5*time.Minute).
			CreateUserProfile(uProfile).RegisterDevice("", 1)
	}, panicHandler)
	utils.PanicableInvocation(func() {
		NewBuilderFactory().CreateNewSession().CreateSession(5*time.Minute).
			CreateUserProfile(uProfile).RegisterDevice("aasd:123", 1)
	}, panicHandler)
	NewBuilderFactory().CreateNewSession().CreateSession(5*time.Minute).
		CreateUserProfile(uProfile).RegisterDevice("127.0.0.1:123", 1).BuildRegisterPacket()
	// Output:
	// As expected panic handled: No reply-to value provided
	// As expected panic handled: reply-to not provided in `ip-address:port` format!
}

func TestRegisterPacketCreation(t *testing.T) {
	age, username, displayName, email, connectionStr := 5*time.Minute, "a", "a", "a@a.com", "127.0.0.1:3000"
	var deviceIndex uint8
	deviceIndex = 2
	regPacket := NewBuilderFactory().CreateNewSession().CreateSession(age).
		CreateUserProfile(profile.NewUserProfile(username, displayName, email)).RegisterDevice(connectionStr, deviceIndex).
		BuildRegisterPacket()
	if regPacket == nil {
		t.Error("Registration packet is nil!")
	}
	if regPacket.GetPacketID() <= 0 {
		t.Error("Not a valid packet ID")
	}
	if utils.IsStringBlank(regPacket.GetSessionID()) {
		t.Error("Blank String not expected for Device ID or Session ID")
	}
	expiryTime := regPacket.GetExpiryTime()
	if expiryTime.Before(time.Now()) || expiryTime.After(time.Now().Add(age)) {
		t.Error("Invalid expiry time")
	}
	if username != regPacket.GetUserProfile().GetUsername() ||
		displayName != regPacket.GetUserProfile().GetDisplayName() ||
		email != regPacket.GetUserProfile().GetEmail() {
		t.Error("User profile did not match")
	}
	if connectionStr != regPacket.GetReplyTo() || deviceIndex != regPacket.GetDevicePreferenceIndex() {
		t.Error("Device configuration did not match")
	}
}

func checkSignOffPacket(t *testing.T, deregPacket SignOffPacket) {
	if deregPacket == nil {
		t.Error("Deregistration packet is nil!")
	}
	if deregPacket.GetPacketID() <= 0 {
		t.Error("Not a valid packet ID")
	}
	if utils.IsStringBlank(deregPacket.GetSessionID()) {
		t.Error("Blank String not expected for Session ID")
	}
}

func TestDeregisterPacketCreation(t *testing.T) {
	deregPacket := NewBuilderFactory().SignOff().BuildSignOffPacket()
	checkSignOffPacket(t, deregPacket)
}

func checkPingPacket(t *testing.T, pingPacket PingPacket, age time.Duration) {
	if pingPacket == nil {
		t.Error("Ping packet is nil!")
	}
	if pingPacket.GetPacketID() <= 0 {
		t.Error("Not a valid packet ID")
	}
	if utils.IsStringBlank(pingPacket.GetSessionID()) {
		t.Error("Blank String not expected for Session ID")
	}
	expiryTime := pingPacket.GetExpiryTime()
	if expiryTime.Before(time.Now()) || expiryTime.After(time.Now().Add(age)) {
		t.Error("Invalid expiry time")
	}
}

func TestPingPacketCreation(t *testing.T) {
	age := 5 * time.Minute
	pingPacket := NewBuilderFactory().Ping().RenewSession(age).BuildPingPacket()
	checkPingPacket(t, pingPacket, age)
}
