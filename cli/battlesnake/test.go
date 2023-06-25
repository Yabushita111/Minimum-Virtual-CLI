package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/board"
	"github.com/BattlesnakeOfficial/rules/maps"
	"github.com/pkg/browser"
	log "github.com/spf13/jwalterweatherman"
)

type SnakeState struct {
	URL        string
	Name       string
	ID         string
	LastMove   string
	Character  rune
	Color      string
	Head       string
	Tail       string
	Author     string
	Version    string
	Error      error
	StatusCode int
	Latency    time.Duration
}

type GameState struct {
	// Options
	Width               int
	Height              int
	Names               []string
	URLs                []string
	Timeout             int
	TurnDuration        int
	Sequential          bool
	GameType            string
	MapName             string
	ViewMap             bool
	UseColor            bool
	Seed                int64
	TurnDelay           int
	OutputPath          string
	ViewInBrowser       bool
	BoardURL            string
	FoodSpawnChance     int
	MinimumFood         int
	HazardDamagePerTurn int
	ShrinkEveryNTurns   int

	// Internal game state
	settings    map[string]string
	snakeStates map[string]SnakeState
	gameID      string
	httpClient  TimedHttpClient
	ruleset     rules.Ruleset
	gameMap     maps.GameMap
	outputFile  io.WriteCloser
	idGenerator func(int) string
}

func (gameState *GameState) Initialize() error {
	// Generate game ID
	gameState.gameID = "8d25c551-d275-4fb5-948e-2baa48f32a7a"

	// Set up HTTP client with request timeout
	if gameState.Timeout == 0 {
		gameState.Timeout = 500
	}
	gameState.httpClient = timedHTTPClient{
		&http.Client{
			Timeout: time.Duration(gameState.Timeout) * time.Millisecond,
		},
	}
	return nil
}

func (gameState *GameState) Run() error {
	boardGame := board.Game{
		ID:     gameState.gameID,
		Status: "running",
		Width:  gameState.Width,
		Height: gameState.Height,
		Ruleset: map[string]string{
			rules.ParamGameType: gameState.GameType,
		},
		RulesetName: gameState.GameType,
		RulesStages: []string{},
		Map:         gameState.MapName,
	}
	//modified by yabust
	// read turn0 (first line) in json file
	battlelogPath := "/Users/yabu/Battlesnake-rules/cli/battlesnake/battlelog/"
	battleid := "8d25c551-d275-4fb5-948e-2baa48f32a7a"
	filename := battlelogPath + battleid + ".json"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()

	// function for calling browsers
	boardServer := board.NewBoardServer(boardGame)
	serverURL, err := boardServer.Listen()
	if err != nil {
		fmt.Println("Error starting HTTP server: %w", err)
	}
	defer boardServer.Shutdown()
	fmt.Println(serverURL)
	boardURL := "http://localhost:3000"
	browserURL := fmt.Sprintf(boardURL+"?engine=%s&game=%s&autoplay=true", serverURL, gameState.gameID)
	browser.OpenURL(browserURL)

	//modified by yabust
	// send turn zero to websocket server
	// decord json file into event
	var event board.GameEvent
	json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&event)
	fmt.Println(event)
	fmt.Println("*****************************")
	boardServer.SendEvent(event)
	for scanner.Scan() {
		// modified by yabust
		// decord json file into event
		var event board.GameEvent
		json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&event)
		boardServer.SendEvent(event)
		fmt.Println(event)
		fmt.Println("*****************************")
	}
	boardServer.SendEvent(board.GameEvent{
		EventType: board.EVENT_TYPE_GAME_END,
		Data:      boardGame,
	})
	return nil
}

func main() {
	gameState := &GameState{}
	if err := gameState.Initialize(); err != nil {
		log.ERROR.Fatalf("Error initializing game: %v", err)
	}
	if err := gameState.Run(); err != nil {
		log.ERROR.Fatalf("Error running game: %v", err)
	}
}
