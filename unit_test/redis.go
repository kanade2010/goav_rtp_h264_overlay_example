package main

/*import(
	//"github.com/go-redis/redis"
  "fmt"
  "time"
  //"reflect"
  "unsafe"
	"../redigo/redis"
	log "github.com/astaxie/beego/logs"
)

type SubscribeCallback func (channel, message string)

type Subscriber struct {
  client redis.PubSubConn
  cbMap map[string]SubscribeCallback
}

func (c *Subscriber) Connect(ip string, port uint16) {
  conn, err := redis.Dial("tcp", "127.0.0.1:6379")
  if err != nil {
      log.Critical("redis dial failed.")
  }

  c.client = redis.PubSubConn{conn}
  c.cbMap = make(map[string]SubscribeCallback)

  go func() {
    for {
        log.Debug("wait...")
        switch res := c.client.Receive().(type) {
          case redis.Message:
              channel := (*string)(unsafe.Pointer(&res.Channel))
              message := (*string)(unsafe.Pointer(&res.Data))
              c.cbMap[*channel](*channel, *message)
          case redis.Subscription:
              fmt.Printf("%s: %s %d\n", res.Channel, res.Kind, res.Count)
          case error:
            log.Error("errorr...")
            continue
        }
    }
  }()

}

func (c *Subscriber) Close() {
  err := c.client.Close()
  if err != nil{
    log.Error("redis close error.")
  }
}

func (c *Subscriber) Subscribe(channel interface{}, cb SubscribeCallback) {
  err := c.client.Subscribe(channel)
  if err != nil{
    log.Critical("redis Subscribe error.")
  }

  c.cbMap[channel.(string)] = cb
}

func TestCallback1(chann, msg string){
  log.Debug("TestCallback1 channel : ", chann, " message : ", msg)
}

func TestCallback2(chann, msg string){
  log.Debug("TestCallback2 channel : ", chann, " message : ", msg)
}

func TestCallback3(chann, msg string){
  log.Debug("TestCallback3 channel : ", chann, " message : ", msg)
}

func main() {

  log.Info("===========main start============")

  var sub Subscriber
  sub.Connect("127.0.0.1", 6397)
  sub.Subscribe("test_chan1", TestCallback1)
  sub.Subscribe("test_chan2", TestCallback2)
  sub.Subscribe("test_chan3", TestCallback3)

  for{
   time.Sleep(1 * time.Second)
  }
}*/


import(
	"../redigo/redis"
	log "github.com/astaxie/beego/logs"
)

func main() {
  conn, err := redis.Dial("tcp", "127.0.0.1:6379")
  if err != nil {
      log.Critical("redis dial failed.")
  }
  defer conn.Close()

  _, err  = conn.Do("select", 3)
  if err != nil {
    log.Critical("redis select failed.")
    return 
  }

  log.Debug("--------start---------")

  v, err := conn.Do("get", "xiaozhang")
  if err != nil {
    log.Critical("redis get failed.")
    return 
  }
  log.Debug("-------- end ---------")

  log.Debug(v)

}