package main

import (
	log "github.com/astaxie/beego/logs"

	"./goav/avcodec"
	"./goav/avfilter"
	"./goav/avutil"

	//"io/ioutil"
	"time"
	"os"
	"unsafe"
	"net"
	"sync"
	//"encoding/hex"

	"./rtp"
	"./decoder"
	"./encoder"
	"./filter"
	"./h264"
)

var lastFrames []*avutil.Frame = make([]*avutil.Frame, 4)
var newFrames  []*avutil.Frame = make([]*avutil.Frame, 4)

var frameCacheChans []chan *avutil.Frame = make([]chan *avutil.Frame, 4)
var wait sync.WaitGroup
var mux sync.Mutex

var pts int64 = 0

var onEncFrame chan *avutil.Frame = make(chan *avutil.Frame, 3)

type decChannel struct{
	id  int // chan id
	dec *decoder.Decoder
}

func (dc *decChannel) Dec() *decoder.Decoder {
	return dc.dec
}

func (dc *decChannel) Id() int {
	return dc.id
}

func updateFrameIn30Ms(c chan *avutil.Frame, index int) {

	timeout := time.NewTimer(10*time.Millisecond)
	select {
		case frame := <- c :
			mux.Lock()
			log.Trace(index, " wb+++1 lastFrames : ", lastFrames)
			avutil.AvFrameUnref(lastFrames[index])
			lastFrames[index] = frame
			log.Trace(index, " wb+++2 lastFrames : ", lastFrames)
			mux.Unlock()
		case <- timeout.C :
			log.Trace("updateFrameIn30Ms index : ", index, "timeout.")
	}

	wait.Done()
}

func fourChannelFilter() {
	f := filter.New(filter.P720Description4)
	f.GraphDump()

	black := filter.New(filter.P720BlackColor)

	
	for i:=0; i<4; i++ {
		frameCacheChans[i] = make(chan *avutil.Frame, 6)
	}
	
	bf := black.GetFilterOutFrame()
	for bf == nil {
		time.Sleep(30*time.Millisecond)
		bf = black.GetFilterOutFrame()
	}

	for i:=0; i<4; i++ {

		f := avutil.AvFrameClone(bf)
		lastFrames[i] = f
		
		//chan to notify fix me
	}
	avutil.AvFrameFree(bf)
	
	log.Debug("-------fourChannelFilter start--------")

	for {
		for i:=0; i<4; i++ {
			wait.Add(1)
			go updateFrameIn30Ms(frameCacheChans[i], i)
		}
	
		wait.Wait()

		pts++

		for i:=0; i<4; i++ {
			lastFrames[i].SetPts(pts)
			ret := avfilter.AvBuffersrcWriteFrame(f.Ins()[i], (*avfilter.Frame)(unsafe.Pointer(lastFrames[i])))
			if ret < 0 {
				log.Error("AvBuffersrcWriteFrame error :", avutil.ErrorFromCode(ret))
			}
		}

		out := f.GetFilterOutFrame()
		if out != nil {
			log.Debug("--------------------filter out !")
			frame := avutil.AvFrameClone(out)
			avutil.AvFrameUnref(out)

			onEncFrame <- frame
		}
	}
}

// in rtp payload
// out h264 slice pkt
func parserH264Pkt(rp []byte) []byte {

	h264Praser := h264.NewParser()
	h264Praser.FillNaluHead(rp[0])

	var h264Buffer []byte
	if h264Praser.NaluType() == h264.NalueTypeFuA {
		h264Praser.FillShadUnitA([2]byte{rp[0], rp[1]})
		if h264Praser.ShardA().IsStart() {
			//log.Trace("-----is fuA start------")
			h264Buffer = append(h264Buffer, h264.StartCode[0:]...)
			h264Buffer = append(h264Buffer, h264Praser.ShardA().NaluHeader())
			h264Buffer = append(h264Buffer, rp[2:]...)
		} else if h264Praser.ShardA().IsEnd() {
			//log.Trace("-----is fuA end--------")
			h264Buffer = append(h264Buffer, rp[2:]...)
			//log.Trace("len: ", len(h264Buffer), ":\n", hex.EncodeToString(h264Buffer))
		} else {
			//log.Trace("-----is fuA slice------")

			h264Buffer = append(h264Buffer, rp[2:]...)
		}
	} else {
		//log.Trace("nalu : ", h264Praser.NaluType())			

		h264Buffer = append(h264Buffer, h264.StartCode[0:]...)
		h264Buffer = append(h264Buffer, rp[0:]...)
		//log.Trace("len: ", len(h264Buffer), ":\n", hex.EncodeToString(h264Buffer))
		
	}

	return h264Buffer
}

