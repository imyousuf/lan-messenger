package packet

import (
	"testing"
	"time"

	"github.com/imyousuf/lan-messenger/profile"
)

func TestFromJSON(t *testing.T) {
	age, username, displayName, email, connectionStr := 5*time.Minute, "a", "a", "a@a.com", "127.0.0.1:3000"
	var deviceIndex uint8
	deviceIndex = 2
	deregPacket := NewBuilderFactory().SignOff().BuildSignOffPacket()
	var basePack BasePacket
	basePack, _ = FromJSON([]byte(deregPacket.ToJSON()), SignOffPacketType)
	checkSignOffPacket(t, basePack.(SignOffPacket))
	regPacket := NewBuilderFactory().CreateNewSession().CreateSession(age).
		CreateUserProfile(profile.NewUserProfile(username, displayName, email)).
		RegisterDevice(connectionStr, deviceIndex).
		BuildRegisterPacket()
	basePack, _ = FromJSON([]byte(regPacket.ToJSON()), RegisterPacketType)
	checkPingPacket(t, basePack.(RegisterPacket), age)
	pingPacket := NewBuilderFactory().Ping().RenewSession(age).BuildPingPacket()
	basePack, _ = FromJSON([]byte(pingPacket.ToJSON()), PingPacketType)
	checkPingPacket(t, basePack.(PingPacket), age)
}
