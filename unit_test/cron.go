package main


import(
	"../cron"
	"fmt"
	"time"
	"os"
)

func main() {

	args := os.Args
	fmt.Println(args)

	cron := cron.New()

	cron.Start()

	cron.AddFunc("*/3 * * * * ?", func(){fmt.Println("tick happen!")})

	time.Sleep(10000*time.Second)
}