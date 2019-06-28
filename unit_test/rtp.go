package main

import (
	"os"
	"net"
	"time"
	"encoding/hex"
	"../rtp"
	"../h264"
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

func testRtpSendH264() {

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8001})
	if err != nil {
		log.Critical("net ListenUDP.")
		return
	}

	defer conn.Close()

	file, err := os.Open("../des4.h264")
	if err != nil {
		panic("error open file :")
	}
	s, err := file.Stat()
	if err != nil {
		panic("error Stat file.")
	}
	l := s.Size()

	cache := make([]byte, l)

	n, err := file.Read(cache)
	if err != nil {
		panic("read file error.")
	}

	/*he, err := os.Create("./tmp.hex")
	if err != nil {
		panic("create file error.")
	}

	he.WriteString(hex.EncodeToString(cache))*/

	log.Debug("read n :", n)
	rtpPacket := rtp.NewDefaultPacketWithH264Type()
	rtpPacket.SetTimeStamp(0)
	
	parser := h264.NewParser()

	var startPos []int
	var nalus [][]byte
	j := 0 // split nalu in cache to nalus 
	
	for i := 0; i < n - 5; i++ {
		if cache[i] == 0 && cache[i+1] == 0 && cache[i+2] == 1 {
			if i > 0 && cache[i-1] == 0 {//parameter set startpos
				startPos = append(startPos, i-1)
			} else {
				startPos = append(startPos, i)
			}
			j++
			if j > 1 {
				b := cache[startPos[j-2]:startPos[j-1]]
				nalus = append(nalus, b)
			}
		}
	}
	nalus = append(nalus, cache[startPos[j-1]:])
	if len(nalus) != len(startPos) {
		panic("unknown error at split nalu in cache to nalus ")
	}
	//log.Trace("nalus : \n", nalus)

	for _, v := range nalus {
		rps := rtpPacket.ParserNaluToRtpPayload(v)

		rtpPacket.SetTimeStamp(rtpPacket.TimeStamp() + 3000)

		for _, q := range rps {

			{
				if parser.ParserToInternalSlice(q) == true {
					log.Debug(hex.EncodeToString(parser.GetInternalBuffer()))
					parser.ClearInternalBuffer()
				}
			}


			rtpPacket.SetPayload(q)		
			rtpPacket.SetSequence(rtpPacket.Sequence() + 1)

			r :=rtp.NewParser(1500)
			copy(r.Buffer(), rtpPacket.GetRtpBytes())
			r.SetPacketLength(len(rtpPacket.GetRtpBytes()))
			r.Print("tag")
			//log.Trace("rps : ",p," : \n", hex.EncodeToString(rtpPacket.GetRtpBytes()))
			conn.WriteToUDP(rtpPacket.GetRtpBytes(), &net.UDPAddr{IP: net.ParseIP("192.168.0.78"), Port: 1236})
		}
//		time.Sleep(5*time.Millisecond)
	}
}

func main(){
	//testRtpPacket()
	testRtpSendH264()
}
