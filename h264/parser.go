package h264

import (
)

const (
	i_frame byte = 0
	p_frame byte = 1
	b_frame byte = 2
)

const (
	NalueTypeNotDefined byte = 0
	NalueTypeSlice      byte = 1  //slice_layer_without_partioning_rbsp() sliceheader
	NalueTypeDpa        byte = 2  // slice_data_partition_a_layer_rbsp( ), slice_header
	NalueTypeDpb        byte = 3  // slice_data_partition_b_layer_rbsp( )
	NalueTypeDpc        byte = 4  // slice_data_partition_c_layer_rbsp( )
	NalueTypeIdr        byte = 5  // slice_layer_without_partitioning_rbsp( ),sliceheader
	NalueTypeSei        byte = 6  //sei_rbsp( )
	NalueTypeSps        byte = 7  //seq_parameter_set_rbsp( )
	NalueTypePps        byte = 8  //pic_parameter_set_rbsp( )
	NalueTypeAud        byte = 9  // access_unit_delimiter_rbsp( )
	NalueTypeEoesq      byte = 10 //end_of_seq_rbsp( )
	NalueTypeEostream   byte = 11 //end_of_stream_rbsp( )
	NalueTypeFiller     byte = 12 //filler_data_rbsp( )
	NalueTypeFuA        byte = 28 //Shard unitA
	NalueTypeFuB        byte = 29 //Shard unitB
)

var ParameterSetStartCode = []byte{0x00, 0x00, 0x00, 0x01}
var StartCode 			  = []byte{0x00, 0x00, 0x01}

type Parser struct {
	naluByte 		byte
	shardA   		*shardUnitA

	// deprecated
	internalBuffer	[]byte   // use ParserToInternalSlice() to get a complete nalu packet
}

type shardUnitA struct {
	fuIndicator byte
	fuHeader	byte
}

func NewParser() *Parser {
	return &Parser{
		naluByte : 0,
		shardA   : &shardUnitA{0, 0},
	}
}

func (p *Parser) FillNaluHead(h byte) {
	p.naluByte = h
	p.shardA.fuIndicator = 0
	p.shardA.fuHeader    = 0
}

func (p *Parser) NaluType() byte {
	return p.naluByte & 0x1f
}

func (p *Parser) ShardA() *shardUnitA {
	return p.shardA
}

func (p *Parser) FillShadUnitA(s [2]byte) {
	p.shardA.fuIndicator = s[0]
	p.shardA.fuHeader 	 = s[1]
}

func (s *shardUnitA) IsStart() bool {
	return (s.fuHeader & 0x80) == 0x80
}

func (s *shardUnitA) IsEnd() bool {
	return (s.fuHeader & 0x40) == 0x40 
}

func (s *shardUnitA) NaluType() byte {
	return s.fuHeader & 0x1f
}

func (s *shardUnitA) NaluHeader() byte {
	s1 := s.fuIndicator & 0xe0
	s2 := s.fuHeader & 0x1f

	return (s1 | s2)
}


// deprecated
// must clear with ClearInternalBuffer()
func (p *Parser) ParserToInternalSlice(pkt []byte) bool {
	p.FillNaluHead(pkt[0])
	
	if p.NaluType() == NalueTypeFuA {
		p.FillShadUnitA([2]byte{pkt[0], pkt[1]})
		
		if p.ShardA().IsStart() {
			//fmt.Println("-----is fuA start------")
			p.internalBuffer = append(p.internalBuffer, StartCode[0:]...)
			p.internalBuffer = append(p.internalBuffer, p.ShardA().NaluHeader())
			p.internalBuffer = append(p.internalBuffer, pkt[2:]...)
			return false
		} else if p.ShardA().IsEnd() {
			//fmt.Println("-----is fuA end--------")
			p.internalBuffer = append(p.internalBuffer, pkt[2:]...)
			//fmt.Println("len: ", len(p.internalBuffer), ":\n", hex.EncodeToString(p.internalBuffer))
			return true
		} else {
			//fmt.Println("-----is fuA slice------")
			if len(p.internalBuffer) > 1920*1080*3 {
				panic("internalBuffer to Large, fix me")
			}

			p.internalBuffer = append(p.internalBuffer, pkt[2:]...)
			return false
		}
	} else {
		//fmt.Println("nalu : ", p.NaluType())			

		p.internalBuffer = append(p.internalBuffer, StartCode[0:]...)
		p.internalBuffer = append(p.internalBuffer, pkt[0:]...)
		//fmt.Println("len: ", len(p.internalBuffer), ":\n", hex.EncodeToString(p.internalBuffer))
		
		return true
	}
}

func (p *Parser) GetInternalBuffer() []byte{
	return p.internalBuffer
}

func (p *Parser) ClearInternalBuffer() {
	p.internalBuffer = p.internalBuffer[0:0]
}