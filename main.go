package main

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"os"
//	"sync/atomic"
)

type Order struct {
	Id string
	Name string
	Temp string// this should be an enum. TODO: Does Go have enums?
	ShelfLife uint
	DecayRate float32
}

func courier(order Order, shelf string,
		counter *uint, arrival_time uint,
		modifier uint){
	// we'll want a pointer to the shelf as well
	// as the incrementer
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	a := float32(order.ShelfLife)
	b := order.DecayRate*float32(arrival_time) * float32(modifier)
	value := (a - b)/a
	// remove item from shelf in either scenario
	// need to log the score in either scenario
	if (value <= 0){
		fmt.Println("Discarded item due to expiration")
	} else {
		fmt.Println("Courier fetched item")
	}
}

func selectShelf(order *Order,over_ct int,
		cold_ct int,
		hot_ct int, frozen_ct int) int{

	available := []string{}
	if (over_ct > 0){
		available := append(available,"overflow")
	}
	if (cold_ct > 0 && order.Temp == "cold"){
		available := append(available,"cold")
	} else if (hot_ct > 0 && order.Temp == "hot"){
		available := append(available,"hot")

	} else if (frozen_ct > 0 && order.Temp == "frozen"){
		available := append(available,"frozen")

	} else {
		panic("unknown temp")
	}
	if (len(available) == 0){
		return 0
	}
	return 1
	//switch available[0]{
	//	case "overflow":
	//		return 1
	//	case "hot":
	//		return 2
	//	case "frozen":
	//		return 3
	//	default:
	//		return 0
	//}
}
// TODO: add in parameters for CLI options
func runQueue(){

	var orders []Order
	inputFile, err := os.Open("orders.json")
	if err != nil{
		panic(err)
	}
	fmt.Println("opened file")
	defer inputFile.Close()
	byteArray, err := ioutil.ReadAll(inputFile)
	if err != nil{
		panic(err)
	}
	json.Unmarshal(byteArray, &orders)
	arrlen := len(orders)
	fmt.Printf("array length is %d\n", arrlen)
	over_ct,cold_ct,hot_ct,frozen_ct := 15,10,10,10
	for i:= 1; i < arrlen; i += 2 {
		fmt.Println(orders[i],i)
		fmt.Println(orders[i-1],i-1)
		
	}
}

func main(){
	runQueue()

}
