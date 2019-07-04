package main

import(
	"time"
	"fmt"
)

var w1 chan struct{} = make(chan struct{})
var w2 chan struct{} = make(chan struct{})
var wall chan struct{} = make(chan struct{})



func w1fun() {

    time.Sleep(6*time.Second)

	w1 <- struct{}{}
	<- wall
	fmt.Println("---w1  wall")

}

func w2fun() {

	time.Sleep(4*time.Second)

	w2 <- struct{}{}
	<- wall
	fmt.Println("---w2  wall")

}

func main() {

	fmt.Println("---main---")

	go w1fun()
	go w2fun()

	<- w1
	fmt.Println("---w1")
	<- w2
	fmt.Println("---w2")

	wall <- struct{}{}
	wall <- struct{}{}

	time.Sleep(30*time.Hour)

}