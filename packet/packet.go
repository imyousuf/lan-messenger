package packet

import "time"

// BasePacket represents the packet information required for all messages
type BasePacket interface {
	GetPacketID() uint64
	GetSessionID() string
	ToJSON() string
}

// PingPacket represents the packet used to notify of a peers presence
type PingPacket interface {
	BasePacket
	GetExpiryTime() time.Time
}

// RegisterPacket represents information broadcasted when a device comes up live
// and when it pings its peers notifying its presence
type RegisterPacket interface {
	PingPacket
	GetReplyTo() string
	GetUsername() string
	GetDisplayName() string
	GetEmail() string
	GetDevicePreferenceIndex() uint8
}

// SignOffPacket represents the packet sent when a device exits
type SignOffPacket interface {
	BasePacket
}
