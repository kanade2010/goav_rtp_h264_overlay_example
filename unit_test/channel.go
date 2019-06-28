package main

import(
	"time"
	"fmt"
	"net"
	"../cron"
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

type onFrame struct {
	chanId int
	frame  string
}

func (of *onFrame) Id() int {
	return of.chanId
}

func (of *onFrame) Frame() string {
	return of.frame
}

var flush chan bool = make(chan bool)
var updateFrame chan *onFrame = make(chan *onFrame)

flushTick := time.NewTicker(300*time.Millisecond)

func testChanSelect() {

	back_frames := []string{"chan 0", "chan 1", "chan 2", "chan 3"} 

//	back_str := "I'm Back String."


	for {
		fmt.Println(back_frames)
		select {
			case frame := <- updateFrame :
				back_frames[frame.Id()] = frame.Frame()
				fmt.Println("update frame : ", frame.Frame())

			case <- flushTick.C :
				fmt.Println("on time flush frame to filter")
				fmt.Println(back_frames)
		}
	}
}

func main() {
	cron := cron.New()

	
	go testChanSelect()

	i := "*"

	/*for j := 0; j < 30 ; j++ {
		//updateFrame <- &onFrame{0, i}
		flush <- true
		fmt.Println("update frame on chan 0 : ", i)
		time.Sleep(1*time.Second)
		i += " *"
	}

	time.Sleep(10000*time.Second)*/



	//cron.AddFunc("*/0.5 * * * * ?", func(){flush <- true})
	cron.AddFunc("*/1 * * * * ?", func(){i += " *"})

	cron.AddFunc("*/3 * * * * ?", func(){updateFrame <- &onFrame{0, i}})
	
	//cron.AddFunc("*/3 * * * * ?", func(){updateFrame <- &onFrame{1, i}})
	//cron.AddFunc("*/3 * * * * ?", func(){updateFrame <- &onFrame{2, i}})

	cron.Start()

	time.Sleep(10000*time.Second)
}