func startServe(conn *net.UDPConn) {

	rParser := rtp.NewDefaultParser()
	//dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	chan_id := 0
	decChannels := make(map[string]*decChannel)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(rParser.Buffer())
		if err != nil {
			log.Error("failed to read UDP msg because of ", err.Error())
			return
		}
		rParser.SetPacketLength(n)

		_, ok := decChannels[remoteAddr.String()]
		if !ok && chan_id < 4 {
			decChannels[remoteAddr.String()] = &decChannel{
				chan_id,
				decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264)),
			}
			chan_id++
		}

		log.Debug("recv ", n, " message from ", remoteAddr)//, ": ", hex.EncodeToString(rtpParser.Buffer()))

		h264Slice := parserH264Pkt(rParser.Payload())

		remain := len(h264Slice)
		nRead  := 0
		for remain > 0 {
			nRead = decChannels[remoteAddr.String()].Dec().ParserPacket(h264Slice[nRead:remain + nRead], remain)
			remain = remain - nRead

			//log.Trace("-----", hex.EncodeToString([]byte{'-','-','-',}), dec.Packet().GetPacketSize())
			//log.Trace("decode parsered ", dec.Packet().GetPacketSize(), " Bytes")
			
			if decChannels[remoteAddr.String()].Dec().Packet().GetPacketSize() > 0 {
				if decChannels[remoteAddr.String()].Dec().GenerateFrame() == 0 {
					f := avutil.AvFrameClone(decChannels[remoteAddr.String()].Dec().Frame())
					//log.Debug("wb+++++1 :", f)
					//log.Debug("wb+++++2 :", dec.Frame())
					avutil.AvFrameUnref(decChannels[remoteAddr.String()].Dec().Frame())
					
					//pictType := avutil.AvGetPictureTypeChar(f.PictureType())
					//log.Trace("dec.FrameNumber() :", dec.Context().FrameNumber(), " AvGetPictureTypeChar:", avutil.AvGetPictureTypeChar(dec.Frame().PictureType()))
					//log.Trace(dec.Frame().)
					//dec.Frame().SetPts()
					//log.Trace("----- pictType : ", pictType)
					//if pictType == "I" || pictType == "P" {
						select {
							case frameCacheChans[decChannels[remoteAddr.String()].Id()] <- f :
							
							default :
								log.Debug("frame Cache Chans full discard this frame")
								avutil.AvFrameFree(f)
						}
					//}
				}
			}
		}
	}

}

func PackH264FrameToNalus(bytes []byte) [][]byte {
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

func rtpSend(conn *net.UDPConn){
	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1280, 720,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)

	// use debug
	file, err := os.Create("./des4.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	rtpPacket  := rtp.NewDefaultPacketWithH264Type()		


	for {
		select {
		case frame := <- onEncFrame :
			log.Debug("on frame :", frame)
			if enc.GeneratePacket(frame) == 0 {
				bytes := enc.ToBytes()
				file.Write(bytes)
		
				nalus := PackH264FrameToNalus(bytes)
		
				for _, v := range nalus {
					rps := rtpPacket.ParserNaluToRtpPayload(v)
			
					// H264 30FPS : 90000 / 30 : diff = 3000
					rtpPacket.SetTimeStamp(rtpPacket.TimeStamp() + 3000)
		
					for _, q := range rps {
						rtpPacket.SetSequence(rtpPacket.Sequence() + 1)
						rtpPacket.SetPayload(q)
			
						// debug
						/*{
							rp := rtp.NewDefaultParser()
							copy(rp.Buffer(), rtpPacket.GetRtpBytes())
							rp.SetPacketLength(len(rtpPacket.GetRtpBytes()))
							rp.Print("tag")
						}*/

						conn.WriteToUDP(rtpPacket.GetRtpBytes(), &net.UDPAddr{IP: net.ParseIP("192.168.0.78"), Port: 1236})
					}
				}
			}
			avutil.AvFrameFree(frame)
		}
	}

}

func main() {
	
	vedioAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8000")
	if err != nil{
		log.Critical("net ResolveUDPAddr Error.")
	}

	log.Debug("local vedio addresses : ", vedioAddr.IP, ":", vedioAddr.Port)

	conn, err := net.ListenUDP("udp", vedioAddr)
	if err != nil {
		log.Critical("net ListenUDP.")
	}

	defer conn.Close()

	go fourChannelFilter()

	go rtpSend(conn)

	go startServe(conn)

	for {
		time.Sleep(30*time.Hour)
	}

}