package main

import (
	log "github.com/astaxie/beego/logs"

	"./goav/avcodec"
	"./goav/avfilter"
	"./goav/avutil"

	"io/ioutil"
	"time"
	"os"
	"unsafe"
	"net"
	"encoding/hex"

	"./rtp"
	"./decoder"
	"./encoder"
	"./filter"
	"./h264"

)

func testSaveDescription4() {
	filters := filter.New(filter.P720Description4)
	filters.GraphDump()

	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1280, 720,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./des4.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	go func() {
		filter0 := filter.New(filter.P720BlackColor)

		time.Sleep(1*time.Second)
		
		for {
			//time.Sleep(30*time.Millisecond)
			frame0 := filter0.GetFilterOutFrame()

			if frame0 != nil { 
				log.Trace(len(filters.Ins()), " 0")
				avfilter.AvBuffersrcWriteFrame(filters.Ins()[0], (*avfilter.Frame)(unsafe.Pointer(frame0)))
				avfilter.AvBuffersrcWriteFrame(filters.Ins()[1], (*avfilter.Frame)(unsafe.Pointer(frame0)))
				avfilter.AvBuffersrcWriteFrame(filters.Ins()[2], (*avfilter.Frame)(unsafe.Pointer(frame0)))
				avfilter.AvBuffersrcWriteFrame(filters.Ins()[3], (*avfilter.Frame)(unsafe.Pointer(frame0)))
			}

			log.Debug("----Get-----")
			frame := filters.GetFilterOutFrame()
			log.Debug("----End-----")
			if frame != nil {
				if enc.GeneratePacket(frame) == 0 {
					cache := enc.ToBytes()
					file.Write(cache)
				}
			}

			avutil.AvFrameUnref(frame0)
		}

	}()

	time.Sleep(500*time.Second)
}

func testSaveWhiteBackground() {
	filters := filter.New(filter.WhiteColor)
	filters.GraphDump()

	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1920, 1080,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./white.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	for {

		frame := filters.GetFilterOutFrame()
		if frame != nil {
			if enc.GeneratePacket(frame) == 0 {
				cache := enc.ToBytes()
				file.Write(cache)
			}
		}
		time.Sleep(10*time.Millisecond)
	}
}

func testSaveBlackBackground() {
	filters := filter.New(filter.BlackColor)
	filters.GraphDump()

	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1920, 1080,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./black.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	for {

		frame := filters.GetFilterOutFrame()
		if frame != nil {
			if enc.GeneratePacket(frame) == 0 {
				cache := enc.ToBytes()
				file.Write(cache)
			}
		}
		time.Sleep(10*time.Millisecond)
	}
}

func testDecodeEncode() {
	data, err := ioutil.ReadFile("../720p/720p.h264")
    if err != nil {
        log.Debug("File reading error", err)
        return
	}
	log.Debug("Open Success.")
	l := len(data)
    log.Debug("size of file:", l)

	dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	
	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1280, 720,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	
	b := make([]byte, 4096 + 64)
	
	file, err := os.Create("./out.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	sum := 0
	for sum < l {
		remain := 4096
		for remain > 0 {
			copy(b, data[sum:sum + 4096])
			n := dec.ParserPacket(b, remain)
			log.Debug("parser ", n, "bytes")
			
			sum     = sum + n
			remain  = remain - n;

			log.Trace("--------", dec.Packet().GetPacketSize())
			if dec.Packet().GetPacketSize() > 0 {
				//log.Debug(*dec.Packet())
				
				if dec.GenerateFrame() == 0 {
					/*data0 := avutil.Data(dec.Frame())[0]
					buf := make([]byte, dec.Packet().GetPacketSize())
					startPos := uintptr(unsafe.Pointer(data0))
					for i := 0; i < dec.Packet().GetPacketSize(); i++ {
						element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
						buf[i] = element
					}*/

					if enc.GeneratePacket(dec.Frame()) == 0 {
						cache := enc.ToBytes()
						file.Write(cache)
					}
					
					avutil.AvFrameUnref(dec.Frame())
				}
			}
		}
	}
}


func testUdpH264Des4() {
	filters := filter.New(filter.P720Description4)
	filters.GraphDump()

	dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1280, 720,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./des4.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	go func() {
		black := filter.New(filter.P720BlackColor)

		time.Sleep(1*time.Second)
		
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

		buffer := make([]byte, 4096 + 64)
		
		for {	
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Error("failed to read UDP msg because of ", err.Error())
				return
			}
			log.Debug("recv ", n, " message from ", remoteAddr)
			//log.Debug("------h264 Data: \n", buffer[0:remain])

			remain := n
			nRead  := 0
			for remain > 0 {
				log.Trace("nRead : ", nRead, "  remain : ", remain)
				nRead = dec.ParserPacket(buffer[nRead:remain + nRead], remain)
				remain = remain - nRead

				log.Trace("decode parsered ", dec.Packet().GetPacketSize(), " Bytes")
				if dec.Packet().GetPacketSize() > 0 {
					if dec.GenerateFrame() == 0 {

						blackFrame := black.GetFilterOutFrame()
						log.Trace("dec.FrameNumber() :", dec.Context().FrameNumber(), " AvGetPictureTypeChar:", avutil.AvGetPictureTypeChar(dec.Frame().PictureType()))
						//log.Trace(dec.Frame().)
						//dec.Frame().SetPts()
						log.Trace("PTS : ", avutil.GetBestEffortTimestamp(dec.Frame()), "/", avutil.GetBestEffortTimestamp(blackFrame))
						log.Trace("Pkt_PTS : ", avutil.GetPktPts(dec.Frame()), "/", avutil.GetPktPts(blackFrame))
	
						pictType := avutil.AvGetPictureTypeChar(dec.Frame().PictureType())
						if pictType == "B" || pictType == "I" || pictType == "P" {
							dec.IncrIbpFrameCount()
							dec.Frame().SetPts(dec.IbpFrameCount())
						}
						dec.Frame().SetPts(int64(dec.Context().FrameNumber()))

						if blackFrame != nil { 
							log.Trace(len(filters.Ins()), " 0")
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[0], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[1], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[3], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
						}

						log.Trace(len(filters.Ins()), " 2")
						avfilter.AvBuffersrcAddFrame(filters.Ins()[2], (*avfilter.Frame)(unsafe.Pointer(dec.Frame())))

						log.Debug("----Get-----")
						frame := filters.GetFilterOutFrame()
						log.Debug("----End-----")
						if frame != nil {
							if enc.GeneratePacket(frame) == 0 {
								cache := enc.ToBytes()
								file.Write(cache)
							}
						}
						avutil.AvFrameUnref(blackFrame)
						avutil.AvFrameUnref(dec.Frame())
					}
				}
			}
		}
	}()

	for {
		time.Sleep(500*time.Second)
	}

}


