package main

import (
	"encoding/json"
	"fmt"
	"github.com/stianeikeland/go-rpio"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"strconv"
	"time"
)

type Signal struct {
	Time       int64 `json:"time"`
	ActionType int   `json:"action_type"`
}

type Response struct {
	Code    int    `json:"code"`
	Content Signal `json:"content"`
}

func RunOther() (chan bool, chan bool) {
	chanSignal := make(chan bool, 5)

	backSignal := make(chan bool, 5)
	go func() {
		status := true

		for status {
			ch := pinOnOneSecond()
			select {
			case c := <-chanSignal:
				fmt.Printf("c: %v", c)
				backSignal <- c
				status = false
				break
			case <-ch:
			default:
				fmt.Println("default")
			}

		}
	}()

	return chanSignal, backSignal
}
func main() {
	origin := "http://192.168.147.79"
	url := "ws://192.168.147.79:6011/ws"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}

	chanSignal, backSignal := RunOther()

	for {

		var msg = make([]byte, 1024)
		var n int

		if n, err = ws.Read(msg); err == nil {
			var res Response
			fmt.Println(json.Unmarshal(msg[:n], &res))
			log.Println("action Type", res.Content.ActionType)
			switch res.Content.ActionType {

			// case 0: turn green light
			case 0:
				chanSignal <- true
				<-backSignal
				pinOn("green")
				pinOnAll(res.Content.Time)
				Close()
				fmt.Println("time", res.Content.Time)
				break
			// case 0: turn red light
			case 1:
				chanSignal <- true
				<-backSignal
				pinOn("red")
				log.Println("time", res.Content.Time)
				pinOnAll(res.Content.Time)
				Close()
				fmt.Println(res.Content.Time)
				break
			// mux red and green
			case 2:
				chanSignal <- true
				<-backSignal
				pinOn("red")
				pinOn("green")
				pinOnAll(res.Content.Time)
				Close()
				break
			default:
				log.Printf("type dont not supported %v \n", res.Content.ActionType)
			}

			chanSignal, backSignal = RunOther()
			fmt.Println("close")
		}
		//	fmt.Printf("Received: %s.\n", msg[:n])
	}

}

var (
	// Use mcu pin 10, corresponds to physical pin 19 on the pi
	pin = rpio.Pin(10)

	pinMapping = map[string]int{
		"red":   2,
		"green": 3,
		"0":     4,
		"1":     14,
		"2":     15,
		"3":     18,
		"4":     17,
		"5":     27,
		"6":     22,
		"7":     23,
		"8":     24,
		"9":     10,
	}
)

func pinOnAll(duration int64) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(duration*100) * time.Millisecond)
		pinOff("" + strconv.Itoa(i))
	}
	time.Sleep(time.Duration(duration*100) * time.Millisecond)
}

func pinOnOneSecond() <-chan bool {
	// if chan true turn pin One Second

	out := make(chan bool, 5)

	pinOn("red")
	for i := 0; i < 10; i++ {
		pinOff("" + strconv.Itoa(i))
		time.Sleep(50 * time.Millisecond)
	}
	Close()
	out <- true
	return out

}
func pinOffAll() {
	Close()
}
func pinOn(s string) {
	num := pinMapping[s]
	pin := rpio.Pin(num)
	pin.Output()
	pin.High()
}

func pinOff(s string) {
	num := pinMapping[s]
	pin := rpio.Pin(num)
	pin.Output()
	pin.Low()
}

func Close() {
	pinOff("red")
	pinOff("green")
	for i := 0; i < 10; i++ {
		pinOn("" + strconv.Itoa(i))
		time.Sleep(1 * time.Millisecond)
	}
}
func init() {
	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	Close()
}
