package main

import (
	"os"
	"os/signal"

	app "github.com/imyousuf/lan-messenger/application"
	"github.com/imyousuf/lan-messenger/network"
	"github.com/imyousuf/lan-messenger/profile"
)

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
	messageListener := app.NewEventListener(completeNotificationChannel)
	udpComm := network.NewUDPCommunication()
	config := network.NewConfig(app.GetNetworkConfig())
	exit(udpComm)
	udpComm.AddMessageListener(messageListener)
	udpComm.AddBroadcastListener(messageListener)
	udpComm.SetupCommunication(config)
	udpComm.InitCommunication(profile.NewUserProfile(app.GetUserProfile()))
	<-completeNotificationChannel
	<-completeNotificationChannel
}
