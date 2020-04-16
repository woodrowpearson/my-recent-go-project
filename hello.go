package main

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"os"
)

type Order struct {
	Id string
	Name string
	Temp string// this should be an enum. TODO: Does Go have enums?
	ShelfLife uint
	DecayRate float32
}


func say(s string){
	time.Sleep(2000*time.Millisecond)
	fmt.Println(s)
}

func handle() int {
	go say("zort!")
	fmt.Println("crap")
	return 1
}

func read_from_file(){

	var orders []Order
	inputFile, err := os.Open("orders.json")
	if err != nil{
		panic (err)
	}
	fmt.Println("opened file")
	defer inputFile.Close()
	byteValue, err := ioutil.ReadAll(inputFile)
	if err != nil{
		panic(err)
	}
	json.Unmarshal(byteValue,&orders)
	fmt.Printf("Orders: %+v",orders)
	for order, idx := range orders {
		fmt.Printf("idx: %d, val: %+v\n", order,idx)
	}
}

func main(){
	res := handle()
	fmt.Println(res)
	time.Sleep(3000*time.Millisecond)
	fmt.Println("Bob Loblaw's Law Blog")
	read_from_file()
}
