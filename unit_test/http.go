package main
 
import (
    "github.com/gin-gonic/gin"
	"net/http"
	"../redigo/redis"
	//"../gjson"
	"encoding/json"
	"os/exec"
	"strconv"
	"time"
	log "github.com/astaxie/beego/logs"
)

const (
	BitmapDB 	= 2
	AccountsDB 	= 3
	RoomDB 		= 4
	ListDB	 	= 5
)

// bitmap statistics online numbers
const (
	BitmapLessonsKey 	       = "bitmap:lessons"
	BitmapRoomsKey 	 	       = "bitmap:rooms"
	BitmapAccountsKey          = "bitmap:accounts"
) 

// server dynamic 
const (
	ListServerDynamicKey	   = "list:server:dynamic"
	ListMaxSize				   = 8

	EnterRoom   			   = "enter_room:"
	ExitsRoom                  = "exits_room:"
	NewRoom     			   = "new_room:"
	NewAccount   			   = "new_account:"
)

var redisConn redis.Conn

type RoomInfo struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address struct {
		AudioPort int `json:"audio_port"`
		VedioPort int `json:"vedio_port"`
	} `json:"address"`
	Accounts []string `json:"accounts"`
}

func Lpush(value string) error {
	err := redisConn.Send("select", ListDB) 
	if err != nil {
		return err
	}

	now := time.Now().Format("2006/1/02 15:04:05")
	err = redisConn.Send("lpush", ListServerDynamicKey, value + ":" + now)
	if err != nil {
		return err
	}

	err = redisConn.Send("ltrim", ListServerDynamicKey, 0, ListMaxSize -1)
	if err != nil {
		return err
	}

	err = redisConn.Flush()
	return err
}

func SetBit(key string, offset int, value string) error {
	err := redisConn.Send("select", BitmapDB) 
	if err != nil {
		return err
	}

	err = redisConn.Send("setbit", key, offset, value)
	if err != nil {
		return err
	}

	err = redisConn.Flush()
	return err
}

// 临时 use
var audio_port int = 0
var vedio_port int = 9998
func getVport() int {
	audio_port = vedio_port + 2
	vedio_port = audio_port + 2
	return vedio_port
}

var id int = -1
func inCreId() int {
	id++
	return id
}

