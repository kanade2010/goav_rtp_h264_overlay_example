package main

import (
	log "github.com/astaxie/beego/logs"

	"github.com/giorgisio/goav/avcodec"
	//"github.com/giorgisio/goav/avdevice"
	"github.com/giorgisio/goav/avfilter"
	//"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	//"github.com/giorgisio/goav/swresample"
	//"github.com/giorgisio/goav/swscale"
	"unsafe"
	"io/ioutil"
	"time"
	//"fmt"
)


type Decoder struct {
	codec 			   *avcodec.Codec
	context 		   *avcodec.Context
	parserContext      *avcodec.ParserContext
	pkt 			   *avcodec.Packet
	frame			   *avutil.Frame
}

func AllocAll(codecId avcodec.CodecId) *Decoder{
	pkt 		   := avcodec.AvPacketAlloc()
	if pkt == nil {
		log.Critical("AvPacketAlloc failed.")
	}

	codec 		   := avcodec.AvcodecFindDecoder(codecId)
	if codec == nil {
		log.Critical("AvcodecFindDecoder failed.")
	}

	context 	   := codec.AvcodecAllocContext3()
	if context == nil {
		log.Critical("AvcodecAllocContext3 failed.")
	}

	parserContext  := avcodec.AvParserInit(int(codecId))
	if parserContext == nil {
		log.Critical("AvParserInit failed.")
	}

	frame   	   := avutil.AvFrameAlloc()
	if frame == nil {
		log.Critical("AvFrameAlloc failed.")
	}

	err := context.AvcodecOpen2(codec, nil)
	if err < 0 {
		log.Critical("AvcodecOpen2 failed.")
	}

	return &Decoder{
		codec,
		context,
		parserContext,
		pkt,
		frame,
	}
}

func (d *Decoder) parserPacket(buf []byte, size int) int {
	return d.context.AvParserParse2(d.parserContext, d.pkt, buf, 
							size, avcodec.AV_NOPTS_VALUE, avcodec.AV_NOPTS_VALUE, 0)
}

// 0 success
func (d *Decoder) generateFrame() int {
	ret := d.context.AvcodecSendPacket(d.pkt)
	if ret < 0 {
		log.Error("AvcodecSendPacket err ", avutil.ErrorFromCode(ret))
		return ret
	}

	ret = d.context.AvcodecReceiveFrame((*avcodec.Frame)(unsafe.Pointer(d.frame)))
	if ret < 0 {
		log.Error("AvcodecReceiveFrame err ", avutil.ErrorFromCode(ret))
		return ret
	}

	return ret
}

func (d *Decoder) FreeAll() {
	avutil.AvFrameFree(d.frame)
	d.pkt.AvFreePacket()
	avcodec.AvParserClose(d.parserContext)
	d.context.AvcodecClose()
	//d.context.AvcodecFreeContext()
}

func main() {
	log.Info("--------decoder init---------")
	log.Debug("AvFilter Version:\t%v", avfilter.AvfilterVersion())
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

	decoder := AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))

	b := make([]byte, 4096 + 64)
	
	sum := 0
	for sum < l {
		remain := 4096
		for remain > 0 {
			copy(b, data[sum:sum + 4096])
			n := decoder.parserPacket(b, remain)
			log.Debug("parser ", n, "bytes")
			
			sum     = sum + n
			remain  = remain - n;

			log.Trace("--------", decoder.pkt.GetPacketSize())
			if decoder.pkt.GetPacketSize() > 0 {
				log.Debug(*decoder.pkt)
				
				if decoder.generateFrame() == 0 {
					data0 := avutil.Data(decoder.frame)[0]
					buf := make([]byte, decoder.pkt.GetPacketSize())
					startPos := uintptr(unsafe.Pointer(data0))
					for i := 0; i < decoder.pkt.GetPacketSize(); i++ {
						element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
						buf[i] = element
					}
				}

				avutil.AvFrameUnref((*avutil.Frame)(unsafe.Pointer(decoder.frame)))
				time.Sleep(1*time.Second)
				decoder.FreeAll()
				return
			}
		}
	}

	for {
		time.Sleep(1*time.Second)
	}

}



/*
func (p *Packet) GetPacketData() **uint8 {
	return (**uint8)(unsafe.Pointer(&p.data))
}

func (p *Packet) GetPacketSize() *int {
	return (*int)(unsafe.Pointer(&p.size))
}
*/