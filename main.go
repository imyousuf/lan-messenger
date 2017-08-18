package main

import "net"
import "log"
import "github.com/imyousuf/lan-messenger/network"

func messageHandler(messageChannel chan string, completeNotificationChannel chan int) {
	for message := range messageChannel {
		log.Println(message)
	}
	completeNotificationChannel <- 1
}

func broadcastHandler(broadcastChannel chan string, completeNotificationChannel chan int) {
	for message := range broadcastChannel {
		log.Println(message)
	}
	completeNotificationChannel <- 1
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
	udpComm := network.NewUDPCommunication()
	config := network.NewConfig(30000)
	listeners, messageChannel, broadcastChannel, _ := udpComm.Listen(config)
	log.Println("Listening to: ", listeners)
	completeNotificationChannel := make(chan int)
	go broadcastHandler(broadcastChannel, completeNotificationChannel)
	go messageHandler(messageChannel, completeNotificationChannel)
	udpComm.SetupAndFireBroadcast(config)
	<-completeNotificationChannel
	<-completeNotificationChannel

}
