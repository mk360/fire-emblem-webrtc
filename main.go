package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"os"

	"log"

	"net/http"

	"webrtctest/screenshot"
	"webrtctest/util"

	"github.com/gonutz/w32"
	"github.com/pion/webrtc/v3"

	"github.com/gorilla/websocket"
)

// Define constants and variables

const webPort = ":8779"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	var handle = util.GetProcess()
	// Setup HTTP server
	http.HandleFunc("/ws", handleWebSocket)
	// var rect w32.RECT = w32.RECT{}
	// var rect = w32.GetWindowRect()
	// var height = rect.Bottom - rect.Top
	// var width = rect.Right - rect.Left

	http.HandleFunc("/test", func(res http.ResponseWriter, req *http.Request) {
		var rect = w32.GetWindowRect(w32.HWND(handle))
		img, err := screenshot.CaptureRect(handle, *rect)
		fmt.Println(rect)
		if err != nil {
			fmt.Println(err)
		}
		file, _ := os.Create("samir.png")
		defer file.Close()

		res.Header().Add("Content-Type", "image/png")
		png.Encode(file, img)
		png.Encode(res, img)

	})

	fmt.Printf("Starting server at http://localhost%s\n", webPort)
	log.Fatal(http.ListenAndServe(webPort, nil))
}

func createPeerConnection() (*webrtc.PeerConnection, error) {

	// Define ICE servers

	iceServers := []webrtc.ICEServer{

		{

			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	}

	// Create a new RTCPeerConnection

	config := webrtc.Configuration{

		ICEServers: iceServers,
	}

	peerConnection, err := webrtc.NewPeerConnection(config)

	if err != nil {

		return nil, err

	}

	// Handle ICE connection state changes

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {

		fmt.Printf("ICE Connection State has changed: %s\n", state.String())

	})

	return peerConnection, nil

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)

			break
		}

		var msg map[string]interface{}

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println(err)

			continue
		}

		switch msg["type"] {
		case "join":
			log.Printf("%s joined the session", msg["name"])

		case "offer":
			// Handle offer message

		case "answer":
			// Handle answer message

		case "candidate":
			// Handle ICE candidate message

		}
	}
}
