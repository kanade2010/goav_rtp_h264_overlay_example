package rtp

import (
    "encoding/binary"
    "encoding/hex"
    "fmt"
)

const (
    rtpHeaderLength    = 12
    MaxRtpPacketSize   = 1500
)

const (
    markerPtOffset   = 1
    sequenceOffset   = 2
    timestampOffset  = sequenceOffset + 2
    ssrcOffsetRtp    = timestampOffset + 4
)

//! used to parser rtp header
const (
    version2Bit  = 0x80
    paddingBit   = 0x20
    extensionBit = 0x10
    markerBit    = 0x80
    ccMask       = 0x0f
    ptMask       = 0x7f
    countMask    = 0x1f
)

//! RtpParser Interface
type RtpParser interface {
	Version()        uint8
	Padding()        bool
	Extension()      bool
	CsrcCount()      uint8
    Marker()         bool
    PayloadType()    uint8
    SequenceNumber() uint16
    Timestamp()      uint32
    Ssrc()           uint32
    ExtensionData()  []byte
    Payload()        []byte
}

type RawPacket struct {
    buffer   []byte
}

//! RtpParser
type Parser struct{
    RawPacket
    PacketLength int
}

// Buffer returns the internal buffer in raw format.
// Usually only other Transports use the buffer in raw format, for example to encrypt
// or decrypt the buffer.
// Always call Buffer() just before the the buffer is actually used because several packet 
// handling functions may re-allocate buffers.
func NewParser(size int) *Parser {
    return &Parser{RawPacket{make([]byte, size + 64)}, 0}
}

func NewDefaultParser() *Parser {
    return &Parser{RawPacket{make([]byte, MaxRtpPacketSize + 64)}, 0}
}

func (raw *RawPacket) Buffer() []byte {
    return raw.buffer
}

func (rp *Parser) SetPacketLength(len int) {
    rp.PacketLength = len
}

// Version returns the version number.
func (rp *Parser) Version() uint8 {
    return (rp.buffer[0] & 0xc0)>>6
}

// Padding returns the state of the Padding bit.
// If the Padding bit is set the method return true, otherwise it returns false
func (rp *Parser) Padding() bool {
    return (rp.buffer[0] & paddingBit) == paddingBit
}

// ExtensionBit returns true if the Extension bit is set in the header, false otherwise.
func (rp *Parser) Extension() bool {
    return (rp.buffer[0] & extensionBit) == extensionBit
}

// CsrcCount return the number of CSRC values in this packet
func (rp *Parser) CsrcCount() uint8 {
    return rp.buffer[0] & ccMask
}

// Marker returns the state of the Marker bit.
// If the Marker bit is set the method return true, otherwise it returns false
func (rp *Parser) Marker() bool {
    return (rp.buffer[markerPtOffset] & markerBit) == markerBit
}

// PayloadType return the payload type value from RTP packet header.
func (rp *Parser) PayloadType() byte {
    return rp.buffer[markerPtOffset] & ptMask
}

// Sequence returns the sequence number as uint16 in host order.
func (rp *Parser) SequenceNumber() uint16 {
    return binary.BigEndian.Uint16(rp.buffer[sequenceOffset:])
}

// Timestamp returns the Timestamp as uint32 in host order.
func (rp *Parser) Timestamp() uint32 {
    return binary.BigEndian.Uint32(rp.buffer[timestampOffset:])
}

// Ssrc returns the SSRC as uint32 in host order.
func (rp *Parser) Ssrc() uint32 {
    return binary.BigEndian.Uint32(rp.buffer[ssrcOffsetRtp:])
}

// Extension returns the byte slice of the RTP packet extension part, if not extension available it returns nil.
// This is not a copy of the extension part but the slice points into the real RTP packet buffer.
func (rp *Parser) ExtensionData() []byte {
    if !rp.Extension() {
        return nil
    }
    offset := int(rp.CsrcCount()*4 + rtpHeaderLength)
    return rp.buffer[offset : offset+rp.ExtensionLength()]
}

// Payload returns the byte slice of the payload after removing length of possible padding.
//
// The slice is not a copy of the payload but the slice points into the real RTP packet buffer.
func (rp *Parser) Payload() []byte {
    payOffset := int(rp.CsrcCount()*4+rtpHeaderLength) + rp.ExtensionLength()
    pad := 0
    if rp.Padding() {
        pad = int(rp.buffer[rp.PacketLength-1])
    }
    return rp.buffer[payOffset : rp.PacketLength-pad]
}

// CsrcList returns the list of CSRC values as uint32 slice in host horder
func (rp *Parser) CsrcList() (list []uint32) {
    list = make([]uint32, rp.CsrcCount())
    for i := 0; i < len(list); i++ {
        list[i] = binary.BigEndian.Uint32(rp.buffer[rtpHeaderLength+i*4:])
    }
    return
}

// ExtensionLength returns the full length in bytes of RTP packet extension (including the main extension header).  
func (rp *Parser) ExtensionLength() (length int) {
    if !rp.Extension() {
        return 0
    }
    offset := int16(rp.CsrcCount()*4 + rtpHeaderLength) // offset to extension header 32bit word
    offset += 2
    length = int(binary.BigEndian.Uint16(rp.buffer[offset:])) + 1 // +1 for the main extension header word
    length *= 4
    return
}

// Print outputs a formatted dump of the RTP packet.
func (rp *Parser) Print(label string) {
    fmt.Printf("RTP Packet at: %s\n", label)
    fmt.Printf("  fixed header dump:   %s\n", hex.EncodeToString(rp.buffer[0:rtpHeaderLength]))
    fmt.Printf("    Version:           %d\n", rp.Version())
    fmt.Printf("    Padding:           %t\n", rp.Padding())
    fmt.Printf("    Extension:         %t\n", rp.Extension())
    fmt.Printf("    Contributing SRCs: %d\n", rp.CsrcCount())
    fmt.Printf("    Marker:            %t\n", rp.Marker())
    fmt.Printf("    Payload type:      %d (0x%x)\n", rp.PayloadType(), rp.PayloadType())
    fmt.Printf("    Sequence number:   %d (0x%x)\n", rp.SequenceNumber(), rp.SequenceNumber())
    fmt.Printf("    Timestamp:         %d (0x%x)\n", rp.Timestamp(), rp.Timestamp())
    fmt.Printf("    SSRC:              %d (0x%x)\n", rp.Ssrc(), rp.Ssrc())

    if rp.CsrcCount() > 0 {
        cscr := rp.CsrcList()
        fmt.Printf("  CSRC list:\n")
        for i, v := range cscr {
            fmt.Printf("      %d: %d (0x%x)\n", i, v, v)
        }
    }
    if rp.Extension() {
        extLen := rp.ExtensionLength()
        fmt.Printf("  Extentsion length: %d\n", extLen)
        offsetExt := rtpHeaderLength + int(rp.CsrcCount()*4)
        fmt.Printf("    extension: %s\n", hex.EncodeToString(rp.buffer[offsetExt:offsetExt+extLen]))
    }
    payOffset := rtpHeaderLength + int(rp.CsrcCount()*4) + rp.ExtensionLength()
    fmt.Printf("  payload: %s\n", hex.EncodeToString(rp.buffer[payOffset:rp.PacketLength]))
}
