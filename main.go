package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial.v1"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type smartDevice struct {
	IdDevice int
	IdIot    int
	Name     string
	Status   bool
}

type Websoket_data struct {
	Id_device   int
	Device_name string
	Divice_info string
}

type AuthDevice struct {
	Username  string
	Email     string
	Processor string
}

var addr = flag.String("addr", "95.31.37.182:80", "http service address")

func main() {

	m := smartDevice{
		IdDevice: 001,
		IdIot:    1,
		Name:     "Led",
		Status:   true,
	}

	findPorts()
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	logrus.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			err := c.ReadJSON(&m)
			if err != nil {
				logrus.Println("Error reading json.", err)
			}

			logrus.Printf("Got message: %#v\n", m)

		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:

			err = c.WriteJSON(&m)
			if err != nil {
				logrus.Println("Error reading json.", err)
			}
			logrus.Println(t)
		case <-interrupt:
			logrus.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logrus.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func findPorts() {
	ports, err := serial.GetPortsList()
	if err != nil {
		logrus.Fatal(err)
	}
	if len(ports) == 0 {
		logrus.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		logrus.Printf("Found port: %v\n", port)
	}

}
