package packet

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/imyousuf/lan-messenger/profile"
	"github.com/imyousuf/lan-messenger/utils"
)

// DeviceProfileBuilder builds towards RegisterPacket
type DeviceProfileBuilder interface {
	RegisterDevice(replyToConnectionStr string, devicePreference uint8) RegisterPacketBuilder
}

// UserProfileBuilder builds towards RegisterPacket
type UserProfileBuilder interface {
	CreateUserProfile(userProfile profile.UserProfile) DeviceProfileBuilder
}

// SessionBuilder builds towards RegisterPacket
type SessionBuilder interface {
	CreateSession(age time.Duration) UserProfileBuilder
}

// SessionRenewBuilder starts building towards PingPacket by renewing session
type SessionRenewBuilder interface {
	RenewSession(age time.Duration) PingPacketBuilder
}

// RegisterPacketBuilder builds RegisterPacket for registering a peer
type RegisterPacketBuilder interface {
	BuildRegisterPacket() RegisterPacket
}

// SignOffPacketBuilder builds DeregisterPacket for logging off
type SignOffPacketBuilder interface {
	BuildSignOffPacket() SignOffPacket
}

// PingPacketBuilder builds a PingPacket for pinging presence to peers
type PingPacketBuilder interface {
	BuildPingPacket() PingPacket
}

// BuilderFactory is the central builder that allows communication to build packets
type BuilderFactory interface {
	CreateNewSession() SessionBuilder
	SignOff() SignOffPacketBuilder
	Ping() SessionRenewBuilder
}

type _Builder struct {
	// Persistent singleton fields
	sessionID        uuid.UUID
	packetSequenceID uint64
	// transient fields
	expiryTime            time.Time
	devicePreferenceIndex uint8
	replyTo               string
	userProfile           profile.UserProfile
}

func (builder *_Builder) CreateNewSession() SessionBuilder {
	atomic.AddUint64(&builder.packetSequenceID, 1)
	return builder
}
func (builder *_Builder) SignOff() SignOffPacketBuilder {
	atomic.AddUint64(&(builder.packetSequenceID), 1)
	return builder
}
func (builder *_Builder) Ping() SessionRenewBuilder {
	atomic.AddUint64(&builder.packetSequenceID, 1)
	return builder
}
func (builder _Builder) CreateSession(age time.Duration) UserProfileBuilder {
	builder.expiryTime = time.Now().Add(age)
	return builder
}
func (builder _Builder) RenewSession(age time.Duration) PingPacketBuilder {
	builder.expiryTime = time.Now().Add(age)
	return builder
}
func (builder _Builder) CreateUserProfile(userProfile profile.UserProfile) DeviceProfileBuilder {
	builder.userProfile = userProfile
	return builder
}
func (builder _Builder) RegisterDevice(
	replyToConnectionStr string, devicePreference uint8) RegisterPacketBuilder {
	if utils.IsStringBlank(replyToConnectionStr) {
		panic("No reply-to value provided")
	}
	if !utils.IsValidConnectionString(replyToConnectionStr) {
		panic("reply-to not provided in `ip-address:port` format!")
	}
	builder.replyTo = replyToConnectionStr
	builder.devicePreferenceIndex = devicePreference
	return builder
}

func (builder _Builder) BuildPingPacket() PingPacket {
	packet := &_PingPacket{}
	packet.PacketID = builder.packetSequenceID
	packet.SessionID = builder.sessionID.String()
	packet.ExpiryTime = builder.expiryTime
	return packet
}
func (builder _Builder) BuildSignOffPacket() SignOffPacket {
	packet := &_BasePacket{}
	packet.PacketID = builder.packetSequenceID
	packet.SessionID = builder.sessionID.String()
	return packet
}
func (builder _Builder) BuildRegisterPacket() RegisterPacket {
	packet := &_RegisterPacket{}
	packet.PacketID = builder.packetSequenceID
	packet.SessionID = builder.sessionID.String()
	packet.ExpiryTime = builder.expiryTime
	packet.ReplyTo = builder.replyTo
	packet.DevicePreferenceIndex = builder.devicePreferenceIndex
	packet.Username, packet.DisplayName, packet.Email = builder.userProfile.GetUsername(), builder.userProfile.GetDisplayName(), builder.userProfile.GetEmail()
	return packet
}

var builder *_Builder
var once sync.Once

func initBuilder() {
	once.Do(func() {
		builder = &_Builder{}
		sessionID, err := uuid.NewRandom()
		if err == nil {
			builder.sessionID = sessionID
		} else {
			panic("Could not generate Session ID")
		}

	})
}

// NewBuilderFactory retrieves singleton builder factory
func NewBuilderFactory() BuilderFactory {
	initBuilder()
	return builder
}

// GetCurrentSessionID returns the Session ID currently in progress by this process
func GetCurrentSessionID() string {
	initBuilder()
	return builder.sessionID.String()
}
