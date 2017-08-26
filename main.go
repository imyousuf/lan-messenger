package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/imyousuf/lan-messenger/network"
	"github.com/imyousuf/lan-messenger/profile"
)

type _EventListener struct {
	completeNotificationChannel chan int
}

func (el _EventListener) HandleMessageReceived(event network.MessageEvent) {
	log.Println(event.GetMessage())
}

func (el _EventListener) HandleRegisterEvent(event network.RegisterEvent) {
	log.Println("Handled RE Broadcast: ", event.GetRegisterPacket().ToJSON())
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

func exit(udpComm network.Communication) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		udpComm.CloseCommunication()
	}()
}

func main() {
	completeNotificationChannel := make(chan int)
	messageListener := _EventListener{completeNotificationChannel: completeNotificationChannel}
	udpComm := network.NewUDPCommunication()
	config := network.NewConfig(getNetworkConfig())
	exit(udpComm)
	udpComm.AddMessageListener(&messageListener)
	udpComm.AddBroadcastListener(&messageListener)
	udpComm.SetupCommunication(config)
	udpComm.InitCommunication(profile.NewUserProfile(getUserProfile()))
	<-completeNotificationChannel
	<-completeNotificationChannel
}
