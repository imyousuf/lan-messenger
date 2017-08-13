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
	listeners, messageChannel, broadcastChannel, _ := network.Listen(30000)
	log.Println("Listening to: ", listeners)
	completeNotificationChannel := make(chan int)
	go broadcastHandler(broadcastChannel, completeNotificationChannel)
	go messageHandler(messageChannel, completeNotificationChannel)
	network.Broadcast()
	<-completeNotificationChannel
	<-completeNotificationChannel

}
