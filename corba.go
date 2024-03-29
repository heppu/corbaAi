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
	OurCount         int
	EnemyCount       int
	MyTeam           client.Team
	OtherTeams       []client.Team
	Config           client.GameConfig
	Map              *hexMap.HexMap
	Actions          map[int]*client.Action   // Here we store bot actions and map bots that are alive
	WasLocated       map[int]bool             // True if our bot was located
	WasHit           map[int]bool             // True if our bot was shot
	Radared          map[int]*client.Position // This is used to open points we radared on last round
	EnemyLocations   []*client.Position       // Enemy positions
	LastShotPosition *client.Position         // Position for bot we shot in last round
}

// Name for Our Ai
const AiName string = "corbaKotka"

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
	//log.Println("[corba][move]")

	var positions = make([]client.Position, 0)

	// This is our battle beast mode.
	// It will be activated if we know enemylocation and we have more than two bots alive
	if len(c.EnemyLocations) > 0 && len(c.Actions) > 1 {
		log.Println("BATTLE BEAST")

		var running int
		var runningId *int
		var runningPos *client.Position

		for botId, a := range c.Actions {
			// Allow one bot to run
			if located, ok := c.WasLocated[botId]; ok && located && running == 0 {
				// Check if we have detected enemies
				// Run towards them hoping they use friendly fire d:D
				if len(c.EnemyLocations) > 0 {
					log.Println("A : ", *c.EnemyLocations[0])
					a.Position = c.Map.RunTowardsPosition(botId, *c.EnemyLocations[0])
				} else {
					// Get optimal new position from map
					a.Position = c.Map.Run(botId)
				}
				log.Println("RUN ", a.Position)
				a.Type = client.BOT_MOVE

				// Add action to list
				actions = append(actions, *a)
				running++
				runningId = &botId
				runningPos = &a.Position
				break
			}
		}

		var botPos *client.Position
		// Pick one bot from detected bots and get valid shooting points for each bot
		for _, pos := range c.EnemyLocations {
			botPos = pos
			positions = c.Map.ShootAround(*botPos, len(c.Actions)-1, runningPos)
			if len(positions) > 0 {
				break
			}
		}

		// We got valid shooting positions so let's cannon the shit out of that bot
		if len(positions) > 0 {
			var i int

			for botId, a := range c.Actions {
				// Reset hits
				defer func(botId int) {
					c.WasHit[botId] = false
					c.WasLocated[botId] = false
				}(botId)

				if runningId != nil && *runningId == botId {
					continue
				}

				// This happens when one bot is running and we have only two bots
				if 0 == (len(c.Actions) - 1 - running) {
					if c.Actions[botId].Type == client.BOT_RADAR {
						a.Type = client.BOT_CANNON
					} else {
						log.Println("THIS SHOULD NOT HAPPEN!!!!!")
					}

					c.LastShotPosition = botPos
					a.Position = *botPos
					actions = append(actions, *a)
					break
				}

				if i < (len(c.Actions) - 1 - running) {
					// These are one or two first bots and they will use cannon
					if i < len(positions) {
						a.Position = positions[i]
					} else {
						a.Position = positions[len(positions)-1]
					}
					a.Type = client.BOT_CANNON
					c.LastShotPosition = botPos
				} else {
					// This is our last bot so it will use radar
					a.Position = *botPos
					a.Type = client.BOT_RADAR
				}
				actions = append(actions, *a)
				i++
			}
			//c.Map.Send()
			return
		}
	}

	for botId, a := range c.Actions {
		// Reset hits
		defer func(botId int) {
			c.WasHit[botId] = false
			c.WasLocated[botId] = false
		}(botId)

		// Set previous radared to nil
		c.Radared[botId] = nil

		// If our bot was hit activate run tactic
		// Or if this is our last bot and it was radared run
		if c.WasHit[botId] ||
			(len(c.Actions) == 1 && c.WasLocated[botId] && len(c.EnemyLocations) == 0) {
			log.Printf("Bot %d was located run!", botId)

			// Check if we have detected enemies
			// Run towards them hoping they use friendly fire d:D
			if len(c.EnemyLocations) > 0 {
				a.Position = c.Map.RunTowardsPosition(botId, *c.EnemyLocations[0])
			} else {
				// Get optimal new position from map
				a.Position = c.Map.Run(botId)
			}

			a.Type = client.BOT_MOVE
			actions = append(actions, *a)
			continue

		}

		// This happens if we have only 1 bot left and we know enemy location
		if len(c.EnemyLocations) > 0 {
			lastPos := *c.EnemyLocations[0]

			positions = c.Map.ShootAround(lastPos, 1, nil)

			// This is to prevent our shooting our self
			if len(positions) == 0 {
				a.Position = c.Map.RunTowardsPosition(botId, *c.EnemyLocations[0])
				a.Type = client.BOT_MOVE
				actions = append(actions, *a)
				continue
			}

			a.Type = client.BOT_CANNON
			a.Position = positions[0]
			c.LastShotPosition = &lastPos
			actions = append(actions, *a)
			continue
		}

		// Fall back to radaring
		// If we shot some bot in last round use that as radaring point
		if c.LastShotPosition != nil {
			log.Println("USE LAST POSITION : ", *c.LastShotPosition)
			a.Position = *c.LastShotPosition
			c.LastShotPosition = nil
		} else {
			a.Position = c.Map.GetBotRadaringPoint(botId)
		}
		a.Type = client.BOT_RADAR
		c.Radared[botId] = &a.Position

		// Add action to list
		actions = append(actions, *a)
	}

	//spew.Dump(actions)
	//c.Map.Send()
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
	//c.Map.Send()
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
	c.WasHit = make(map[int]bool)
	c.EnemyLocations = make([]*client.Position, 0)
	c.OurCount = len(msg.You.Bots)
	c.EnemyCount = len(msg.You.Bots) * len(msg.OtherTeams)

	startPoints := c.Map.GetStartPoints(len(msg.You.Bots))

	for i := 0; i < len(msg.You.Bots); i++ {
		c.Map.SetMyBot(&msg.You.Bots[i], startPoints[i])
		c.Actions[msg.You.Bots[i].BotId] = &client.Action{BotId: msg.You.Bots[i].BotId}
		c.WasLocated[msg.You.Bots[i].BotId] = false
	}
	//c.Map.Send()
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
	c.EnemyLocations = make([]*client.Position, 0)

	for _, e := range msg.Events {
		switch e.Type {

		// This can happen to our or enemy bot so lets use damaged to detect hits on our bots
		case client.EVENT_HIT:
			log.Printf("[corba][OnEvents][[hit] : Bot %d\n", e.BotId.Int64)
			break

		case client.EVENT_DIE:
			log.Printf("[corba][OnEvents][die] : Bot %d\n", e.BotId.Int64)
			if _, ok := c.Actions[int(e.BotId.Int64)]; ok {
				// Bot was ours so remove bot from data structures
				delete(c.Actions, int(e.BotId.Int64))
				delete(c.WasLocated, int(e.BotId.Int64))
				delete(c.WasHit, int(e.BotId.Int64))
				c.OurCount--
			} else {
				c.EnemyCount--
			}
			break

		case client.EVENT_RADAR_ECHO:
			log.Printf("[corba][OnEvents][radarEcho] : Pos %v\n", e.Position)
			c.Map.DetectEnemyBot(int(e.BotId.Int64), e.Position)
			c.EnemyLocations = append(c.EnemyLocations, &client.Position{e.Position.X, e.Position.Y})
			break

		// This will happen when too bots see each other
		case client.EVENT_SEE:
			log.Printf("[corba][OnEvents][see] : Bot %d Source %d Pos %v\n", e.BotId.Int64, e.Source.Int64, e.Position)
			c.Map.DetectEnemyBot(int(e.BotId.Int64), e.Position)
			c.EnemyLocations = append(c.EnemyLocations, &client.Position{e.Position.X, e.Position.Y})
			break

		case client.EVENT_DETECTED:
			log.Printf("[corba][OnEvents][detected] : Bot %d\n", e.BotId.Int64)
			c.WasLocated[int(e.BotId.Int64)] = true
			break

		case client.EVENT_DAMAGED:
			log.Printf("[corba][OnEvents][damaged] : Bot %d damage %d\n", e.BotId.Int64, e.Damage.Int64)
			c.WasLocated[int(e.BotId.Int64)] = true
			c.WasHit[int(e.BotId.Int64)] = true
			c.Map.HitBot(int(e.BotId.Int64), int(e.Damage.Int64))
			break

		case client.EVENT_MOVE:
			//log.Printf("[corba][OnEvents][move] : Bot %d\n", e.BotId.Int64)
			c.Map.MoveMyBot(int(e.BotId.Int64), e.Position)
			delete(stay, int(e.BotId.Int64))
			break

		case client.EVENT_NOACTION:
			//log.Printf("[corba][OnEvents][noaction] : Bot %d\n", e.BotId.Int64)
			c.Map.Stay(int(e.BotId.Int64))
			break

		case client.EVENT_END:
			log.Printf("[corba][OnEvents][end]\n")
			break

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
