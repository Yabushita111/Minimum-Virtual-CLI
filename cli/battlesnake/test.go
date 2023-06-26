package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/websocket"

	"github.com/pkg/browser"
	log "github.com/spf13/jwalterweatherman"
)

var events chan GameEvent

type GameEvent struct {
	MyEvent string `json:"myData"`
}

var gameID = "8d25c551-d275-4fb5-948e-2baa48f32a7a"
var battlelogPath = "/Users/yabu/Battlesnake-rules/cli/battlesnake/battlelog/"
var filename = battlelogPath + gameID + ".json"
var file, _ = os.Open(filename)
var scanner = bufio.NewScanner(file)

func main() {
	events = make(chan GameEvent, 1000)
	http.HandleFunc("/games/"+gameID, handleGame)
	http.HandleFunc("/games/"+gameID+"/events", handleWebsocket)
	go func() {
		err := http.ListenAndServe(":8888", nil)
		if err != http.ErrServerClosed {
			log.ERROR.Printf("Error in board HTTP server: %v", err)
		}
	}()
	board := "localhost:3000"
	serverURL := "localhost:8080"
	boardURL := fmt.Sprintf(board+"?engine=%s&game=%s&autoplay=true", serverURL, gameID)
	log.INFO.Printf("Opening board URL: %s", boardURL)
	if err := browser.OpenURL(boardURL); err != nil {
		log.ERROR.Printf("Failed to open browser: %v", err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleGame(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	gameID := "8d25c551-d275-4fb5-948e-2baa48f32a7a"
	battlelogPath := "/Users/yabu/Battlesnake-rules/cli/battlesnake/battlelog/"
	filename := battlelogPath + gameID + ".json"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	events <- GameEvent{scanner.Text()}
}
func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	gameID := "8d25c551-d275-4fb5-948e-2baa48f32a7a"
	battlelogPath := "/Users/yabu/Battlesnake-rules/cli/battlesnake/battlelog/"
	filename := battlelogPath + gameID + ".json"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	websocket.JSON.Send(ws, scanner.Text())
}
