package hexMap

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/heppu/space-tyckiting/clients/go/client"
)

var upgrader = websocket.Upgrader{}

type WebsocketHandler struct {
	connections map[*websocket.Conn]interface{}
	hm          *HexMap
}

type HexMap struct {
	points          map[int]map[int]Point
	myBots          map[int]*client.Bot
	config          client.GameConfig
	positionHistory map[int][2]client.Position
}

type Point struct {
	PossibleBots map[int]bool
	Empty        bool
}

type Info struct {
	Map  []InfoPoint  `json:"map"`
	Bots []client.Bot `json:"bots"`
}

type InfoPoint struct {
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Empty bool  `json:"empty"`
	Bots  []int `json:"bots"`
}

func NewHexMap(c client.GameConfig, visualize bool) *HexMap {
	hm := &HexMap{config: c}

	hm.points = make(map[int]map[int]Point)
	hm.myBots = make(map[int]*client.Bot)
	hm.positionHistory = make(map[int][2]client.Position)

	// Initialize map with points
	var x = 0
	var y = 0
	var z = 0
	for i := -c.FieldRadius; i < c.FieldRadius+1; i++ {
		hm.points[i] = make(map[int]Point)

		for j := -x; j < c.FieldRadius+1-y; j++ {
			pb := make(map[int]bool)
			hm.points[i][j] = Point{
				PossibleBots: pb,
				Empty:        false,
			}
			z++
		}

		if x < c.FieldRadius {
			x++
		} else {
			y++
		}
	}

	if visualize {
		ws := WebsocketHandler{hm: hm}
		ws.connections = make(map[*websocket.Conn]interface{})

		http.HandleFunc("/socket", ws.listen)
		http.Handle("/", http.FileServer(http.Dir("debugger")))

		go http.ListenAndServe("localhost:8888", nil)
		go ws.sender()
	}
	return hm
}

// The json serialization is terrible here
func (ws *WebsocketHandler) sender() {
	ticker := time.NewTicker(1 * time.Second)
	var err error
	go func() {
		for {
			select {
			case <-ticker.C:
				// Create Json data
				r := Info{}
				// Create map points
				r.Map = make([]InfoPoint, 0)
				for i, _ := range ws.hm.points {
					for j, _ := range ws.hm.points[i] {
						bots := make([]int, 0)
						for k, _ := range ws.hm.points[i][j].PossibleBots {
							bots = append(bots, k)
						}
						r.Map = append(r.Map, InfoPoint{
							X:     i,
							Y:     j,
							Empty: ws.hm.points[i][j].Empty,
							Bots:  bots,
						})
					}
				}
				// Add bots
				r.Bots = make([]client.Bot, 0)
				for _, v := range ws.hm.myBots {
					r.Bots = append(r.Bots, *v)
				}

				// Send data to all open connections
				for c := range ws.connections {
					if err = c.WriteJSON(r); err != nil {
						fmt.Printf("[websocket][error] : Could not send new data: %s\n", err)
					}
				}
			}
		}
	}()
}

func (ws *WebsocketHandler) listen(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[websocket][error] : Could not create new socket: %s\n", err)
		return
	}

	fmt.Println("[websocket][verbose] : New socket opened.")
	ws.connections[conn] = nil
	defer func() {
		conn.Close()
		delete(ws.connections, conn)
	}()

	// Send game configurations
	conn.WriteJSON(ws.hm.config)

	for {
		if _, _, err = conn.ReadMessage(); err != nil {
			fmt.Printf("[websocket][error] : ReadMessage: %s\n", err)
			return
		}
	}
}

// This will be called first thing on each round
// We should use flood fill, etc. to keep map up to date
func (h *HexMap) Reduce() {

}

func (h *HexMap) InitEnemies(teams []client.Team) {
	var x = 0
	var y = 0

	for i := -h.config.FieldRadius; i < h.config.FieldRadius+1; i++ {
		for j := -x; j < h.config.FieldRadius+1-y; j++ {
			for _, t := range teams {
				for _, b := range t.Bots {
					h.points[i][j].PossibleBots[b.BotId] = true
				}
			}
		}

		if x < h.config.FieldRadius {
			x++
		} else {
			y++
		}
	}
}

func (h *HexMap) DetectEnemyBot(botId int) {

}

func (h *HexMap) SetMyBot(bot *client.Bot) {
	h.myBots[bot.BotId] = bot
	h.markEmpty(bot.Position.X, bot.Position.Y, h.config.See)

	// Initialize position history
	h.positionHistory[bot.BotId] = [2]client.Position{bot.Position, bot.Position}
}

func (h *HexMap) MoveMyBot(botId int) {
	if _, ok := h.myBots[botId]; ok {
		h.markEmpty(h.myBots[botId].Position.X, h.myBots[botId].Position.Y, h.config.See)
	}
}

func (h *HexMap) HitBot(botId, damage int) {
	if _, ok := h.myBots[botId]; ok {
		h.myBots[botId].Hp -= damage
	}
}

func (h *HexMap) markEmpty(x, y, r int) {
	for dx := -r; dx < r+1; dx++ {
		for dy := max(-r, -dx-r); dy < min(r, -dx+r)+1; dy++ {
			p := h.points[dx+x][dy+y]
			p.Empty = true
			for i := 0; i < len(p.PossibleBots); i++ {
				p.PossibleBots[i] = false
			}
		}
	}
}

// Get valid positios where bot can move
func (h *HexMap) GetValidMoves(botId int) []client.Position {
	return getPositionsInRange(h.myBots[botId].Position.X, h.myBots[botId].Position.Y, h.config.Move)
}

// Get valid positios where bot can use cannon
func (h *HexMap) GetValidCannons(botId int) []client.Position {
	return getPositionsInRange(0, 0, h.config.Cannon)
}

// Get valid positios where bot can use rader
func (h *HexMap) GetValidRadars(botId int) []client.Position {
	return getPositionsInRange(0, 0, h.config.Cannon)
}

// Get valid positions in hexagon for given radius
func getPositionsInRange(x, y, r int) (pos []client.Position) {
	for dx := -r; dx < r+1; dx++ {
		for dy := max(-r, -dx-r); dy < min(r, -dx+r)+1; dy++ {
			pos = append(pos, client.Position{dx + x, dy + y})
		}
	}
	return
}

// Select smallest integer from two integers
func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}

// Select biggest integer from two integers
func max(a, b int) int {
	if b > a {
		return b
	}
	return a
}