func createNewRoom(context *gin.Context){
	
	room_name := context.Query("room_name")
	if room_name == "" {
		context.String(http.StatusBadRequest, "400, BadRequest!")
		return
	}
	
	redisConn.Do("select", RoomDB)
	exists, err := redis.Bool(redisConn.Do("exists", room_name))
	//log.Debug(room_name, exists)
	if exists == true || err != nil {
		log.Error("room already exist or redis error : ", err)
		context.String(http.StatusInternalServerError, "room already exist or database error!")
		return
	}

	// start an ffmpeg server 
	vport := getVport()
	cmdStr := "/home/ailumiyana/pesonal_work_dir/ws_tcp_server/goav_rtp_h264_overlay/ffmpeg-server " + strconv.Itoa(vport) + " &"
	log.Trace(cmdStr)
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	err  = cmd.Run()
	if err != nil {
		log.Error("start an ffmpeg server  failed : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}

	// update room info to redis  
	_, err  = redisConn.Do("select", 4)
	if err != nil {
		log.Error("redis select failed : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}

	new_room := RoomInfo{
		ID:		 inCreId(),
		Name:    room_name,
	}
	new_room.Address.VedioPort = vport
	jsBytes, err := json.Marshal(new_room)
	if err != nil {
		context.String(http.StatusBadRequest, "400, BadRequest!")
		return
	}

	setStr := string(jsBytes)
	log.Debug("Accepted : create new room : ", room_name, ":", setStr)

	_, err = redisConn.Do("set", room_name, setStr)
	if err != nil {
		log.Error("redis set failed : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}

	err = Lpush(NewRoom + room_name)
	if err != nil {
		log.Error("redis op failed : ", err)
	}

	err = SetBit(BitmapRoomsKey, new_room.ID, "1")
	if err != nil {
		log.Error("redis op failed : ", err)
	}

	context.String(http.StatusOK, "ok!")
	return
}

func deleteRoomByName(context *gin.Context){

	room_name := context.Query("room_name")
	if room_name == "" {
		context.String(http.StatusBadRequest, "400, BadRequest!")
		return
	}

	_, err := redisConn.Do("select", RoomDB)
	if err != nil {
		log.Error("redis select failed : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}
	
	room, err := redisConn.Do("get", room_name)
	if err != nil || room == nil {
		log.Error("room inexistence or redis error : ", err)
		context.String(http.StatusInternalServerError, "room inexistence or database error!")
		return
	}
	log.Trace(room)

	r := RoomInfo{}
	err = json.Unmarshal(room.([]byte), &r)
	if err != nil {
		log.Error("Internal json.Unmarshal error : ", err)
		context.String(http.StatusInternalServerError, "Internal error!")
		return
	}

	vPort := r.Address.VedioPort

	//vPort := gjson.Get(string(room.([]byte)), "address.vedio_port")
	cmdStr := "ps -ef |grep " + "\"ffmpeg-server " + strconv.Itoa(vPort) + "\" |grep -v grep | awk '{print $2}'"
	//log.Trace("exec :", cmdStr)

	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {

		log.Error("kill ffmpeg server failed : ", err)
		context.String(http.StatusInternalServerError, "Internal error!")
		return
	}
	
	b := make([]byte, 100)
	if err := cmd.Start(); err != nil {
		log.Error("kill ffmpeg server failed : ", err)
		context.String(http.StatusInternalServerError, "Internal error!")
		return
	}

	n, err := stdout.Read(b)
	if err != nil {
		log.Error("kill ffmpeg server failed : ", err)
		context.String(http.StatusInternalServerError, "Internal error!")
		return
	}
	stdout.Close()
	//log.Trace("b :", string(b[0:n-1]))

	cmd    = exec.Command("kill", "-9", string(b[0:n-1]))
	err    = cmd.Run()
	if err != nil {
		log.Error("kill ffmpeg server failed : ", err)
		context.String(http.StatusInternalServerError, "Internal error!")
		return
	}

	// del room
	_, err = redisConn.Do("del", room_name)
	if err != nil {
		log.Error("Internal redis op error : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}

	err = SetBit(BitmapRoomsKey, r.ID, "0")
	if err != nil {
		log.Error("Internal redis op error : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}

	context.String(http.StatusOK, "ok!")
	return
}

  
func getBusinessDataCounts(context *gin.Context){
	
	type reply struct {
		Data struct {
			AccountsCounts int `json:"accounts_counts"`
			RoomCounts int `json:"room_counts"`
		} `json:"data"`
	}	

	redisConn.Do("select", BitmapDB)
	rc, err := redis.Int(redisConn.Do("bitcount", BitmapRoomsKey))
	ac, err := redis.Int(redisConn.Do("bitcount", BitmapAccountsKey))
	
	r := reply{}

	r.Data.AccountsCounts = ac
	r.Data.RoomCounts     = rc
	
	b, err := json.Marshal(r)
    if err != nil {
		log.Error("Internal redis op error : ", err)
		context.String(http.StatusInternalServerError, "json.Marshal error!")
		return
    }

	context.String(http.StatusOK, string(b))
	return
}


func getBusinessDynamicInfo(context *gin.Context){
	type reply struct {
		Data []string `json:"data"`
	}

	redisConn.Do("select", ListDB)
	lr, err := redis.Strings(redisConn.Do("lrange", ListServerDynamicKey, 0, -1))
    if err != nil {
		log.Error("Internal redis op error : ", err)
		context.String(http.StatusInternalServerError, "database error!")
		return
	}
	
	r := reply{}
	r.Data = lr 

	b, err := json.Marshal(r)
    if err != nil {
		log.Error("Internal redis op error : ", err)
		context.String(http.StatusInternalServerError, "json.Marshal error!")
		return
	}
	
	context.String(http.StatusOK, string(b))
	return
}

/*func initLogger()(err error) {
    config := make(map[string]interface{})
    config["filename"] = beego.AppConfig.String("log_path")

    // map 转 json
    configStr, err := json.Marshal(config)
    if err != nil {
        fmt.Println("initLogger failed, marshal err:", err)
        return
    }
    // log 的配置
    beego.SetLogger(logs.AdapterFile, string(configStr))
    // log打印文件名和行数
    beego.SetLogFuncCall(true)
    fmt.Println(string(configStr))
    return
}*/

func main(){

	log.SetLogFuncCall(true)

	var err error
	redisConn, err = redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Critical("redis dial failed.")
	}
	defer redisConn.Close()
	
	_, err  = redisConn.Do("select", 4)
	if err != nil {
	  log.Critical("redis select failed.")
	} 

	router := gin.Default()
	router.GET("/createNewRoom", createNewRoom)
	router.GET("/deleteRoomByName", deleteRoomByName)
	router.GET("/getBusinessDataCounts", getBusinessDataCounts)
	router.GET("/getBusinessDynamicInfo", getBusinessDynamicInfo)

	router.Run(":9080")


}
