package h264

import (
	"bytes"
	"errors"
	"io"
)

const (
	i_frame byte = 0
	p_frame byte = 1
	b_frame byte = 2
)

const (
	nalu_type_not_define byte = 0
	nalu_type_slice      byte = 1  //slice_layer_without_partioning_rbsp() sliceheader
	nalu_type_dpa        byte = 2  // slice_data_partition_a_layer_rbsp( ), slice_header
	nalu_type_dpb        byte = 3  // slice_data_partition_b_layer_rbsp( )
	nalu_type_dpc        byte = 4  // slice_data_partition_c_layer_rbsp( )
	nalu_type_idr        byte = 5  // slice_layer_without_partitioning_rbsp( ),sliceheader
	nalu_type_sei        byte = 6  //sei_rbsp( )
	nalu_type_sps        byte = 7  //seq_parameter_set_rbsp( )
	nalu_type_pps        byte = 8  //pic_parameter_set_rbsp( )
	nalu_type_aud        byte = 9  // access_unit_delimiter_rbsp( )
	nalu_type_eoesq      byte = 10 //end_of_seq_rbsp( )
	nalu_type_eostream   byte = 11 //end_of_stream_rbsp( )
	nalu_type_filler     byte = 12 //filler_data_rbsp( )
	nalu_type_fu_a       byte = 28 //Shard unitA
	nalu_type_fu_b       byte = 29 //Shard unitB
)

var startCode = []byte{0x00, 0x00, 0x00, 0x01}

type Parser struct {
	naluType 	byte
	shard   	*shardUnit
}

type shardUnit struct {
	fuIndicator byte
	fuHeader	byte
}

func NewParser() *Parser {
	return &Parser{
		0,
		nil,
	}
}

func (p *Parser) parserNaluHead(h byte) {
	p.naluType = h & 0x1f
}
