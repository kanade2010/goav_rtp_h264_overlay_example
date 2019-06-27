package main

import (
	log "github.com/astaxie/beego/logs"

	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avfilter"
	"github.com/giorgisio/goav/avutil"
	"io/ioutil"
	"time"
	"./decoder"
	"./encoder"
	"os"
	"strconv"
	"unsafe"
)

const (
	SplitScreen0  = 0
	SplitScreen1  = 1
	SplitScreen4  = 4
	SplitScreen8  = 8
	SplitScreen16 = 16
)
const (
	BlackColor   = "color=black@0.8:size=1280x720:rate=30:sar=1/1 [out]"
	WhiteColor   = "color=white@0.8:size=1280x720:rate=30:sar=1/1 [out]"
	Description1 = "color=white@0.8:size=1280x720:rate=30:sar=1/1 [background];" +
				   "[background][in0] overlay=x=0:y=0 [out]"
	Description4 = "color=white@0.8:size=1280x720:rate=30:sar=1/1 [background];" +
				   "[in0] scale=639x359 [in0_scale];" +
				   "[in1] scale=639x359 [in1_scale];" +
				   "[in2] scale=639x359 [in2_scale];" +
				   "[in3] scale=639x359 [in3_scale];" +
				   "[background][in0_scale] overlay=x=0:y=0 [background+in0_scale];" +
				   "[background+in0_scale][in1_scale] overlay=x=641:y=0 [background+in0_scale+in1_scale];" +
				   "[background+in0_scale+in1_scale][in2_scale] overlay=x=0:y=361 [background+in0_scale+in1_scale+in2_scale];" +
				   "[background+in0_scale+in1_scale+in2_scale][in3_scale] overlay=x=641:y=361 [out]"
)

type filter struct {
	ins 		[]*avfilter.Context
	out 		*avfilter.Context
	graph   	*avfilter.Graph
	frame 		*avutil.Frame
	description string //SplitScreen 0, 1, 4, 8, 16 
}

func (f *filter) GraphDump() {
	o := f.graph.AvfilterGraphDump("")
	log.Info("GraphDump : \n", o)
}

func New(description string) *filter {
	graph := avfilter.AvfilterGraphAlloc()
	if graph == nil {
		log.Critical("AvfilterGraphAlloc Failed.")
	}
	
	frame := avutil.AvFrameAlloc()
	if frame == nil {
		log.Critical("AvFrameAlloc failed.")
	}

	inputs  := avfilter.AvfilterInoutAlloc()
	outputs := avfilter.AvfilterInoutAlloc()
	if inputs == nil || outputs == nil {
		log.Critical("AvfilterInoutAlloc Failed.")
	}

	defer avfilter.AvfilterInoutFree(inputs)
	defer avfilter.AvfilterInoutFree(outputs)

	buffersrc  := avfilter.AvfilterGetByName("buffer")
	buffersink := avfilter.AvfilterGetByName("buffersink")
	if buffersink == nil || buffersrc == nil {
		log.Critical("AvfilterGetByName Failed.")
	}


	ret := graph.AvfilterGraphParse2(description, &inputs, &outputs)
	if ret < 0 {
		log.Critical("AvfilterInoutAlloc Failed des : ", avutil.ErrorFromCode(ret))
	}

	var ins []*avfilter.Context

	switch description {
		case BlackColor:
			log.Info("Create BlackColor Filter.")

		case WhiteColor:
			log.Info("Create WhiteColor Filter.")

		case Description1:
			log.Info("Create Description1 Filter.")

		case Description4:
			log.Info("Create Description4 Filter.")
			
			args := "video_size=1280x720:pix_fmt=0:time_base=1/30:pixel_aspect=1/1";
			for i := 0; i < 4; i++ {
				var in *avfilter.Context
				inName := "in" + strconv.Itoa(i)
				ret = avfilter.AvfilterGraphCreateFilter(&in, buffersrc, inName, args, 0, graph)
				if ret < 0 {
					log.Critical("AvfilterGraphCreateFilter Failed des : ", avutil.ErrorFromCode(ret))
				}
				ins = append(ins, in)
				log.Trace("-----append-ins-----", len(ins))
			}
			
			index := 0
			for cur := inputs; cur != nil; cur = cur.Next() {
				log.Debug("index :", index)
				ret = avfilter.AvfilterLink(ins[index], 0, cur.FilterContext(), cur.PadIdx())
				if ret < 0 {
					log.Critical("AvfilterLink Failed des : ", avutil.ErrorFromCode(ret))
				}
				index++
			}

		default:
			log.Critical("no such filter.")
	}

	var out *avfilter.Context
	ret = avfilter.AvfilterGraphCreateFilter(&out, buffersink, "out", "", 0, graph)
	if ret < 0 {
		log.Critical("AvfilterGraphCreateFilter Failed des : ", avutil.ErrorFromCode(ret))
	}

	ret = avfilter.AvfilterLink(outputs.FilterContext(), outputs.PadIdx(), out, 0)
	if ret < 0 {
		log.Critical("AvfilterLink Failed des : ", avutil.ErrorFromCode(ret))
	}

	ret = graph.AvfilterGraphConfig(0)
	if ret < 0 {
		log.Critical("AvfilterGraphConfig Failed des : ", avutil.ErrorFromCode(ret))
	}

	/*cur := inputs
	for i := 0; cur; cur = cur->next, i++ {
		ret := avfilter.AvfilterGraphCreateFilter( graph)
	}*/
	//log.Trace("GraphDump : \n", graph.AvfilterGraphDump(""))

	return &filter{
		ins         : ins,
		out 	    : out,
		graph       : graph,
		frame 		: frame,
		description : description,
	}
	//graph.AvfilterGraphCreateFilter()
}



