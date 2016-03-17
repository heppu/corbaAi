package hexMap

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/heppu/space-tyckiting/clients/go/client"
)

var upgrader = websocket.Upgrader{}

type HexMap struct {
	points          map[int]map[int]*Point
	connections     map[*websocket.Conn]interface{}
	myBots          map[int]*client.Bot
	config          client.GameConfig
	positionHistory map[int]*[2]client.Position
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

	hm.points = make(map[int]map[int]*Point)
	hm.myBots = make(map[int]*client.Bot)
	hm.positionHistory = make(map[int]*[2]client.Position)

	// Initialize map with points
	var x = 0
	var y = 0
	var z = 0
	for i := -c.FieldRadius; i < c.FieldRadius+1; i++ {
		hm.points[i] = make(map[int]*Point)

		for j := -x; j < c.FieldRadius+1-y; j++ {
			pb := make(map[int]bool)
			hm.points[i][j] = &Point{
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
		hm.connections = make(map[*websocket.Conn]interface{})
		http.HandleFunc("/socket", hm.listen)
		http.Handle("/", http.FileServer(http.Dir("debugger")))
		go http.ListenAndServe("localhost:8888", nil)
	}
	return hm
}

// The json serialization is terrible here
func (h *HexMap) Send() {
	// Create Json data
	r := Info{}
	// Create map points
	r.Map = make([]InfoPoint, 0)
	for i, _ := range h.points {
		for j, _ := range h.points[i] {
			bots := make([]int, 0)
			for k, _ := range h.points[i][j].PossibleBots {
				bots = append(bots, k)
			}
			r.Map = append(r.Map, InfoPoint{
				X:     i,
				Y:     j,
				Empty: h.points[i][j].Empty,
				Bots:  bots,
			})
		}
	}
	// Add bots
	r.Bots = make([]client.Bot, 0)
	for _, v := range h.myBots {
		r.Bots = append(r.Bots, *v)
	}

	// Send data to all open connections
	var err error
	for c := range h.connections {
		if err = c.WriteJSON(r); err != nil {
			fmt.Printf("[websocket][error] : Could not send new data: %s\n", err)
		}
	}

}

func (h *HexMap) listen(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[websocket][error] : Could not create new socket: %s\n", err)
		return
	}

	fmt.Println("[websocket][verbose] : New socket opened.")
	h.connections[conn] = nil
	defer func() {
		conn.Close()
		delete(h.connections, conn)
	}()

	// Send game configurations
	conn.WriteJSON(h.config)

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
	for m := 0; m < h.config.Move; m++ {
		markUnOpened := make([]*Point, 0)
		// Loop through all points
		for i, _ := range h.points {
			for j, p := range h.points[i] {
				// Check if point is open
				if p.Empty {
					// Check if point has unopened points around
					if h.checkIfBorderPoint(i, j) {
						markUnOpened = append(markUnOpened, p)
					}

				}
			}
		}

		// Mark points unopened
		for _, p := range markUnOpened {
			p.Empty = false
		}

		// Exapand possible bot positions
	}
}

func (h *HexMap) checkIfBorderPoint(x, y int) bool {
	r := 1
	for dx := -r; dx < r+1; dx++ {
		for dy := max(-r, -dx-r); dy < min(r, -dx+r)+1; dy++ {
			if p, ok := h.points[dx+x][dy+y]; ok {
				if p.Empty {
					return true
				}
			}
		}
	}
	return false
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
	h.positionHistory[bot.BotId] = &[2]client.Position{bot.Position, bot.Position}
}

func (h *HexMap) MoveMyBot(botId int, pos client.Position) {
	if bot, ok := h.myBots[botId]; ok {
		h.markEmpty(pos.X, pos.Y, h.config.See)
		bot.Position = pos

		// Keep bot mvoe history up to date
		h.positionHistory[bot.BotId][1].X = h.positionHistory[bot.BotId][0].X
		h.positionHistory[bot.BotId][1].Y = h.positionHistory[bot.BotId][0].Y
		h.positionHistory[bot.BotId][0].X = pos.X
		h.positionHistory[bot.BotId][0].Y = pos.Y
	}
}

func (h *HexMap) Stay(botId int) {
	if positions, ok := h.positionHistory[botId]; ok {
		h.markEmpty(positions[0].X, positions[0].Y, h.config.See)
	}
}

func (h *HexMap) HitBot(botId, damage int) {
	if bot, ok := h.myBots[botId]; ok {
		bot.Hp -= damage
	}
}

func (h *HexMap) markEmpty(x, y, r int) {
	r = r - 1
	// Single point case
	if r == 0 {
		if p, ok := h.points[x][y]; ok {
			p.Empty = true
			for i := 0; i < len(p.PossibleBots); i++ {
				p.PossibleBots[i] = false
			}
			h.points[x][y] = p
		}
		return
	}

	for dx := -r; dx < r+1; dx++ {
		for dy := max(-r, -dx-r); dy < min(r, -dx+r)+1; dy++ {
			if p, ok := h.points[dx+x][dy+y]; ok {
				h.points[dx+x][dy+y].Empty = true
				for i := 0; i < len(p.PossibleBots); i++ {
					p.PossibleBots[i] = false
				}
			}
		}
	}
}

// Get run away move
// 1. Gte all valid moves for bot
// 2. Filter bad moves out
// 3. Get random move from best moves
// 4. If no best moves get one from valids
func (h *HexMap) Run(botId int) client.Position {
	validMoves := h.GetValidMoves(botId)
	// Use these to detect our movign direction

	//currentPosition := h.positionHistory[botId][0]
	//lastPosition := h.positionHistory[botId][1]

	return validMoves[rand.Intn(len(validMoves))]
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
