package main

import (
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"math/rand"
)

type Order struct {
	Id string
	Name string
	Temp string// this should be an enum. TODO: Does Go have enums?
	ShelfLife uint
	DecayRate float32
}

func courier(order Order,
		counter *int32, arrival_time int,
		modifier uint, wg *sync.WaitGroup){
	// TODO: add proper logging for fetching,
	// displaying shelf contents after fetch.
	// TODO: add pointer for the array so we can remove it
	time.Sleep(time.Duration(1000*arrival_time)*time.Millisecond)
	a := float32(order.ShelfLife)
	b := order.DecayRate*float32(arrival_time) * float32(modifier)
	value := (a - b)/a
	// remove item from shelf in either scenario
	atomic.AddInt32(counter,1)
	// need to log the score in either scenario
	wg.Done()
	if (value <= 0){
		fmt.Println("Discarded item due to expiration")
	} else {
		fmt.Println("Courier fetched item")
	}
}

func selectShelf(order *Order,over_ct *int32,
		cold_ct *int32,
		hot_ct *int32, frozen_ct *int32,
		dead *int32) *int32{
	// TODO: add in moving average selection here
	// TODO: make this return an enum instead of a number

	if (*over_ct > 0){
		return over_ct
	}
	if (*cold_ct > 0 && order.Temp == "cold"){
		return cold_ct
	} else if (*hot_ct > 0 && order.Temp == "hot"){
		return hot_ct
	} else if (*frozen_ct > 0 && order.Temp == "frozen"){
		return frozen_ct
	} else {
		fmt.Println(order)
		panic("unknown temp")
	}
	return dead
}

func decrement(ct *int32){
	// TODO: add in a pointer to the array
	// so we can store the ID
	atomic.AddInt32(ct,-1)

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

	var wg  sync.WaitGroup
	over_ct,cold_ct,hot_ct,frozen_ct,dead := int32(15),int32(10),int32(10),int32(10),int32(0)
	for i:= 1; i < arrlen; i += 2 {
		fmt.Println(orders[i],i)
		fmt.Println(orders[i-1],i-1)
		blob_1,blob_2 := orders[i],orders[i-1]
		// TODO: add stuff for shelf contents
		shelf_1 := selectShelf(&blob_1, &over_ct,
				&cold_ct,&hot_ct,&frozen_ct,
				&dead)
		if (shelf_1 != &dead){
			decrement(shelf_1)
		}
		shelf_2 := selectShelf(&blob_2, &over_ct,
				&cold_ct,&hot_ct,&frozen_ct,
				&dead)
		if (shelf_2 != &dead){
			decrement(shelf_2)
		}
		arrival_1 := rand.Intn(6-2)+2
		arrival_2 := rand.Intn(6-2)+2
		// TODO: Add logging for dispatch
		if (shelf_1 != &dead){
			wg.Add(1)
			if (shelf_1 == &over_ct){
				go courier(blob_1, &over_ct,
				arrival_1, 2, &wg)
			} else {
				go courier(blob_1, shelf_1,
					arrival_1,1,&wg)
			}
		}
		if (shelf_2 != &dead){
			wg.Add(1)
			if (shelf_2 == &over_ct){
				go courier(blob_2, &over_ct,
				arrival_2, 2,&wg)
			} else {
				go courier(blob_2, shelf_2,
					arrival_2,1,&wg)
			}
		}
		time.Sleep(2000*time.Millisecond)
	}
	wg.Wait()
	fmt.Println("complete")
}

func main(){
	runQueue()

}
