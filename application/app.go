package application

import (
	"log"
	"time"

	d "github.com/imyousuf/lan-messenger/application/domains"
	"github.com/imyousuf/lan-messenger/network"
)

// EventListener encapsulates the two network interfaces into one as the application listens to both
type EventListener interface {
	network.BroadcastListener
	network.MessageListener
}

type _EventListener struct {
	completeNotificationChannel chan int
}

func (el _EventListener) HandleMessageReceived(event network.MessageEvent) {
	log.Println("MESSAGE: ", event.GetMessage())
}

func (el _EventListener) HandleRegisterEvent(event network.RegisterEvent) {
	regPacket := event.GetRegisterPacket()
	log.Println("Handled RE Broadcast: ", regPacket.ToJSON())
	user := d.NewUser(regPacket.GetUserProfile())
	session := d.NewSession(regPacket.GetSessionID(), regPacket.GetDevicePreferenceIndex(),
		regPacket.GetExpiryTime(), regPacket.GetReplyTo())
	user.AddSession(session)
}

func (el _EventListener) HandlePingEvent(event network.PingEvent) {
	pingPacket := event.GetPingPacket()
	log.Println("Handled PE Broadcast: ", pingPacket.ToJSON())
	if session, found := d.GetSessionBySessionID(pingPacket.GetSessionID()); found &&
		pingPacket.GetExpiryTime().After(time.Now()) {
		if err := session.Renew(pingPacket.GetExpiryTime()); err != nil {
			log.Println(err)
		}
	}
}

func (el _EventListener) HandleSignOffEvent(event network.SignOffEvent) {
	signoffPacket := event.GetSignOffPacket()
	log.Println("Handled SOE Broadcast: ", signoffPacket.ToJSON())
	if session, found := d.GetSessionBySessionID(signoffPacket.GetSessionID()); found {
		session.SignOff()
	}
}

func (el _EventListener) HandleEndOfMessages() {
	el.completeNotificationChannel <- 1
}

func (el _EventListener) HandleEndOfBroadcasts() {
	el.completeNotificationChannel <- 2
}

// NewEventListener creates a new instance of EventListener
func NewEventListener(completeNotificationChannel chan int) EventListener {
	return &_EventListener{completeNotificationChannel: completeNotificationChannel}
}
