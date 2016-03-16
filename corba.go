package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/heppu/corbaAi/hexMap"
	"github.com/heppu/space-tyckiting/clients/go/client"
)

// Struct for our CorbaAi
type CorbaAi struct {
	MyTeam     client.Team
	OtherTeams []client.Team
	Config     client.GameConfig
	Map        *hexMap.HexMap
	Actions    map[int]client.Action
}

// Name for Our Ai
const AiName string = "Corba"

func main() {
	// Create pointer to your AI struct
	a := &CorbaAi{}

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Call Run function from client package with pointer to your AI struct
	client.Run(a, AiName)
}

// Move function will be called after OnEvents function returns
func (c *CorbaAi) Move() (actions []client.Action) {
	log.Println("[corba][move]")

	// Move all bots randomly around the area
	for botId, a := range c.Actions {
		// Get valid move points
		validMoves := c.Map.GetValidMoves(botId)

		// Set Action
		a.Position = validMoves[rand.Intn(len(validMoves))]
		a.Type = client.BOT_MOVE
		c.Actions[botId] = a

		// Add action to list
		actions = append(actions, a)
	}
	c.Map.Reduce()
	return
}

// OnConnected function will be called when websocket connection is established
// and server has send connected message.
func (c *CorbaAi) OnConnected(msg client.ConnectedMessage) {
	log.Println("[corba][OnConnected]")
	//spew.Dump(msg)

	// Create new map
	c.Config = msg.Config
	c.Map = hexMap.NewHexMap(msg.Config, true)
}

// OnStart will be called when game starts and server sends start message
func (c *CorbaAi) OnStart(msg client.StartMessage) {
	log.Println("[corba][OnStart]")
	//spew.Dump(msg)

	// Save important information about game to our Ai struct
	c.MyTeam = msg.You
	c.OtherTeams = msg.OtherTeams
	c.Actions = make(map[int]client.Action)

	for i := 0; i < len(msg.You.Bots); i++ {
		c.Map.SetMyBot(&msg.You.Bots[i])
		c.Actions[msg.You.Bots[i].BotId] = client.Action{BotId: msg.You.Bots[i].BotId}
	}

	c.Map.InitEnemies(msg.OtherTeams)
}

// OnEvents will be called when server sends events message.
func (c *CorbaAi) OnEvents(msg client.EventsMessage) {
	//spew.Dump(msg)

	for _, e := range msg.Events {
		switch e.Type {

		case client.EVENT_HIT:
			log.Printf("[corba][OnEvents][[hit] : Bot %d\n", e.BotId.Int64)

		case client.EVENT_DIE:
			log.Printf("[corba][OnEvents][die] : Bot %d\n", e.BotId.Int64)

		case client.EVENT_RADAR_ECHO:
			log.Printf("[corba][OnEvents][radarEcho] : Pos %v\n", e.Position)

		case client.EVENT_SEE:
			log.Printf("[corba][OnEvents][see] : Bot %d\n", e.BotId.Int64)

		case client.EVENT_DETECTED:
			log.Printf("[corba][OnEvents][detected] : Bot %d\n", e.BotId.Int64)
			spew.Dump(e)

		case client.EVENT_DAMAGED:
			log.Printf("[corba][OnEvents][damaged] : Bot %d\n", e.BotId.Int64)

		case client.EVENT_MOVE:
			log.Printf("[corba][OnEvents][move] : Bot %d\n", e.BotId.Int64)
			c.Map.MoveMyBot(int(e.BotId.Int64))

		case client.EVENT_NOACTION:
			log.Printf("[corba][OnEvents][noaction] : Bot %d\n", e.BotId.Int64)

		case client.EVENT_END:
			log.Printf("[corba][OnEvents][end]\n")

		default:
			log.Printf("[corba][OnEvents][wut] : This shouldn't happen...\n")
		}
	}
}

// OnEvents will be called when server sends events message.
func (c *CorbaAi) OnEnd(msg client.EndMessage) {
	log.Println("[corba][OnEnd]")
	spew.Dump(msg)
}

// OnError will be called when server sends message that has unknow message type.
func (c *CorbaAi) OnError(msg string) {
	log.Println("[corba][OnError]")
	log.Println(msg)
}
