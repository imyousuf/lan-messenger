package network

import "github.com/imyousuf/lan-messenger/packet"
import "strings"

const (
	// RegisterEventName is the name of event type that represents the RegisterEvent
	RegisterEventName = "REGISTER"
	// PingEventName is the name of event type that represents the PingEvent
	PingEventName = "PING"
	// SignOffEventName is the name of event type that represents the SignOffEvent
	SignOffEventName = "SIGNOFF"
	// UnknownEventName represents all event name not explicitly supported by this network layer
	UnknownEventName = "UNKNOWN"
	newline          = "\n"
)

// Event presents an event being sent and/or received
type Event interface {
	GetName() string
	GetEventData() []byte
	GetEventIdentifier() (string, uint64)
}

// RegisterEvent represents an event with RegisterPacket
type RegisterEvent interface {
	Event
	GetRegisterPacket() packet.RegisterPacket
}

// PingEvent represents an event with PingPacket
type PingEvent interface {
	Event
	GetPingPacket() packet.PingPacket
}

// SignOffEvent represents an event with  SignOffPacket
type SignOffEvent interface {
	Event
	GetSignOffPacket() packet.SignOffPacket
}

type _Event struct {
	Name    string
	RawData []byte
}

func (event _Event) GetName() string {
	return event.Name
}

func (event _Event) GetEventData() []byte {
	return event.RawData
}

func (event _Event) GetEventIdentifier() (string, uint64) {
	return "", 0
}

type _RegisterEvent struct {
	_Event
	packet packet.RegisterPacket
}

func (event _RegisterEvent) GetRegisterPacket() packet.RegisterPacket {
	return event.packet
}

func (event _RegisterEvent) GetEventIdentifier() (string, uint64) {
	return event.packet.GetSessionID(), event.packet.GetPacketID()
}

type _PingEvent struct {
	_Event
	packet packet.PingPacket
}

func (event _PingEvent) GetPingPacket() packet.PingPacket {
	return event.packet
}

func (event _PingEvent) GetEventIdentifier() (string, uint64) {
	return event.packet.GetSessionID(), event.packet.GetPacketID()
}

type _SignOffEvent struct {
	_Event
	packet packet.SignOffPacket
}

func (event _SignOffEvent) GetSignOffPacket() packet.SignOffPacket {
	return event.packet
}

func (event _SignOffEvent) GetEventIdentifier() (string, uint64) {
	return event.packet.GetSessionID(), event.packet.GetPacketID()
}

// convertPacketToEventData converts a packet to a byte data format that can be transported
func convertPacketToEventData(pPacket packet.BasePacket) []byte {
	switch pPacket.(type) {
	case packet.RegisterPacket:
		regPacket := pPacket.(packet.RegisterPacket)
		return []byte(RegisterEventName + "\n" + regPacket.ToJSON())
	case packet.PingPacket:
		pingPacket := pPacket.(packet.PingPacket)
		return []byte(PingEventName + "\n" + pingPacket.ToJSON())
	case packet.SignOffPacket:
		signOffPacket := pPacket.(packet.SignOffPacket)
		return []byte(SignOffEventName + "\n" + signOffPacket.ToJSON())
	default:
		panic("Converting unsupported packet to data buffer")
	}
}

// createEventFromEventData helps consume data received from communication so that app can
// consume and work with the data
func createEventFromEventData(eventData []byte) Event {
	data := string(eventData)
	switch strings.Split(data, newline)[0] {
	case RegisterEventName:
		packetData := eventData[len([]byte(RegisterEventName+newline)):]
		parsedPacket, _ := packet.FromJSON(packetData, packet.RegisterPacketType)
		regEvent := _RegisterEvent{}
		regEvent.Name, regEvent.RawData, regEvent.packet = RegisterEventName, eventData, parsedPacket.(packet.RegisterPacket)
		return regEvent
	case PingEventName:
		packetData := eventData[len([]byte(PingEventName+newline)):]
		parsedPacket, _ := packet.FromJSON(packetData, packet.PingPacketType)
		pingEvent := _PingEvent{}
		pingEvent.Name, pingEvent.RawData, pingEvent.packet = PingEventName, eventData, parsedPacket.(packet.PingPacket)
		return pingEvent
	case SignOffEventName:
		packetData := eventData[len([]byte(SignOffEventName+newline)):]
		parsedPacket, _ := packet.FromJSON(packetData, packet.SignOffPacketType)
		signOffEvent := _SignOffEvent{}
		signOffEvent.Name, signOffEvent.RawData, signOffEvent.packet = SignOffEventName, eventData, parsedPacket
		return signOffEvent
	default:
		return _Event{Name: UnknownEventName, RawData: eventData}
	}
}
