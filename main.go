package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/go-ini/ini"
	"github.com/imyousuf/lan-messenger/network"
)

type _MessageListener struct {
	completeNotificationChannel chan int
}

func (me _MessageListener) HandleMessageReceived(event network.MessageEvent) {
	log.Println(event.GetMessage())
}
func (me _MessageListener) HandleEndOfMessages() {
	me.completeNotificationChannel <- 1
}

type _BroadcastListener struct {
	completeNotificationChannel chan int
}

func (be _BroadcastListener) HandleBroadcastReceived(event network.BroadcastEvent) {
	log.Println(event.GetBroadcastMessage())
}
func (be _BroadcastListener) HandleEndOfBroadcasts() {
	be.completeNotificationChannel <- 2
}

func exit(udpComm network.Communication) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		udpComm.CloseCommunication()
	}()
}

func getNetworkConfig() (int, string) {
	cfg, err := ini.InsensitiveLoad("lamess.cfg")
	if err != nil {
		log.Fatal(err)
	}
	section, sErr := cfg.GetSection("network")
	if sErr != nil {
		log.Fatal(sErr)
	}
	sPort, pErr := section.GetKey("port")
	port := 0
	if pErr == nil {
		port, _ = sPort.Int()
	}
	if port <= 0 {
		port = 30000
	}
	sInterfaceName, iErr := section.GetKey("interface")
	interfaceName := ""
	if iErr == nil {
		interfaceName = sInterfaceName.String()
	}
	if len(interfaceName) <= 0 {
		interfaceName = "wlan0"
	}
	return port, interfaceName
}

func main() {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, netInterface := range interfaces {
			// Unicast addresses
			network.PrintIPv4Addresses(netInterface, true)
			// Multicast addresses
			network.PrintIPv4Addresses(netInterface, false)
		}
	} else {
		log.Fatal("Error getting interfaces", err)
	}
	completeNotificationChannel := make(chan int)
	messageListener := _MessageListener{completeNotificationChannel: completeNotificationChannel}
	broadcastListener := _BroadcastListener{completeNotificationChannel: completeNotificationChannel}
	udpComm := network.NewUDPCommunication()
	config := network.NewConfig(getNetworkConfig())
	exit(udpComm)
	udpComm.AddMessageListener(&messageListener)
	udpComm.AddBroadcastListener(&broadcastListener)
	udpComm.SetupCommunication(config)
	<-completeNotificationChannel
	<-completeNotificationChannel
}