/*func (f *filter) InsertInput() {
	buffersrc  := avfilter.AvfilterGetByName("buffer")
	if buffersrc == nil {
		log.Critical("AvfilterGetByName Failed des : ", avutil.ErrorFromCode(ret))
	}
}*/

// warning need unref !
func (f *filter) GetFilterOutFrame() *avutil.Frame {
	ret := avfilter.AvBufferSinkGetFrame(f.out, (*avfilter.Frame)(unsafe.Pointer(f.frame)))

	if ret == avutil.AvErrorEOF || ret == avutil.AvErrorEAGAIN {
		log.Trace(avutil.ErrorFromCode(ret))
		return nil
	}

	if ret < 0 {
		log.Critical("AvBufferSinkGetFrame Failed des : ", avutil.ErrorFromCode(ret))
		return nil
	}

	return f.frame

}

func testSaveDescription4() {
	filter := New(Description4)
	filter.GraphDump()

	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	enc.SetEncodeParams(1920, 1080,
					   avcodec.AV_PIX_FMT_YUV420P,
					   false, 2,
					   1, 30)
	file, err := os.Create("./des4.h264")
	if err != nil {
		log.Critical("Error Reading")
	}
	defer file.Close()

	go func() {
		filter0 := New(BlackColor)
		filter1 := New(BlackColor)
		filter2 := New(BlackColor)
		filter3 := New(BlackColor)

		time.Sleep(1*time.Second)
		
		for {
			time.Sleep(30*time.Millisecond)
			frame0 := filter0.GetFilterOutFrame()
			frame1 := filter1.GetFilterOutFrame()
			frame2 := filter2.GetFilterOutFrame()
			frame3 := filter3.GetFilterOutFrame()

			if frame0 != nil { 
				log.Trace(len(filter.ins), " 0")
				avfilter.AvBuffersrcAddFrame(filter.ins[0], (*avfilter.Frame)(unsafe.Pointer(frame0)))
			}
			if frame1 != nil {
				log.Trace(len(filter.ins), " 1")
				avfilter.AvBuffersrcAddFrame(filter.ins[1], (*avfilter.Frame)(unsafe.Pointer(frame1)))
			}
			if frame2 != nil {
				log.Trace(len(filter.ins), " 2")
				avfilter.AvBuffersrcAddFrame(filter.ins[2], (*avfilter.Frame)(unsafe.Pointer(frame2)))
			}
			if frame3 != nil { 
				log.Trace(len(filter.ins), " 3")
				avfilter.AvBuffersrcAddFrame(filter.ins[3], (*avfilter.Frame)(unsafe.Pointer(frame3)))
			}

			log.Debug("----Get-----")
			frame := filter.GetFilterOutFrame()
			log.Debug("----End-----")
			if frame != nil {
				if enc.GeneratePacket(frame) == 0 {
					cache := enc.ToBytes()
					file.Write(cache)
				}
			}
		}

	}()

	time.Sleep(500*time.Second)

	/*for {
		log.Debug("----Get-----")
		frame := filter.GetFilterOutFrame()
		log.Debug("----End-----")
		if frame != nil {
			if enc.GeneratePacket(frame) == 0 {
				cache := enc.ToBytes()
				file.Write(cache)
			}
		}
		time.Sleep(10*time.Millisecond)
	}*/
}

func testSaveWhiteBackground() {
	filter := New(WhiteColor)
	filter.GraphDump()

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

		frame := filter.GetFilterOutFrame()
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
	filter := New(BlackColor)
	filter.GraphDump()

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

		frame := filter.GetFilterOutFrame()
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
	data, err := ioutil.ReadFile("record.h264")
    if err != nil {
        log.Debug("File reading error", err)
        return
	}
	log.Debug("Open Success.")
	l := len(data)
    log.Debug("size of file:", l)

	dec := decoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_H264))
	
	enc := encoder.AllocAll(avcodec.CodecId(avcodec.AV_CODEC_ID_MPEG4))
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


func main() {
	log.Info("--------main start---------")
	log.Debug("AvFilter Version:\t%v", avfilter.AvfilterVersion())
	log.Debug("AvCodec Version:\t%v", avcodec.AvcodecVersion())
	// Register all formats and codecs
	//avformat.AvRegisterAll()
	//avcodec.AvcodecRegisterAll()

	//go testSaveWhiteBackground()
	//go testSaveBlackBackground()
	
	testSaveDescription4()
}