func testRtpDEncodeDes4() {
	filters := filter.New(filter.P720Description4)
	filters.GraphDump()

	dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1280, 720,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./des4.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	go func() {
		black := filter.New(filter.P720BlackColor)

		time.Sleep(1*time.Second)
		
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

		rtpParser  := rtp.NewParser(1500)
		h264Praser := h264.NewParser()

		for {	
			n, remoteAddr, err := conn.ReadFromUDP(rtpParser.Buffer())
			if err != nil {
				log.Error("failed to read UDP msg because of ", err.Error())
				return
			}

			rtpParser.SetPacketLength(n);
			//rtpParser.Print("rtp vedio");
			//time.Sleep(1*time.Second)
			rtpParser.Payload()
			log.Debug("recv ", n, " message from ", remoteAddr)//, ": ", hex.EncodeToString(rtpParser.Buffer()))
			
			//log.Debug("------h264 Data: \n", rtpParser.Payload())

			h264Praser.FillNaluHead(rtpParser.Payload()[0])

			var h264Buffer []byte
			if h264Praser.NaluType() == h264.NalueTypeFuA {
				h264Praser.FillShadUnitA([2]byte{rtpParser.Payload()[0], rtpParser.Payload()[1]})
				if h264Praser.ShardA().IsStart() {
					log.Trace("-----is fuA start------")
					h264Buffer = append(h264Buffer, h264.StartCode[0:]...)
					h264Buffer = append(h264Buffer, h264Praser.ShardA().NaluHeader())
					h264Buffer = append(h264Buffer, rtpParser.Payload()[2:]...)
				} else if h264Praser.ShardA().IsEnd() {
					log.Trace("-----is fuA end--------")
					h264Buffer = append(h264Buffer, rtpParser.Payload()[2:]...)
					//log.Trace("len: ", len(h264Buffer), ":\n", hex.EncodeToString(h264Buffer))
				} else {
					log.Trace("-----is fuA slice------")

					h264Buffer = append(h264Buffer, rtpParser.Payload()[2:]...)
				}
			} else {
				log.Trace("nalu : ", h264Praser.NaluType())			

				h264Buffer = append(h264Buffer, h264.StartCode[0:]...)
				h264Buffer = append(h264Buffer, rtpParser.Payload()[0:]...)
				//log.Trace("len: ", len(h264Buffer), ":\n", hex.EncodeToString(h264Buffer))
				
			}

			remain := len(h264Buffer)
			nRead  := 0
			for remain > 0 {
				nRead = dec.ParserPacket(h264Buffer[nRead:remain + nRead], remain)
				remain = remain - nRead

				log.Trace(hex.EncodeToString([]byte{'-','-','-',}), dec.Packet().GetPacketSize())
				log.Trace("decode parsered ", dec.Packet().GetPacketSize(), " Bytes")
				if dec.Packet().GetPacketSize() > 0 {
					if dec.GenerateFrame() == 0 {

						blackFrame := black.GetFilterOutFrame()
						log.Trace("dec.FrameNumber() :", dec.Context().FrameNumber(), " AvGetPictureTypeChar:", avutil.AvGetPictureTypeChar(dec.Frame().PictureType()))
						//log.Trace(dec.Frame().)
						//dec.Frame().SetPts()
						log.Trace("PTS : ", avutil.GetBestEffortTimestamp(dec.Frame()), "/", avutil.GetBestEffortTimestamp(blackFrame))
						log.Trace("Pkt_PTS : ", avutil.GetPktPts(dec.Frame()), "/", avutil.GetPktPts(blackFrame))
	
						pictType := avutil.AvGetPictureTypeChar(dec.Frame().PictureType())
						if pictType == "B" || pictType == "I" || pictType == "P" {
							dec.IncrIbpFrameCount()
							dec.Frame().SetPts(dec.IbpFrameCount())
						}
						dec.Frame().SetPts(int64(dec.Context().FrameNumber()))

						if blackFrame != nil { 
							log.Trace(len(filters.Ins()), " 0")
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[0], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[1], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[3], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
						}

						log.Trace(len(filters.Ins()), " 2")
						avfilter.AvBuffersrcAddFrame(filters.Ins()[2], (*avfilter.Frame)(unsafe.Pointer(dec.Frame())))

						log.Debug("----Get-----")
						frame := filters.GetFilterOutFrame()
						log.Debug("----End-----")
						if frame != nil {
							if enc.GeneratePacket(frame) == 0 {
								cache := enc.ToBytes()
								file.Write(cache)
								
								//log.Trace("wb++++\n", hex.EncodeToString(cache))

								/*l := len(cache)
								maxRpSize := 1500 - 64
								if l > maxRpSize {

								} else {
									conn.Send()
								}*/
							}
						}
						avutil.AvFrameUnref(blackFrame)
						avutil.AvFrameUnref(dec.Frame())
					}
				}
			}
		}
	}()

	for {
		time.Sleep(500*time.Second)
	}
}

