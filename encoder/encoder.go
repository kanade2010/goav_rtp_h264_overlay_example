package encoder

import (
	log "github.com/astaxie/beego/logs"

	"../goav/avcodec"
	"../goav/avutil"
	
	"unsafe"
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

	avutil.AvFrameUnref(frame)
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
