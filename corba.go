package main

import (
	"log"
	"math/rand"
	"time"

	//"github.com/davecgh/go-spew/spew"
	"github.com/heppu/corbaAi/hexMap"
	"github.com/heppu/space-tyckiting/clients/go/client"
)

// Struct for our CorbaAi
type CorbaAi struct {
	MyTeam         client.Team
	OtherTeams     []client.Team
	Config         client.GameConfig
	Map            *hexMap.HexMap
	Actions        map[int]*client.Action
	WasLocated     map[int]bool
	Radared        map[int]*client.Position
	EnemyLocations map[int]*client.Position
	LastShot       *int
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

	var shooting = 0

	// Choose tactic base on how many bots we have
	switch len(c.Actions) {
	/*
		case 1:
			log.Println("Do something")
			break
		case 2:
			log.Println("Do something")
			break
	*/
	default:
		for botId, a := range c.Actions {
			// Set previous radared to nil
			c.Radared[botId] = nil

			// If our bot was seen activate run tactic
			if c.WasLocated[botId] {
				log.Printf("Bot %d was located run!", botId)

				// Get optimal new position from map
				a.Position = c.Map.Run(botId)
				a.Type = client.BOT_MOVE

				// Reset hit here
				c.WasLocated[botId] = false

				// Add action to list
				actions = append(actions, *a)
				continue

			}

			// If all other bots are shooting use last one to radar
			if shooting == len(c.Actions)-1 {
				if _, ok := c.EnemyLocations[*c.LastShot]; ok {
					a.Position = *c.EnemyLocations[*c.LastShot]
					a.Type = client.BOT_RADAR
					c.Radared[botId] = &a.Position

					// Add action to list
					actions = append(actions, *a)
					continue
				}
			}

			// We have shot some bot last round so lets go after him
			if c.LastShot != nil {
				// Check that we have location for that bot
				if pos, ok := c.EnemyLocations[*c.LastShot]; ok {
					shooting++
					a.Position = *pos
					a.Type = client.BOT_CANNON
					actions = append(actions, *a)
					continue
				}

			}

			// Check if we have detected new bots
			if len(c.EnemyLocations) > 0 {
				// Pick one bot from detected bots and shoot
				var last int
				for botId, pos := range c.EnemyLocations {
					shooting++
					last = botId
					c.LastShot = &last
					a.Position = *pos
					a.Type = client.BOT_CANNON
					actions = append(actions, *a)
					break
				}
				continue
			}

			// Fall back to radaring
			//validMoves := c.Map.GetValidRadars(botId)

			//a.Position = validMoves[rand.Intn(len(validMoves))]
			a.Position = c.Map.GetBotRadaringPoint(botId)
			a.Type = client.BOT_RADAR
			c.Radared[botId] = &a.Position

			// Add action to list
			actions = append(actions, *a)
		}
		break
	}
	//spew.Dump(actions)
	shooting = 0
	c.Map.Send()
	return
}

// OnConnected function will be called when websocket connection is established
// and server has sent connected message.
func (c *CorbaAi) OnConnected(msg client.ConnectedMessage) {
	log.Println("[corba][OnConnected]")
	//spew.Dump(msg)

	// Create new map
	c.Config = msg.Config
	c.Map = hexMap.NewHexMap(msg.Config, true)
	c.Map.Send()
}

// OnStart will be called when game starts and server sends start message
func (c *CorbaAi) OnStart(msg client.StartMessage) {
	log.Println("[corba][OnStart]")
	//spew.Dump(msg)

	// Save important information about game to our Ai struct
	c.MyTeam = msg.You
	c.OtherTeams = msg.OtherTeams
	c.Actions = make(map[int]*client.Action)
	c.Radared = make(map[int]*client.Position)
	c.WasLocated = make(map[int]bool)
	c.EnemyLocations = make(map[int]*client.Position)
	startPoints := c.Map.GetStartPoints(len(msg.You.Bots))

	for i := 0; i < len(msg.You.Bots); i++ {
		c.Map.SetMyBot(&msg.You.Bots[i], startPoints[i])
		c.Actions[msg.You.Bots[i].BotId] = &client.Action{BotId: msg.You.Bots[i].BotId}
		c.WasLocated[msg.You.Bots[i].BotId] = false
	}

	c.Map.InitEnemies(msg.OtherTeams)
	c.Map.Send()
}

