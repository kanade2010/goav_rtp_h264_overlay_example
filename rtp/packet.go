package rtp

import(
    "encoding/binary"
)

const (
	MaxRtpPayloadSize int = 1500 - 64 
)
type Packet struct {
	header	[12]byte
	payload []byte
}

type shardUnitA struct {
	fuIndicator byte
	fuHeader	byte
}

func NewDefaultPacketWithH264Type() *Packet {
	h   := [12]byte {
		0x80,	//version 2  no P, X, CSRC 0
		0x60,   //type h264, Mark 0
		0x00, 0x00, // initial seq
		0x00, 0x00, 0x00, 0x00,  // timestamp
		0x00, 0x00, 0x00, 0x00,  // SSRC
	}

	return &Packet{
		header   :   h,
	}
}

// Sequence returns the sequence number as uint16 in host order.
func (p *Packet) Sequence() uint16 {
    return binary.BigEndian.Uint16(p.header[2:])
}

func (p *Packet) SetSequence(seq uint16) {
	binary.BigEndian.PutUint16(p.header[2:], seq)
}

func (p *Packet) Header() [12]byte {
	return p.header
}

func (p *Packet) PayloadType() byte {
	return (p.header[1] & 0x7f)
}

func (p *Packet) SetPayloadType(typ byte) {
	p.header[1] = p.header[1] & 0x80
	typ = typ & 0x7f
	p.header[1] = p.header[1] | typ
}

func (p *Packet) TimeStamp() uint32{
   return binary.BigEndian.Uint32(p.header[4:])
}

// setRtpTimeStamp takes a 32 unsigned timestamp in host order and sets it in network order in SR.
func (p *Packet) SetTimeStamp(stamp uint32) {
    binary.BigEndian.PutUint32(p.header[4:], stamp)
}

func (p *Packet) SetSsrc(ssrc uint32) {
    binary.BigEndian.PutUint32(p.header[8:], ssrc)
}

func (p *Packet) SetPayload(pay []byte) {
	p.payload = pay[0:]
}

func (p *Packet) GetRtpBytes() []byte {
	rp := append(p.header[0:], p.payload[0:]...)
	return rp 
}

func (p *Packet) ParserNaluToRtpPayload(nalu []byte) [][]byte {

	var ret [][]byte
	var n []byte

	if nalu[0] == 0 && nalu[1] == 0 && nalu[2] == 1 {
		n = nalu[3:]
	} else {
		n = nalu[4:]
	}

	if len(n) < MaxRtpPayloadSize {
		ret = append(ret, n)
	} else {
		//0                        是F
		//11                       是NRI
		//11100                    是FU Type，这里是28，即FU-A
		//1                        是S，Start，说明是分片的第一包
		//0                        是E，End，如果是分片的最后一包，设置为1，这里不是
		//0                        是R，Remain，保留位，总是0
		//00101                    是NAl Type
		saStart := shardUnitA{0, 0}
		saStart.fuIndicator = saStart.fuIndicator | n[0]
		saStart.fuIndicator = (saStart.fuIndicator & 0xe0) | 0x1c
		saStart.fuHeader    = (saStart.fuHeader | 0x80) | (n[0] & 0x1f)

		saSlice := shardUnitA{0, 0}
		saSlice.fuIndicator = saSlice.fuIndicator | n[0]
		saSlice.fuIndicator = (saSlice.fuIndicator & 0xe0) | 0x1c
		saSlice.fuHeader    = saSlice.fuHeader | (n[0] & 0x1f)

		saEnd := shardUnitA{0, 0}
		saEnd.fuIndicator = saEnd.fuIndicator | n[0]
		saEnd.fuIndicator = (saEnd.fuIndicator & 0xe0) | 0x1c
		saEnd.fuHeader    = (saEnd.fuHeader | 0x40) | (n[0] & 0x1f)


		offset := 1 //n offset
		start  := append(make([]byte, 0), saStart.fuIndicator, saStart.fuHeader)
		start   = append(start, n[offset:MaxRtpPayloadSize + offset - 2]...)
		ret     = append(ret, start)
		offset  = offset + MaxRtpPayloadSize - 2

		remain := len(n) - MaxRtpPayloadSize - 1
	
		for remain > MaxRtpPayloadSize - 2 {
			slice := append(make([]byte, 0), saSlice.fuIndicator, saSlice.fuHeader)
			slice  = append(slice, n[offset:MaxRtpPayloadSize + offset - 2]...)
			ret    = append(ret, slice)
			offset = offset + MaxRtpPayloadSize - 2
			remain = remain - MaxRtpPayloadSize - 2 
		}

		end := append(make([]byte, 0), saEnd.fuIndicator, saEnd.fuHeader)
		end  = append(end, n[offset:]...)
		ret  = append(ret, end)

	}

	return ret
}

