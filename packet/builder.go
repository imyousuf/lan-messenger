package packet

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/imyousuf/lan-messenger/utils"
)

// DeviceProfileBuilder builds towards RegisterPacket
type DeviceProfileBuilder interface {
	RegisterDevice(replyToConnectionStr string, devicePreference uint8) RegisterPacketBuilder
}

// UserProfileBuilder builds towards RegisterPacket
type UserProfileBuilder interface {
	CreateUserProfile(username string, displayName string, email string) DeviceProfileBuilder
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
	deviceID         uuid.UUID
	packetSequenceID uint64
	// transient fields
	expiryTime            time.Time
	devicePreferenceIndex uint8
	replyTo               string
	username              string
	displayName           string
	email                 string
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
func (builder _Builder) CreateUserProfile(username string,
	displayName string, email string) DeviceProfileBuilder {
	if utils.IsStringBlank(displayName) || utils.IsStringBlank(username) ||
		utils.IsStringBlank(email) {
		panic("None of the user profile attributes are optional")
	}
	if !utils.IsStringAlphaNumericWithSpace(username) ||
		!utils.IsStringAlphaNumericWithSpace(displayName) {
		panic("Username and Display Name must be Alpha Numeric only")
	}
	if !utils.IsStringValidEmailFormat(email) {
		panic("Email is not well formatted!")
	}
	builder.displayName = displayName
	builder.username = username
	builder.email = email
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
	packet.DeviceID = builder.deviceID.String()
	packet.ReplyTo = builder.replyTo
	packet.DevicePreferenceIndex = builder.devicePreferenceIndex
	packet.Username = builder.username
	packet.Email = builder.email
	packet.DisplayName = builder.displayName
	return packet
}

var builderFactory BuilderFactory
var once sync.Once

// NewBuilderFactory retrieves singleton builder factory
func NewBuilderFactory() BuilderFactory {
	once.Do(func() {
		builder := &_Builder{}
		builderFactory = builder
		deviceID, err := uuid.NewRandom()
		if err == nil {
			builder.deviceID = deviceID
		} else {
			panic("Could not generate Device ID")
		}
		sessionID, err := uuid.NewRandom()
		if err == nil {
			builder.sessionID = sessionID
		} else {
			panic("Could not generate Session ID")
		}

	})
	return builderFactory
}

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
	DeviceID              string
	DevicePreferenceIndex uint8
	ReplyTo               string
	Username              string
	DisplayName           string
	Email                 string
}

func (packet _RegisterPacket) GetDeviceID() string {
	return packet.DeviceID
}
func (packet _RegisterPacket) GetReplyTo() string {
	return packet.ReplyTo
}
func (packet _RegisterPacket) GetUsername() string {
	return packet.Username
}
func (packet _RegisterPacket) GetDisplayName() string {
	return packet.DisplayName
}
func (packet _RegisterPacket) GetEmail() string {
	return packet.Email
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
			return nil, err
		}
		return packet, err
	case PingPacketType:
		packet := _PingPacket{}
		err := json.Unmarshal(jsonBuf, &packet)
		if err != nil {
			return nil, err
		}
		return packet, err
	case SignOffPacketType:
		packet := _BasePacket{}
		err := json.Unmarshal(jsonBuf, &packet)
		if err != nil {
			return nil, err
		}
		return packet, err
	default:
		panic("Unknown packet type!")
	}
}