// OnEvents will be called when server sends events message.
func (c *CorbaAi) OnEvents(msg client.EventsMessage) {
	//spew.Dump(msg)

	// Run before each round
	c.Map.Reduce()

	// Bots that didn't move
	stay := make(map[int]interface{})
	for botId, _ := range c.Actions {
		stay[botId] = nil
	}

	// Open points that where radared on last round
	for _, pos := range c.Radared {
		if pos != nil {
			c.Map.Radar(pos)
		}
	}

	// Clear enemy locations
	for botId, _ := range c.EnemyLocations {
		delete(c.EnemyLocations, botId)
	}

	for _, e := range msg.Events {
		switch e.Type {

		// This can happen to our or enemy bot so lets use damaged to detect hits on our bots
		case client.EVENT_HIT:
			log.Printf("[corba][OnEvents][[hit] : Bot %d\n", e.BotId.Int64)

		case client.EVENT_DIE:
			log.Printf("[corba][OnEvents][die] : Bot %d\n", e.BotId.Int64)
			// Remove bot from data structures
			if _, ok := c.Actions[int(e.BotId.Int64)]; ok {
				delete(c.Actions, int(e.BotId.Int64))
				delete(c.WasLocated, int(e.BotId.Int64))
			}

		case client.EVENT_RADAR_ECHO:
			log.Printf("[corba][OnEvents][radarEcho] : Pos %v\n", e.Position)
			c.Map.DetectEnemyBot(int(e.BotId.Int64), e.Position)
			c.EnemyLocations[int(e.BotId.Int64)] = &e.Position

		// This will happen when too bots see each other
		case client.EVENT_SEE:
			log.Printf("[corba][OnEvents][see] : Bot %d Source %d\n", e.BotId.Int64, e.Source.Int64)
			c.Map.DetectEnemyBot(int(e.BotId.Int64), e.Position)
			c.EnemyLocations[int(e.BotId.Int64)] = &e.Position

		case client.EVENT_DETECTED:
			log.Printf("[corba][OnEvents][detected] : Bot %d\n", e.BotId.Int64)
			c.WasLocated[int(e.BotId.Int64)] = true

		case client.EVENT_DAMAGED:
			log.Printf("[corba][OnEvents][damaged] : Bot %d damage %d\n", e.BotId.Int64, e.Damage.Int64)
			c.WasLocated[int(e.BotId.Int64)] = true
			c.Map.HitBot(int(e.BotId.Int64), int(e.Damage.Int64))

		case client.EVENT_MOVE:
			log.Printf("[corba][OnEvents][move] : Bot %d\n", e.BotId.Int64)
			c.Map.MoveMyBot(int(e.BotId.Int64), e.Position)
			delete(stay, int(e.BotId.Int64))

		case client.EVENT_NOACTION:
			log.Printf("[corba][OnEvents][noaction] : Bot %d\n", e.BotId.Int64)
			c.Map.Stay(int(e.BotId.Int64))

		case client.EVENT_END:
			log.Printf("[corba][OnEvents][end]\n")

		default:
			log.Printf("[corba][OnEvents][wut] : This shouldn't happen...\n")
		}
	}

	// Tell map that these bot's didn't move
	for botId, _ := range stay {
		c.Map.Stay(int(botId))
	}
}

// OnEvents will be called when server sends events message.
func (c *CorbaAi) OnEnd(msg client.EndMessage) {
	log.Println("[corba][OnEnd]")
	//spew.Dump(msg)
	if c.MyTeam.TeamId == msg.WinnerTeamId {
		log.Println("VICTORY!")
		return
	}
	log.Println("We lost...")
}

// OnError will be called when server sends message that has unknown message type.
func (c *CorbaAi) OnError(msg string) {
	log.Println("[corba][OnError]")
	log.Println(msg)
}
