package main

import(
	"time"
	"fmt"
	"net"
	log "github.com/astaxie/beego/logs"
)

type Channel struct {
	id 		int		     // channel id 
	pktChan chan []byte  // h264 pkt chan
}

func New() *Channel {
	return &Channel{
		-1,
		make(chan []byte),
	}
}

func (c *Channel) Id() int {
	return c.id
}

func (c *Channel) Chan() chan []byte {
	return c.pktChan
}

	var channelMap = make(map[string]*Channel)
	var channels   = make([]*Channel, 16)

func LinstenAndServe() {
	log.Info("==========system init==========")
	
	//localAddress := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8008};
	vedioAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8000")
	if err != nil{
		log.Critical("net ResolveUDPAddr Error.")
	}

	log.Debug("local vedio addresses : ", vedioAddr.IP, ":", vedioAddr.Port)

	serv, err := net.ListenUDP("udp", vedioAddr)
	if err != nil {
    	log.Critical("net ListenUDP.")
	}

	defer serv.Close()

	log.Debug("rtp server started.")
 
	autoId := 0
	buffer := make([]byte, 1400)

	for {
		n, remoteAddr, err := serv.ReadFromUDP(buffer)
		log.Debug("on message")
		if err != nil {
			log.Error("failed to read UDP msg because of ", err.Error())
			return
		}

		value, ok := channelMap[remoteAddr.String()]
		if !ok {
			channelMap[remoteAddr.String()] = channels[autoId]
			value = channels[autoId]
			autoId++
		}

		value.Chan() <- buffer
		time.Sleep(1*time.Second)

		log.Debug("recv ", n, " message from ", remoteAddr.String())//, ": ", hex.EncodeToString(rtpParser.Buffer()))
		log.Debug(buffer)
	}

}

func main() {
	
	for i := 0; i < 4; i++ {
		channels[i]    = New()
		channels[i].id = i
	}

	go LinstenAndServe()

	for {
		select {
			case pkt := <- channels[0].Chan() :
				fmt.Println(pkt)
				pkt[0] = 0x55
	
			case pkt := <- channels[1].Chan() :
				fmt.Println(pkt)
		}
	}

}