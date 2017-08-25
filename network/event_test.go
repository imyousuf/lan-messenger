package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/imyousuf/lan-messenger/packet"
)

func Example_convertPacketToEventData() {
	regPacket := packet.NewBuilderFactory().CreateNewSession().CreateSession(5*time.Minute).CreateUserProfile("a", "a", "a@a.co").RegisterDevice("127.0.0.1:3000", 2).BuildRegisterPacket()
	writeBuf := convertPacketToEventData(regPacket)
	fmt.Println(strings.Split(string(writeBuf), "\n")[0])
	writeBuf = convertPacketToEventData(packet.NewBuilderFactory().Ping().RenewSession(5 * time.Minute).BuildPingPacket())
	fmt.Println(strings.Split(string(writeBuf), "\n")[0])
	writeBuf = convertPacketToEventData(packet.NewBuilderFactory().SignOff().BuildSignOffPacket())
	fmt.Println(strings.Split(string(writeBuf), "\n")[0])
	// Output:
	// REGISTER
	// PING
	// SIGNOFF
}
func Example_createEventFromEventData() {
	regPacket := packet.NewBuilderFactory().CreateNewSession().CreateSession(5*time.Minute).CreateUserProfile("a", "a", "a@a.co").RegisterDevice("127.0.0.1:3000", 2).BuildRegisterPacket()
	parsedRegEvent := createEventFromEventData(convertPacketToEventData(regPacket)).(RegisterEvent)
	fmt.Println(parsedRegEvent.GetName())
	fmt.Println(regPacket.GetPacketID() == parsedRegEvent.GetRegisterPacket().GetPacketID())
	fmt.Println(regPacket.GetSessionID() == parsedRegEvent.GetRegisterPacket().GetSessionID())
	pingPacket := packet.NewBuilderFactory().Ping().RenewSession(5 * time.Minute).BuildPingPacket()
	parsedPingEvent := createEventFromEventData(convertPacketToEventData(pingPacket)).(PingEvent)
	fmt.Println(parsedPingEvent.GetName())
	fmt.Println(pingPacket.GetPacketID() == parsedPingEvent.GetPingPacket().GetPacketID())
	fmt.Println(pingPacket.GetSessionID() == parsedPingEvent.GetPingPacket().GetSessionID())
	signoffPacket := packet.NewBuilderFactory().SignOff().BuildSignOffPacket()
	parsedSignOffEvent := createEventFromEventData(convertPacketToEventData(signoffPacket)).(SignOffEvent)
	fmt.Println(parsedSignOffEvent.GetName())
	fmt.Println(signoffPacket.GetPacketID() == parsedSignOffEvent.GetSignOffPacket().GetPacketID())
	fmt.Println(signoffPacket.GetSessionID() == parsedSignOffEvent.GetSignOffPacket().GetSessionID())
	// Output:
	// REGISTER
	// true
	// true
	// PING
	// true
	// true
	// SIGNOFF
	// true
	// true
}
