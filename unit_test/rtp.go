package main

import (
	"net"
	"time"
	"encoding/hex"
	"../rtp"
	log "github.com/astaxie/beego/logs"
)

//connMap := make(map[string]net.UDPConn)

func testRtpParser() {

	log.Info("==========system init==========")

	vedioAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8000")
	if err != nil{
		log.Critical("net ResolveUDPAddr Error.")
	}

	log.Debug("local audio addresses : ", vedioAddr.IP, ":", vedioAddr.Port)

	vedioConn, err := net.ListenUDP("udp", vedioAddr)
	if err != nil {
    log.Critical("net ListenUDP.")
	}

	defer vedioConn.Close()

	log.Debug("rtp serve started.")
 
	conn := vedioConn

	rtpParser := rtp.NewParser(8124)

	for {
		log.Debug("on message")

		n, remoteAddr, err := conn.ReadFromUDP(rtpParser.Buffer())
		if err != nil {
			log.Error("failed to read UDP msg because of ", err.Error())
			return
		}
		rtpParser.SetPacketLength(n);

		log.Debug("recv ", n, " message from ", remoteAddr)//, ": ", hex.EncodeToString(rtpParser.Buffer()))
		rtpParser.Print("rtp vedio");

		log.Debug("Payload() : ", hex.EncodeToString(rtpParser.Payload()))

	}
	time.Sleep(100*time.Millisecond)
}

func testRtpPacket(){

	rp := rtp.NewDefaultPacketWithH264Type()

	seq := 0xffff

	rp.SetSequence(uint16(seq))
	rp.SetTimeStamp(0x11112222)

	rp.SetPayload([]byte{0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff,0xff})

	log.Debug(hex.EncodeToString(rp.GetRtpBytes()))

}

func main(){
	testRtpPacket()
}
