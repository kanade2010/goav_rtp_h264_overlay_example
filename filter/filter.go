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

const (
	Amix1 = "[in0] aformat=sample_rates=48000:sample_fmts=fltp:channel_layouts=stereo [stereo];" +
	"[stereo] asetnsamples=n=1024:p=1 [out0]"
	Amix2 = "[in0][in1] amix=inputs=2:duration=longest:dropout_transition=2"
	Amix4 = "[in0] amix=inputs=4:duration=longest:dropout_transition=2"
)

const (
	AudioArgsFmt = "sample_rate=%d:time_base=%d/%d:sample_fmt=%s"//:channel_layout=0x%I64x"
	VideoArgsFmt = "video_size=%dx%d:pix_fmt=%d:time_base=%d/%d:pixel_aspect=%d/%d"//:sws_param=flags=%d"

	//defaultAudioArgs = "time_base=1/44100"
	defaultAudioArgs = "time_base=1/48000:sample_rate=48000:sample_fmt=fltp:channel_layout=stereo"
	defaultVedioArgs = "video_size=1280x720:pix_fmt=0:time_base=1/30:pixel_aspect=1/1"
)

type Description struct {
	//inputsNumber int	 // inputs number
	Description  string
	Args 	     string	 
	AudioFilter  bool   // is that a audio filter otherwise a video filter
}

type Filter struct {
	ins 		[]*avfilter.Context
	//out 		*avfilter.Context  
	outs        []*avfilter.Context
	graph   	*avfilter.Graph
	frames 		[]*avutil.Frame		// filter output frames
	//description string //SplitScreen 1, 2, 3, 4, 8 
}

func (f *Filter) GraphDump() {
	o := f.graph.AvfilterGraphDump("")
	log.Info("GraphDump : \n", o)
}

func (f *Filter) Ins() []*avfilter.Context {
	return f.ins
}

// deprecated default video filter
func New(description string) *Filter {
	return NewAVFilter(Description{description, defaultVedioArgs, false})
}

func DefaultAFilter(description string) *Filter {
	return NewAVFilter(Description{description, defaultAudioArgs, true})
}

func DefaultVFilter(description string) *Filter {
	return NewAVFilter(Description{description, defaultVedioArgs, false})
}

