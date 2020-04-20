package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type smartDevice struct {
	idDevice int
	idIOT    int
	name     string
	status   bool
}

type websocketData struct {
	idDevice   int
	deviceName string
	devicaInfo string
}

type authDevice struct {
	username  string
	email     string
	processor string
}
type auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var addr = flag.String("addr", "95.31.37.182:80", "http service address")

func main() {

	m := smartDevice{
		idDevice: 001,
		idIOT:    1,
		name:     "Led",
		status:   true,
	}

	findPorts()
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	Auth()
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
				return
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
				return
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

func Auth() {

	response, err := http.NewRequest("POST", "http://95.31.37.182/echo", nil)
	if err != nil {
		logrus.Fatal(err)
	}
	response.Header.Set("Authorization", "email test password test")
	response.Header.Set("Contetnt-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(response)
	if err != nil {
		logrus.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Print(string(body))
}
