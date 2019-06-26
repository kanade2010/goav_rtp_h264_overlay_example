package filter

import (
	log "github.com/astaxie/beego/logs"

	"../goav/avfilter"
	"../goav/avutil"
	
	"strconv"
	"unsafe"
)

const (
	BlackColor   = "color=black@0.8:size=1920x1080:rate=30:sar=1/1 [out]"
	WhiteColor   = "color=white@0.8:size=1920x1080:rate=30:sar=1/1 [out]"
	Description1 = "color=white@0.8:size=1920x1080:rate=30:sar=1/1 [background];" +
				   "[background][in0] overlay=x=0:y=0 [out]"
	Description4 = "color=white@0.8:size=1920x1080:rate=30:sar=1/1 [background];" +
				   "[in0] scale=959x539 [in0_scale];" +
				   "[in1] scale=959x539 [in1_scale];" +
				   "[in2] scale=959x539 [in2_scale];" +
				   "[in3] scale=959x539 [in3_scale];" +
				   "[background][in0_scale] overlay=x=0:y=0 [background+in0_scale];" +
				   "[background+in0_scale][in1_scale] overlay=x=961:y=0 [background+in0_scale+in1_scale];" +
				   "[background+in0_scale+in1_scale][in2_scale] overlay=x=0:y=541 [background+in0_scale+in1_scale+in2_scale];" +
				   "[background+in0_scale+in1_scale+in2_scale][in3_scale] overlay=x=961:y=541 [out]"
)

const (
	P720BlackColor   =  "color=black@0.8:size=1280x720:rate=30:sar=1/1 [out]"
	P720WhiteColor   =  "color=white@0.8:size=1280x720:rate=30:sar=1/1 [out]"
	P720Description1 =  "color=white@0.8:size=1280x720:rate=30:sar=1/1 [background];" +
				   	    "[background][in0] overlay=x=0:y=0 [out]"
	P720Description4 =  "color=white@0.8:size=1280x720:rate=30:sar=1/1 [background];" +
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

func (f *filter) Ins() []*avfilter.Context {
	return f.ins
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
		case BlackColor,
			 P720BlackColor:
			log.Info("Create BlackColor Filter.")

		case WhiteColor,
			 P720WhiteColor:
			log.Info("Create WhiteColor Filter.")

		case Description1,
			 P720Description1:
			log.Info("Create Description1 Filter.")

		case Description4,
			 P720Description4:
			log.Info("Create Description4 Filter.")

			var args string
			if description == Description4 {
				args = "video_size=1920x1080:pix_fmt=0:time_base=1/30:pixel_aspect=1/1"
			} else {
				args = "video_size=1280x720:pix_fmt=0:time_base=1/30:pixel_aspect=1/1"
			}
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

	//log.Trace("GraphDump : \n", graph.AvfilterGraphDump(""))

	return &filter{
		ins         : ins,
		out 	    : out,
		graph       : graph,
		frame 		: frame,
		description : description,
	}
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
