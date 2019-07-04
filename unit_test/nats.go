package main

import (
	nats "../nats.go"
	"time"
//	"fmt"
)

func main() {

// Connect to a server
nc, _ := nats.Connect(nats.DefaultURL)

b := []byte{0,1}

// Simple Async Subscriber
nc.Publish("aaa:SwitchingChan", b)

// Channel Subscriber
/*ch := make(chan *nats.Msg, 64)
sub, err := nc.ChanSubscribe("foo", ch)
msg := <- ch*/


time.Sleep(10*time.Hour)

// Close connection
nc.Close()

}
