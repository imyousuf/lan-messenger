package packet

import (
	"encoding/json"
	"log"
	"time"

	"github.com/imyousuf/lan-messenger/profile"
)

func toJSON(packet interface{}) string {
	jsonBytes, err := json.Marshal(packet)
	if err == nil {
		return string(jsonBytes)
	}
	log.Fatal(err)
	return ""
}

type _BasePacket struct {
	PacketID  uint64
	SessionID string
}

func (packet _BasePacket) GetPacketID() uint64 {
	return packet.PacketID
}
func (packet _BasePacket) GetSessionID() string {
	return packet.SessionID
}
func (packet _BasePacket) ToJSON() string {
	return toJSON(packet)
}

type _PingPacket struct {
	_BasePacket
	ExpiryTime time.Time
}

func (packet _PingPacket) GetExpiryTime() time.Time {
	return packet.ExpiryTime
}

func (packet _PingPacket) ToJSON() string {
	return toJSON(packet)
}

type _RegisterPacket struct {
	_PingPacket
	DevicePreferenceIndex uint8
	ReplyTo               string
	Username              string
	DisplayName           string
	Email                 string
}

func (packet _RegisterPacket) GetReplyTo() string {
	return packet.ReplyTo
}
func (packet _RegisterPacket) GetUserProfile() profile.UserProfile {
	return profile.NewUserProfile(packet.Username, packet.DisplayName, packet.Email)
}
func (packet _RegisterPacket) GetDevicePreferenceIndex() uint8 {
	return packet.DevicePreferenceIndex
}

func (packet _RegisterPacket) ToJSON() string {
	return toJSON(packet)
}

const (
	// RegisterPacketType should be used when wanting to parse a buffer as RegisterPacket
	RegisterPacketType = iota
	// PingPacketType should be used when wanting to parse a buffer as PingPacket
	PingPacketType
	// SignOffPacketType should be used when wanting to parse a buffer as SignOffPacket
	SignOffPacketType
)

// FromJSON converts a byte array to a packet type as requested the API invoker
func FromJSON(jsonBuf []byte, packetType int) (BasePacket, error) {
	switch packetType {
	case RegisterPacketType:
		packet := _RegisterPacket{}
		err := json.Unmarshal(jsonBuf, &packet)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return packet, err
	case PingPacketType:
		packet := _PingPacket{}
		err := json.Unmarshal(jsonBuf, &packet)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return packet, err
	case SignOffPacketType:
		packet := _BasePacket{}
		err := json.Unmarshal(jsonBuf, &packet)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return packet, err
	default:
		panic("Unknown packet type!")
	}
}