func NewAVFilter(description Description) *Filter {
	graph := avfilter.AvfilterGraphAlloc()
	if graph == nil {
		log.Critical("AvfilterGraphAlloc Failed.")
	}
	
	/*frame := avutil.AvFrameAlloc()
	if frame == nil {
		log.Critical("AvFrameAlloc failed.")
	}*/

	inputs  := avfilter.AvfilterInoutAlloc()
	outputs := avfilter.AvfilterInoutAlloc()
	if inputs == nil || outputs == nil {
		log.Critical("AvfilterInoutAlloc Failed.")
	}

	defer avfilter.AvfilterInoutFree(inputs)
	defer avfilter.AvfilterInoutFree(outputs)

	var buffersrc *avfilter.Filter
	var buffersink *avfilter.Filter
	if description.AudioFilter {
		buffersrc  = avfilter.AvfilterGetByName("abuffer")
		buffersink = avfilter.AvfilterGetByName("abuffersink")

	} else {
		buffersrc  = avfilter.AvfilterGetByName("buffer")
		buffersink = avfilter.AvfilterGetByName("buffersink")
	}
	
	if buffersink == nil || buffersrc == nil {
		log.Critical("AvfilterGetByName Failed.")
	}

	ret := graph.AvfilterGraphParse2(description.Description, &inputs, &outputs)
	if ret < 0 {
		log.Critical("AvfilterInoutAlloc Failed des : ", avutil.ErrorFromCode(ret))
	}

	var ins    []*avfilter.Context
	var outs   []*avfilter.Context
	var frames []*avutil.Frame

	switch description.Description {
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

		default:
			log.Critical("user-defined filter.")
	}
	
	// inputs
	index := 0
	for cur := inputs; cur != nil; cur = cur.Next() {
		//log.Debug("index :", index)
		var in *avfilter.Context
		//var args = "video_size=1280x720:pix_fmt=0:time_base=1/30:pixel_aspect=1/1"
		inName := "in" + strconv.Itoa(index)
		ret = avfilter.AvfilterGraphCreateFilter(&in, buffersrc, inName, description.Args, 0, graph)
		if ret < 0 {
			log.Critical("AvfilterGraphCreateFilter Failed des : ", avutil.ErrorFromCode(ret))
		}

		ins = append(ins, in)
		ret = avfilter.AvfilterLink(ins[index], 0, cur.FilterContext(), cur.PadIdx())
		if ret < 0 {
			log.Critical("AvfilterLink Failed des : ", avutil.ErrorFromCode(ret))
		}
		index++
	}

	// outputs
	index = 0
	for cur := outputs; cur != nil; cur = cur.Next() {
		var out *avfilter.Context
		outName := "out" + strconv.Itoa(index)
		ret = avfilter.AvfilterGraphCreateFilter(&out, buffersink, outName, "", 0, graph)
		if ret < 0 {
			log.Critical("AvfilterGraphCreateFilter Failed des : ", avutil.ErrorFromCode(ret))
		}

		outs = append(outs, out)
		ret = avfilter.AvfilterLink(cur.FilterContext(), cur.PadIdx(), outs[index], 0)
		if ret < 0 {
			log.Critical("AvfilterLink Failed des : ", avutil.ErrorFromCode(ret))
		}
		index++

		f := avutil.AvFrameAlloc()
		if f == nil {
			log.Critical("AvFrameAlloc failed.")
		}
		frames = append(frames, f)
	}
	// out
	/*var out *avfilter.Context
	ret = avfilter.AvfilterGraphCreateFilter(&out, buffersink, "out", "", 0, graph)
	if ret < 0 {
		log.Critical("AvfilterGraphCreateFilter Failed des : ", avutil.ErrorFromCode(ret))
	}

	ret = avfilter.AvfilterLink(outputs.FilterContext(), outputs.PadIdx(), out, 0)
	if ret < 0 {
		log.Critical("AvfilterLink Failed des : ", avutil.ErrorFromCode(ret))
	}*/

	ret = graph.AvfilterGraphConfig(0)
	if ret < 0 {
		log.Critical("AvfilterGraphConfig Failed des : ", avutil.ErrorFromCode(ret))
	}

	//log.Trace("GraphDump : \n", graph.AvfilterGraphDump(""))

	return &Filter{
		ins         : ins,
		outs 	    : outs,
		graph       : graph,
		frames 		: frames,
		//description : description,
	}
}


/*func (f *filter) InsertInput() {
	buffersrc  := avfilter.AvfilterGetByName("buffer")
	if buffersrc == nil {
		log.Critical("AvfilterGetByName Failed des : ", avutil.ErrorFromCode(ret))
	}
}*/

// warning need unref !
// used for one output filter.
func (f *Filter) GetFilterOutFrame() *avutil.Frame {
	ret := avfilter.AvBufferSinkGetFrame(f.outs[0], (*avfilter.Frame)(unsafe.Pointer(f.frames[0])))

	if ret == avutil.AvErrorEOF || ret == avutil.AvErrorEAGAIN {
		log.Trace(avutil.ErrorFromCode(ret))
		return nil
	}

	if ret < 0 {
		log.Critical("AvBufferSinkGetFrame Failed des : ", avutil.ErrorFromCode(ret))
		return nil
	}

	return f.frames[0]
}

// filter output frames.
func (f *Filter) GetFilterOutFrames() []*avutil.Frame {
	for i, _ := range f.frames {
		ret := avfilter.AvBufferSinkGetFrame(f.outs[i], (*avfilter.Frame)(unsafe.Pointer(f.frames[i])))

		if ret == avutil.AvErrorEOF || ret == avutil.AvErrorEAGAIN {
			log.Trace(avutil.ErrorFromCode(ret))
			return nil
		}

		if ret < 0 {
			log.Critical("AvBufferSinkGetFrame Failed des : ", avutil.ErrorFromCode(ret))
			return nil
		}
	}

	return f.frames
}

func (f *Filter) FreeAll() {
	for i := 0; i < len(f.ins); i++ {
		f.ins[i].AvfilterFree()
	}

	for i := 0; i < len(f.outs); i++ {
		f.outs[i].AvfilterFree()
	}

	for i, _ := range f.frames {
		avutil.AvFrameFree(f.frames[i])
	}

	f.graph.AvfilterGraphFree()
}