package main

import (
	"fmt"
	"net/http"
	_ "time"

	"golang.org/x/net/websocket"
)

var (
	maxId int = 0
)

const channelBufSize = 100

type Server struct {
	clients   map[int]*Client
	addCh     chan *Client
	sendAllCh chan *Response
}

func NewServer() *Server {

	return &Server{
		addCh:     make(chan *Client),
		sendAllCh: make(chan *Response),
		clients:   make(map[int]*Client),
	}
}

func (s *Server) Listen() {
	onConnected := func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				fmt.Println("Server Listen Error")
			}
		}()

		client := NewClient(ws, s)
		s.Add(client)
		client.Listen()
	}

	//handle websocket
	http.Handle("/ws", websocket.Handler(onConnected))

	for {
		select {
		case c := <-s.addCh:
			s.clients[c.id] = c
		case c := <-s.sendAllCh:
			s.sendAll(c)
		}
	}
}
func (s *Server) sendAll(response *Response) {
	for _, c := range s.clients {
		c.Write(response)
	}
}

func (s *Server) Add(client *Client) {
	s.addCh <- client
}

type Client struct {
	id       int
	chSignal chan *Response
	server   *Server
	ws       *websocket.Conn
}

func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

func (c *Client) listenWrite() {
	for {
		select {
		case response := <-c.chSignal:
			websocket.JSON.Send(c.ws, response)
		}
	}
}

func (c *Client) listenRead() {
	for {
		select {
		default:
			var signal Signal
			err := websocket.JSON.Receive(c.ws, &signal)

			fmt.Println("listenRead", signal)
			response := Response{Code: http.StatusOK, Content: signal}
			if err != nil {
				websocket.JSON.Send(c.ws, nil)
				fmt.Println("receive websocket loi cmnr")
				return
			}
			fmt.Println(response)
			c.server.sendAll(&response)
		}
	}
}

func (c *Client) Write(response *Response) {
	select {
	case c.chSignal <- response:
	default:
		fmt.Println("client Write default")
	}
}
func NewClient(ws *websocket.Conn, s *Server) *Client {
	maxId++
	ch := make(chan *Response, channelBufSize)
	return &Client{id: maxId, chSignal: ch, server: s, ws: ws}
}
