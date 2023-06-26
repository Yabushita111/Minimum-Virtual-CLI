package commands

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
	"github.com/spf13/cobra"
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

func NewPlayCommand() *cobra.Command {
	gameState := &GameState{}

	var playCmd = &cobra.Command{
		Use:   "play",
		Short: "Play a game of Battlesnake locally.",
		Long:  "Play a game of Battlesnake locally.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := gameState.Initialize(); err != nil {
				log.ERROR.Fatalf("Error initializing game: %v", err)
			}
			if err := gameState.Run(); err != nil {
				log.ERROR.Fatalf("Error running game: %v", err)
			}
		},
	}

	playCmd.Flags().IntVarP(&gameState.Width, "width", "W", 11, "Width of Board")
	playCmd.Flags().IntVarP(&gameState.Height, "height", "H", 11, "Height of Board")
	playCmd.Flags().StringArrayVarP(&gameState.Names, "name", "n", nil, "Name of Snake")
	playCmd.Flags().StringArrayVarP(&gameState.URLs, "url", "u", nil, "URL of Snake")
	playCmd.Flags().IntVarP(&gameState.Timeout, "timeout", "t", 500, "Request Timeout")
	playCmd.Flags().BoolVarP(&gameState.Sequential, "sequential", "s", false, "Use Sequential Processing")
	playCmd.Flags().StringVarP(&gameState.GameType, "gametype", "g", "standard", "Type of Game Rules")
	playCmd.Flags().StringVarP(&gameState.MapName, "map", "m", "standard", "Game map to use to populate the board")
	playCmd.Flags().BoolVarP(&gameState.ViewMap, "viewmap", "v", false, "View the Map Each Turn")
	playCmd.Flags().BoolVarP(&gameState.UseColor, "color", "c", false, "Use color to draw the map")
	playCmd.Flags().Int64VarP(&gameState.Seed, "seed", "r", time.Now().UTC().UnixNano(), "Random Seed")
	playCmd.Flags().IntVarP(&gameState.TurnDelay, "delay", "d", 0, "Turn Delay in Milliseconds")
	playCmd.Flags().IntVarP(&gameState.TurnDuration, "duration", "D", 0, "Minimum Turn Duration in Milliseconds")
	playCmd.Flags().StringVarP(&gameState.OutputPath, "output", "o", "", "File path to output game state to. Existing files will be overwritten")
	playCmd.Flags().BoolVar(&gameState.ViewInBrowser, "browser", false, "View the game in the browser using the Battlesnake game board")
	playCmd.Flags().StringVar(&gameState.BoardURL, "board-url", "https://board.battlesnake.com", "Base URL for the game board when using --browser")

	playCmd.Flags().IntVar(&gameState.FoodSpawnChance, "foodSpawnChance", 15, "Percentage chance of spawning a new food every round")
	playCmd.Flags().IntVar(&gameState.MinimumFood, "minimumFood", 1, "Minimum food to keep on the board every turn")
	playCmd.Flags().IntVar(&gameState.HazardDamagePerTurn, "hazardDamagePerTurn", 14, "Health damage a snake will take when ending its turn in a hazard")
	playCmd.Flags().IntVar(&gameState.ShrinkEveryNTurns, "shrinkEveryNTurns", 25, "In Royale mode, the number of turns between generating new hazards")

	playCmd.Flags().SortFlags = false

	return playCmd
}

// Setup a GameState once all the fields have been parsed from the command-line.
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

// Setup and run a full game.
func (gameState *GameState) Run() error {
	var err error
	if err != nil {
		return fmt.Errorf("Error initializing board: %w", err)
	}

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
	boardServer := board.NewBoardServer(boardGame)

	//modified by yabust
	// read turn0 (first line) in json file
	battlelogPath := "/Users/yabu/Battlesnake-rules/cli/battlesnake/battlelog/"
	gameID := "8d25c551-d275-4fb5-948e-2baa48f32a7a"
	filename := battlelogPath + gameID + ".json"
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan()

	serverURL, err := boardServer.Listen()
	if err != nil {
		return fmt.Errorf("Error starting HTTP server: %w", err)
	}
	defer boardServer.Shutdown()
	log.INFO.Printf("Board server listening on %s", serverURL)

	boardURL := fmt.Sprintf(gameState.BoardURL+"?engine=%s&game=%s&autoplay=true", serverURL, gameState.gameID)

	log.INFO.Printf("Opening board URL: %s", boardURL)
	if err := browser.OpenURL(boardURL); err != nil {
		log.ERROR.Printf("Failed to open browser: %v", err)
	}
	// modified by yabust
	// send turn zero to websocket server
	// ターン0だけここで処理しないとうまくいかない
	// decord json file into event
	var event board.GameEvent
	json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&event)
	boardServer.SendEvent(event)
	fmt.Println(event)
	log.INFO.Printf("Ruleset: %v, Seed: %v", gameState.GameType, gameState.Seed)
	for scanner.Scan() {
		// modified by yabust
		// decord json file into event
		var event board.GameEvent
		json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&event)
		boardServer.SendEvent(event)
	}
	return nil
}
