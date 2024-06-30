package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/neovim/go-client/nvim"
)

type Event struct {
	method  string
	message []interface{}
}

var clients []*nvim.Nvim
var server *nvim.Nvim
var agg chan *Event
var err error

// Create new client for incoming connections and proxies all events to the
// embedded nvim instance
func handleClient(conn net.Conn) {
	log.Println("Handling Client...")
	defer conn.Close()

	client, err := nvim.New(conn, conn, conn, log.Printf)
	check(err)

	client.RegisterHandler("nvim_ui_attach", func(args ...interface{}) {
		log.Println("nvim_ui_attach: ", args)
		agg <- &Event{"nvim_ui_attach", args}
	})

	client.RegisterHandler("nvim_input", func(args []interface{}) {
		log.Println("nvim_input: ", args)
		agg <- &Event{"nvim_input", args}
	})

	client.RegisterHandler("nvim_ui_set_focus", func(args []interface{}) {
		log.Println("nvim_ui_set_focus: ", args)
		agg <- &Event{"nvim_ui_set_focus", args}
	})

	clients = append(clients, client)
	err = client.Serve()
	check(err)
}

// Create socket and listen for new TUI connections
func listen(sock string) {
	os.Remove(sock)
	l, err := net.Listen("unix", sock)
	defer l.Close()
	check(err)
	fmt.Println("Listing on " + sock + "...")

	for {
		conn, err := l.Accept()
		check(err)
		go handleClient(conn)
	}
}

// Start embedded nvim instance and forward TUI client messages
func startNvimInstance() {
	log.Println("Starting Nvim instance...")
	server, err = nvim.NewChildProcess(nvim.ChildProcessArgs("--embed"))
	check(err)

	// proxy messages from nvim to clients
	server.RegisterHandler("redraw", func(updates ...interface{}) {
		for _, client := range clients {
			client.WriteOut("Hello")
			log.Println("Hello")
		}
	})

	// proxy messages from clients to nvim
	agg = make(chan *Event, 10)
	go func() {
		for msg := range agg {
			if msg.method == "nvim_ui_attach" {
				err = server.Request(msg.method, nil, msg.message...)
			} else {
				err = server.Request(msg.method, nil, msg.message)
			}
			check(err)
		}
	}()
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
