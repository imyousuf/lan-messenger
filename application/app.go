package application

import (
	"log"

	"github.com/imyousuf/lan-messenger/network"
)

type EventListener struct {
	completeNotificationChannel chan int
}

func (el EventListener) HandleMessageReceived(event network.MessageEvent) {
	log.Println("MESSAGE: ", event.GetMessage())
}

func (el EventListener) HandleRegisterEvent(event network.RegisterEvent) {
	log.Println("Handled RE Broadcast: ", event.GetRegisterPacket().ToJSON())
}

func (el EventListener) HandlePingEvent(event network.PingEvent) {
	log.Println("Handled PE Broadcast: ", event.GetPingPacket().ToJSON())
}

func (el EventListener) HandleSignOffEvent(event network.SignOffEvent) {
	log.Println("Handled SOE Broadcast: ", event.GetSignOffPacket().ToJSON())
}

func (el EventListener) HandleEndOfMessages() {
	el.completeNotificationChannel <- 1
}

func (el EventListener) HandleEndOfBroadcasts() {
	el.completeNotificationChannel <- 2
}

func NewEventListener(completeNotificationChannel chan int) *EventListener {
	return &EventListener{completeNotificationChannel: completeNotificationChannel}
}