func testRtpTransportDes4() {
	filters := filter.New(filter.P720Description4)
	filters.GraphDump()

	dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1280, 720,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./des4.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	go func() {
		black := filter.New(filter.P720BlackColor)

		time.Sleep(1*time.Second)
		
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

		rtpParser  := rtp.NewParser(1500)
		h264Praser := h264.NewParser()

		rtpPacket  := rtp.NewDefaultPacketWithH264Type()		

		for {	
			n, remoteAddr, err := conn.ReadFromUDP(rtpParser.Buffer())
			if err != nil {
				log.Error("failed to read UDP msg because of ", err.Error())
				return
			}

			rtpParser.SetPacketLength(n);
			//rtpParser.Print("rtp vedio");
			//time.Sleep(1*time.Second)
			rtpParser.Payload()
			log.Debug("recv ", n, " message from ", remoteAddr)//, ": ", hex.EncodeToString(rtpParser.Buffer()))
			
			//log.Debug("------h264 Data: \n", rtpParser.Payload())

			h264Praser.FillNaluHead(rtpParser.Payload()[0])

			var h264Buffer []byte
			if h264Praser.NaluType() == h264.NalueTypeFuA {
				h264Praser.FillShadUnitA([2]byte{rtpParser.Payload()[0], rtpParser.Payload()[1]})
				if h264Praser.ShardA().IsStart() {
					log.Trace("-----is fuA start------")
					if h264Praser.ShardA().NaluType() == h264.NalueTypeSps ||
					h264Praser.ShardA().NaluType() == h264.NalueTypePps {
						h264Buffer = append(h264Buffer, h264.ParameterSetStartCode[0:]...)
					} else {
						h264Buffer = append(h264Buffer, h264.StartCode[0:]...)
					}

					h264Buffer = append(h264Buffer, h264Praser.ShardA().NaluHeader())
					h264Buffer = append(h264Buffer, rtpParser.Payload()[2:]...)
				} else if h264Praser.ShardA().IsEnd() {
					log.Trace("-----is fuA end--------")
					h264Buffer = append(h264Buffer, rtpParser.Payload()[2:]...)
					//log.Trace("len: ", len(h264Buffer), ":\n", hex.EncodeToString(h264Buffer))
				} else {
					log.Trace("-----is fuA slice------")

					h264Buffer = append(h264Buffer, rtpParser.Payload()[2:]...)
				}
			} else {
				log.Trace("nalu : ", h264Praser.NaluType())			
				
				if h264Praser.ShardA().NaluType() == h264.NalueTypeSps ||
				h264Praser.ShardA().NaluType() == h264.NalueTypePps {
					h264Buffer = append(h264Buffer, h264.ParameterSetStartCode[0:]...)
				} else {
					h264Buffer = append(h264Buffer, h264.StartCode[0:]...)
				}

				h264Buffer = append(h264Buffer, rtpParser.Payload()[0:]...)
				//log.Trace("len: ", len(h264Buffer), ":\n", hex.EncodeToString(h264Buffer))
				
			}

			remain := len(h264Buffer)
			nRead  := 0
			for remain > 0 {
				nRead = dec.ParserPacket(h264Buffer[nRead:remain + nRead], remain)
				remain = remain - nRead

				log.Trace(hex.EncodeToString([]byte{'-','-','-',}), dec.Packet().GetPacketSize())
				log.Trace("decode parsered ", dec.Packet().GetPacketSize(), " Bytes")
				if dec.Packet().GetPacketSize() > 0 {
					if dec.GenerateFrame() == 0 {

						blackFrame := black.GetFilterOutFrame()
						log.Trace("dec.FrameNumber() :", dec.Context().FrameNumber(), " AvGetPictureTypeChar:", avutil.AvGetPictureTypeChar(dec.Frame().PictureType()))
						//log.Trace(dec.Frame().)
						//dec.Frame().SetPts()
						log.Trace("PTS : ", avutil.GetBestEffortTimestamp(dec.Frame()), "/", avutil.GetBestEffortTimestamp(blackFrame))
						log.Trace("Pkt_PTS : ", avutil.GetPktPts(dec.Frame()), "/", avutil.GetPktPts(blackFrame))
	
						pictType := avutil.AvGetPictureTypeChar(dec.Frame().PictureType())
						if pictType == "B" || pictType == "I" || pictType == "P" {
							dec.IncrIbpFrameCount()
							dec.Frame().SetPts(dec.IbpFrameCount())
						}
						dec.Frame().SetPts(int64(dec.Context().FrameNumber()))

						if blackFrame != nil { 
							log.Trace(len(filters.Ins()), " 0")
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[0], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[1], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
							avfilter.AvBuffersrcWriteFrame(filters.Ins()[3], (*avfilter.Frame)(unsafe.Pointer(blackFrame)))
						}

						log.Trace(len(filters.Ins()), " 2")
						avfilter.AvBuffersrcAddFrame(filters.Ins()[2], (*avfilter.Frame)(unsafe.Pointer(dec.Frame())))

						log.Debug("----Get-----")
						frame := filters.GetFilterOutFrame()
						log.Debug("----End-----")
						if frame != nil {
							if enc.GeneratePacket(frame) == 0 {
								cache := enc.ToBytes()
								file.Write(cache)
								
								//log.Trace("wb++++\n", hex.EncodeToString(cache))

								l := len(cache)
								var startPos []int
								var nalus [][]byte
								j := 0 // split nalu in cache to nalus 
								for i := 0; i < l - 5; i++ {
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

									// H264 30FPS : 90000 / 30 : diff = 3000
									rtpPacket.SetTimeStamp(rtpPacket.TimeStamp() + 3000)
									
									for _, q := range rps {
										rtpPacket.SetSequence(rtpPacket.Sequence() + 1)
										rtpPacket.SetPayload(q)
										//r :=rtp.NewParser(1500)
										//copy(r.Buffer(), rtpPacket.GetRtpBytes())
										//r.SetPacketLength(len(rtpPacket.GetRtpBytes()))
										//r.Print("tag")
										//log.Trace("rps : ",p," : \n", hex.EncodeToString(rtpPacket.GetRtpBytes()))
										conn.WriteToUDP(rtpPacket.GetRtpBytes(), &net.UDPAddr{IP: net.ParseIP("192.168.0.78"), Port: 1236})
									}
								}
								
							}
						}
						avutil.AvFrameUnref(blackFrame)
						avutil.AvFrameUnref(dec.Frame())
					}
				}
			}
		}
	}()

	for {
		time.Sleep(500*time.Second)
	}
}



func main() {
	log.Info("--------main start---------")
	log.Debug("AvFilter Version:\t%v", avfilter.AvfilterVersion())
	log.Debug("AvCodec Version:\t%v", avcodec.AvcodecVersion())

	avutil.AvLogSetLevel(avutil.AV_LOG_TRACE)

	// Register all formats and codecs
	//avformat.AvRegisterAll()
	//avcodec.AvcodecRegisterAll()

	//testSaveWhiteBackground()
	//testSaveBlackBackground()
	//testDecodeEncode()
	//testSaveDescription4()
	//testRtpDEncodeDes4()
	//testUdpH264Des4()

	testRtpTransportDes4()
}
