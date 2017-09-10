package application

import (
	"testing"
	"time"
)

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
