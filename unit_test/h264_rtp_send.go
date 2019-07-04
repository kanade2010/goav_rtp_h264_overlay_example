package main

import (
	log "github.com/astaxie/beego/logs"

	"net"
	"io/ioutil"
	"time"
	"../rtp"
)

func PackH264DataToNalus(bytes []byte) [][]byte {
	l := len(bytes)
	var startPos []int
	var nalus [][]byte
	j := 0 // split nalu in bytes to nalus 
	for i := 0; i < l - 5; i++ {
		if bytes[i] == 0 && bytes[i+1] == 0 && bytes[i+2] == 1 {
			if i > 0 && bytes[i-1] == 0 {//parameter set startpos
				startPos = append(startPos, i-1)
			} else {
				startPos = append(startPos, i)
			}
			j++
			if j > 1 {
				b := bytes[startPos[j-2]:startPos[j-1]]
				nalus = append(nalus, b)
			}
		}
	}
	nalus = append(nalus, bytes[startPos[j-1]:])
	if len(nalus) != len(startPos) {
		panic("unknown error at split nalu in bytes to nalus ")
	}

	return nalus
}

func main()  {
	raddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:10000")
	if err != nil{
		log.Critical("net ResolveUDPAddr Error.")
	}
	
	log.Debug("remote vedio addresses : ", raddr.IP, ":", raddr.Port)

	conn, err := net.DialUDP("udp4", nil, raddr)
	if err != nil {
		log.Critical("net DialUDP.")
		return 
	}

	defer conn.Close()

	data, err := ioutil.ReadFile("/home/ailumiyana/open_src/h264_to_rtp/720p.h264")
    if err != nil {
        log.Debug("File reading error", err)
        return
	}
	log.Debug("Open Success.")
	l := len(data)
    log.Debug("size of file:", l)
	
	rtpPacket  := rtp.NewDefaultPacketWithH264Type()

	nalus := PackH264DataToNalus(data)
	
	for _, v := range nalus {
		rps := rtpPacket.ParserNaluToRtpPayload(v)

		// H264 30FPS : 90000 / 30 : diff = 3000
		rtpPacket.SetTimeStamp(rtpPacket.TimeStamp() + 3000)

		for _, q := range rps {
			rtpPacket.SetSequence(rtpPacket.Sequence() + 1)
			rtpPacket.SetPayload(q)
			conn.Write(rtpPacket.GetRtpBytes())
		}

		time.Sleep(30*time.Millisecond)
	}
}
