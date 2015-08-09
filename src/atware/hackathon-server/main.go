package main

import (
	"encoding/json"
	"fmt"
	_ "io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

var (
	signalMutex   = sync.Mutex{}
	currentSignal *Signal
)

const (
	SECOND = 1000
)

type Signal struct {
	Time       int64 `json:"time"`
	ActionType int   `json:"action_type"`
}

type Response struct {
	Code    int         `json:"code"`
	Content interface{} `json:"content"`
}

func MiddleWare(fn http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("URL: %v, Method %v", req.URL, req.Method)
		origin := req.Header.Get("Origin")

		log.Printf("origin: %v", origin)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if req.Method == "OPTIONS" {
			return
		}
		fn(w, req)
	}
}

func handlerWebsocket(ws *websocket.Conn) {
	ws.Write([]byte("Hello"))
}

func main() {

	fmt.Printf("server running at port :%v \n", 6011)
	http.HandleFunc("/signal", MiddleWare(signalHandler))
	http.Handle("/", http.FileServer(http.Dir("./public")))

	server := NewServer()
	go server.Listen()

	err := http.ListenAndServe(":6011", nil)

	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func signalHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "GET":
		if currentSignal == nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write(makeResponse(http.StatusFound, "don't have signal"))
			return
		}

		signalMutex.Lock()
		defer signalMutex.Unlock()

		//create new get signal
		getSignal := Signal{}
		getSignal.Time = currentSignal.Time - time.Now().Unix()
		getSignal.ActionType = currentSignal.ActionType
		rw.WriteHeader(http.StatusOK)
		rw.Write(makeResponse(http.StatusOK, getSignal))
		return
	case "POST":
		var signal Signal
		data, err := ioutil.ReadAll(req.Body)

		if err != nil {
			rw.WriteHeader(http.StatusBadGateway)
			rw.Write(makeResponse(http.StatusBadRequest, err))
			return
		}

		err = json.Unmarshal(data, &signal)
		if err == nil {
			signalMutex.Lock()
			defer signalMutex.Unlock()

			currentSignal = &signal
			fmt.Println(time.Now().Unix())
			currentSignal.Time = signal.Time + time.Now().Unix()
			rw.WriteHeader(http.StatusCreated)
			rw.Write(makeResponse(http.StatusCreated, currentSignal))
			return
		}

		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(makeResponse(http.StatusBadRequest, err))
		return
	default:
		response := Response{Code: http.StatusNotAcceptable, Content: "Method Not Support"}
		data, _ := json.Marshal(response)
		rw.WriteHeader(http.StatusNotAcceptable)
		rw.Write(data)
	}
}

func makeResponse(code int, content interface{}) []byte {
	data, _ := json.Marshal(Response{Code: code, Content: content})
	return data
}
