package main

import (
	log "github.com/astaxie/beego/logs"

	"github.com/giorgisio/goav/avcodec"
	//"github.com/giorgisio/goav/avdevice"
	//"github.com/giorgisio/goav/avfilter"
	//"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	//"github.com/giorgisio/goav/swresample"
	//"github.com/giorgisio/goav/swscale"
	"unsafe"
	"io/ioutil"
	"time"
	"./decoder"
	"os"
	//"fmt"
)


type Encoder struct {
	codec 			   *avcodec.Codec
	context 		   *avcodec.Context
	pkt 			   *avcodec.Packet
}

func AllocAll(codecId avcodec.CodecId) *Encoder{
	pkt 		   := avcodec.AvPacketAlloc()
	if pkt == nil {
		log.Critical("AvPacketAlloc failed.")
	}

	codec 		   := avcodec.AvcodecFindEncoder(codecId)
	if codec == nil {
		log.Critical("AvcodecFindDecoder failed.")
	}

	context 	   := codec.AvcodecAllocContext3()
	if context == nil {
		log.Critical("AvcodecAllocContext3 failed.")
	}

	return &Encoder{
		codec,
		context,
		pkt,
	}
}

func (d *Encoder) Packet() *avcodec.Packet {
	return d.pkt
}

func (e *Encoder) SetEncodeParams(width int, height int, pxlFmt avcodec.PixelFormat, hasBframes bool, gopSize, num, den int) {
	e.context.SetEncodeParams2(width, height, pxlFmt, hasBframes, gopSize)
	e.context.SetTimebase(num, den)

	err := e.context.AvcodecOpen2(e.codec, nil)
	if err < 0 {
		log.Critical("AvcodecOpen2 failed.")
	}
}

// 0 success
func (e *Encoder) GeneratePacket(frame *avutil.Frame) int {
	ret := e.context.AvcodecSendFrame((*avcodec.Frame)(unsafe.Pointer(frame)))
	if ret < 0 {
		log.Trace("AvcodecSendPacket err ", avutil.ErrorFromCode(ret))
		return ret
	}

	ret = e.context.AvcodecReceivePacket(e.pkt)
	if ret < 0 {
		log.Trace("AvcodecReceiveFrame err ", avutil.ErrorFromCode(ret))
		return ret
	}

	return ret
}

func (e *Encoder) ToBytes() []byte {
	data0 := e.Packet().Data()
	buf := make([]byte, e.Packet().GetPacketSize())
	start := uintptr(unsafe.Pointer(data0))
	for i := 0; i < e.Packet().GetPacketSize(); i++ {
		elem := *(*uint8)(unsafe.Pointer(start + uintptr(i)))
		buf[i] = elem
	}

	e.Packet().AvPacketUnref()
	return buf
}

func main() {
	log.Info("--------decoder init---------")
	//log.Debug("AvFilter Version:\t%v", avfilter.AvfilterVersion())
	log.Debug("AvCodec Version:\t%v", avcodec.AvcodecVersion())
	// Register all formats and codecs
	//avformat.AvRegisterAll()
	//avcodec.AvcodecRegisterAll()

	data, err := ioutil.ReadFile("record.h264")
    if err != nil {
        log.Debug("File reading error", err)
        return
	}
	log.Debug("Open Success.")
	l := len(data)
    log.Debug("size of file:", l)

	dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	
	enc := AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_MPEG4))
	enc.SetEncodeParams(1104, 622,
					   avcodec.AV_PIX_FMT_YUV420P,
					   true, 2,
					   1, 30)
	
	b := make([]byte, 4096 + 64)
	
	file, err := os.Create("./out.mp4")
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
				log.Debug(*dec.Packet())
				
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
				}

				avutil.AvFrameUnref(dec.Frame())
				//time.Sleep(1*time.Second)
			}
		}
	}

	for {
		time.Sleep(1*time.Second)
	}

}