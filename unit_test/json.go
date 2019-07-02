package main

import(
	"../gjson"
	"fmt"
	"encoding/json"
	//"os"
)

type RoomInfo struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address struct {
		AudioPort int `json:"audio_port"`
		VedioPort int `json:"vedio_port"`
	} `json:"address"`
	Accounts []string `json:"accounts"`
}

func main() {

	// parse json
	room := "{\"id\":3,\"name\":\"A401\",\"address\":{\"audio_port\":10000,\"vedio_port\":10002},\"accounts\":[\"1\",\"2\"]}"

	id   := gjson.Get(room, "id")
	name := gjson.Get(room, "name")
	address := gjson.Get(room, "address")
	accounts := gjson.Get(room, "accounts")
	vPort := gjson.Get(room, "address.vedio_port")

	fmt.Println("id :", id.Int())
	fmt.Println("name :", name.String())
	fmt.Println("vport to string :", vPort.String())
	fmt.Println("address : ", address.Type , address)
	fmt.Println("accounts : ", accounts.Type , accounts, accounts.Array())

	// encode json
    type ColorGroup struct {
        ID     int
        Name   string
        Colors []string
    }
    group := ColorGroup{
        ID:     1,
        Name:   "Reds",
        Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
    }
    b, err := json.Marshal(group)
    if err != nil {
        fmt.Println("error:", err)
    }

	fmt.Println(string(b))


	//test room info

	js := RoomInfo{
		ID: 	 0,
		Name:    "default",
		//Address: {0, 0},
	}

	js.Accounts = append(js.Accounts, "s0", "s1")

	b, err = json.Marshal(js)
	if err != nil {
        fmt.Println("error:", err)
	}

	fmt.Println(string(b))

}