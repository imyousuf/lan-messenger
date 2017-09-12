package application

import (
	"log"

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
	log.Println("Handled RE Broadcast: ", event.GetRegisterPacket().ToJSON())
	NewUser(event.GetRegisterPacket().GetUserProfile())
}

func (el _EventListener) HandlePingEvent(event network.PingEvent) {
	log.Println("Handled PE Broadcast: ", event.GetPingPacket().ToJSON())
}

func (el _EventListener) HandleSignOffEvent(event network.SignOffEvent) {
	log.Println("Handled SOE Broadcast: ", event.GetSignOffPacket().ToJSON())
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
