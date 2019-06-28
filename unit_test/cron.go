package main


import(
	"../cron"
	"fmt"
	"time"
)

func main() {

	cron := cron.New()

	cron.Start()

	cron.AddFunc("*/3 * * * * ?", func(){fmt.Println("tick happen!")})

	time.Sleep(10000*time.Second)
}