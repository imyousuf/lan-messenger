package application

import (
	"sync"
	"testing"
	"time"

	"github.com/imyousuf/lan-messenger/application/conf"
	"github.com/imyousuf/lan-messenger/application/domains"
	s "github.com/imyousuf/lan-messenger/application/storage"
	"github.com/imyousuf/lan-messenger/application/testutils"
	"github.com/imyousuf/lan-messenger/network"
	"github.com/imyousuf/lan-messenger/packet"
	"github.com/imyousuf/lan-messenger/profile"
)

//

func TestHandleEndOfMessages(t *testing.T) {
	endOfMsgChan := make(chan int)
	eventListener := NewEventListener(endOfMsgChan)
	counter := 0
	go func() {
		<-endOfMsgChan
		counter++
	}()
	eventListener.HandleEndOfMessages()
	time.Sleep(1 * time.Millisecond)
	if counter != 1 {
		t.Error("End of message channel notification not received!")
	}
}

func TestHandleEndOfBroadcasts(t *testing.T) {
	endOfBroadcastChan := make(chan int)
	eventListener := NewEventListener(endOfBroadcastChan)
	counter := 0
	go func() {
		<-endOfBroadcastChan
		counter++
	}()
	eventListener.HandleEndOfBroadcasts()
	time.Sleep(1 * time.Millisecond)
	if counter != 1 {
		t.Error("End of message channel notification not received!")
	}
}

var globalDbSetupForAppTests = sync.Once{}

func setupCleanTestTablesForHandlerTests() {
	globalDbSetupForAppTests.Do(func() {
		conf.SetupNewConfiguration(testutils.MockLoadFunc)
		s.ReInitDBConnection()
	})
	s.GetDB().Exec(testutils.DeleteSessionModelsSQL)
	s.GetDB().Exec(testutils.DeleteUserModelsSQL)

}

type _MockRegisterEvent struct {
	regPacket         packet.RegisterPacket
	packetInitializer sync.Once
}

func (mockEvent *_MockRegisterEvent) GetName() string {
	return network.RegisterEventName
}
func (mockEvent *_MockRegisterEvent) GetEventData() []byte {
	return []byte{}
}
func (mockEvent *_MockRegisterEvent) GetEventIdentifier() (string, uint64) {
	return "SessionID", 10001
}
func (mockEvent *_MockRegisterEvent) GetRegisterPacket() packet.RegisterPacket {
	mockEvent.packetInitializer.Do(func() {
		mockEvent.regPacket = packet.NewBuilderFactory().
			CreateNewSession().CreateSession(5*time.Minute).
			CreateUserProfile(profile.NewUserProfile(conf.GetUserProfile())).
			RegisterDevice("127.0.0.1:30000", 1).BuildRegisterPacket()
	})
	return mockEvent.regPacket
}

func TestHandleRegisterEvent(t *testing.T) {
	setupCleanTestTablesForHandlerTests()
	endOfBroadcastChan := make(chan int)
	eventListener := NewEventListener(endOfBroadcastChan)
	regEvent := &_MockRegisterEvent{}
	eventListener.HandleRegisterEvent(regEvent)
	loadedSession, found := domains.GetSessionBySessionID(regEvent.GetRegisterPacket().GetSessionID())
	if !found {
		t.Error("Could not find the session just registered")
	}
	if loadedSession.IsExpired() {
		t.Error("Just registered session is already expired")
	}
}

type _MockPingEvent struct {
	pingPacket        packet.PingPacket
	packetInitializer sync.Once
}

func (mockEvent *_MockPingEvent) GetName() string {
	return network.PingEventName
}
func (mockEvent *_MockPingEvent) GetEventData() []byte {
	return []byte{}
}
func (mockEvent *_MockPingEvent) GetEventIdentifier() (string, uint64) {
	return "SessionID", 10001
}
func (mockEvent *_MockPingEvent) GetPingPacket() packet.PingPacket {
	return packet.NewBuilderFactory().Ping().RenewSession(15 * time.Minute).BuildPingPacket()
}

func TestHandlePingEvent(t *testing.T) {
	setupCleanTestTablesForHandlerTests()
	endOfBroadcastChan := make(chan int)
	eventListener := NewEventListener(endOfBroadcastChan)
	regEvent := &_MockRegisterEvent{}
	eventListener.HandleRegisterEvent(regEvent)
	pingEvent := &_MockPingEvent{}
	eventListener.HandlePingEvent(pingEvent)
	loadedSession, found := domains.GetSessionBySessionID(pingEvent.GetPingPacket().GetSessionID())
	if !found {
		t.Error("Could not find the session just registered")
	}
	newExpiryTime := pingEvent.GetPingPacket().GetExpiryTime().Truncate(time.Second)
	if !loadedSession.GetExpiryTime().Truncate(time.Second).Equal(newExpiryTime) {
		t.Error("Expiry time not updated", loadedSession.GetExpiryTime(), newExpiryTime)
	}
}

type _MockSignOffEvent struct {
}

func (mockEvent _MockSignOffEvent) GetName() string {
	return network.SignOffEventName
}
func (mockEvent _MockSignOffEvent) GetEventData() []byte {
	return []byte{}
}
func (mockEvent _MockSignOffEvent) GetEventIdentifier() (string, uint64) {
	return "SessionID", 10001
}
func (mockEvent _MockSignOffEvent) GetSignOffPacket() packet.SignOffPacket {
	return packet.NewBuilderFactory().SignOff().BuildSignOffPacket()
}

func TestHandleSignOffEvent(t *testing.T) {
	setupCleanTestTablesForHandlerTests()
	endOfBroadcastChan := make(chan int)
	eventListener := NewEventListener(endOfBroadcastChan)
	regEvent := &_MockRegisterEvent{}
	eventListener.HandleRegisterEvent(regEvent)
	signOffEvent := _MockSignOffEvent{}
	eventListener.HandleSignOffEvent(signOffEvent)
	loadedSession, found := domains.GetSessionBySessionID(signOffEvent.GetSignOffPacket().GetSessionID())
	if !found {
		t.Error("Could not find the session just registered")
	}
	if !loadedSession.IsExpired() {
		t.Error("Sign off did not expire session")
	}
}
