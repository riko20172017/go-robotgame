// Package gameloop implements a very simple game loop.
// This code is based on: github.com/kutase/go-gameloop
package gameloop

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/adriancable/webtransport-go"
)

// GameLoop implements a simple game loop.
type GameLoop struct {
	onUpdate          func(float32) // update function called by loop
	tickRate          time.Duration // tick interval
	Quit              chan bool
	connectionChannel chan *webtransport.Session // channel used for exiting the loop
	dataChannel       chan []byte
	entities          map[int8]Entity
	sessions          map[int8]Session
	messages          []Data
}

type Entity struct {
	uid     int8
	x       float32
	y       float32
	session *webtransport.Session
}

type Data struct {
	Uid   int8    `json:"uid"`
	Tik   int     `json:"tik"`
	deltaTime float64 `json:"deltaTime"`
	Keys  Keys    `json:"keys"`
}

type Keys struct {
	Space int   `json:"space"`
	Left  int   `json:"left"`
	Up    int   `json:"up"`
	Right int   `json:"right"`
	Down  int   `json:"down"`
	Mouse Mouse `json:"mouse"`
}

type Mouse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Session struct {
	session *webtransport.Session
}

type Command struct {
	Type string `json:"type"`
}

type Req struct {
	Type string `json:"type"`
	Id   int    `json:"id"`
}

type Messages struct {
}

// Create new game loop
func New(tickRate time.Duration, connectionChannel chan *webtransport.Session, dataChannel chan []byte, onUpdate func(float32)) *GameLoop {
	return &GameLoop{
		onUpdate:          onUpdate,
		tickRate:          tickRate,
		Quit:              make(chan bool),
		connectionChannel: connectionChannel,
		dataChannel:       dataChannel,
		entities:          make(map[int8]Entity),
		sessions:          make(map[int8]Session),
		messages:          make([]Data, 0),
	}
}

// startLoop sets up and runs the loop until we exit.
func (g *GameLoop) startLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Set up ticker.
	t := time.NewTicker(g.tickRate)

	var uid int8 = 0
	var now int64
	var deltaTime float32
	start := time.Now().UnixNano()

	for {
		select {
		case <-t.C:
			// Calculate deltaTime T in fractions of seconds.
			now = time.Now().UnixNano() 
			deltaTime = float32(now-start) / 1000000000
			start = now
			g.onUpdate(deltaTime)  
			// fmt.Printf("%+v\n", g.messages)
			for _, m := range g.messages {
			// fmt.Printf("%+v", m)
			// print(len(g.messages))
			
				index := m.Uid
				entity := g.entities[index]
				deltaMove := deltaTime * 200

				if (m.Keys.Down == 1) {
					entity.y += deltaMove;
				}
				if (m.Keys.Up == 1) {
					entity.y -= deltaMove;
				}
				if (m.Keys.Left == 1) {
					entity.x -= deltaMove;
				}
				if (m.Keys.Right == 1) {
					entity.x += deltaMove;
				}

				g.entities[index] = entity
				g.entities[index].session.SendMessage(
					[]byte(fmt.Sprintf(
						"{\"type\":\"DATA\", \"states\": [{\"uid\":%d,\"x\":%f,\"y\":%f}]}", entity.uid, entity.x, entity.y)))
			}
			g.messages = g.messages[:0]
		case s := <-g.connectionChannel: 
			uid = uid + 1
			g.sessions[uid] = Session{session: s}
			s.SendMessage([]byte(fmt.Sprintf("{\"type\": \"OFFER\",\"uid\": %d}", uid)))

		case d := <-g.dataChannel:
			command := Command{}
			err := json.Unmarshal(d, &command)
			if err != nil {
				// Используем Fatal только для примера,
				// нельзя использовать в реальных приложениях
				log.Fatalln("unmarshal ", err.Error())
			}

			switch command.Type {
			case "DATA":
				data := Data{}
				err := json.Unmarshal(d, &data)
				if err != nil {
					// Используем Fatal только для примера,
					// нельзя использовать в реальных приложениях
					log.Fatalln("unmarshal ", err.Error())
				}
				// fmt.Printf("%+v", g.entities[int8(data.Uid)])
				g.messages = append(g.messages, data)

			case "REQUEST":
				respons := Req{}
				json.Unmarshal(d, &respons)
				session := g.sessions[int8(respons.Id)].session
				uid := int8(respons.Id)
				g.entities[uid] = Entity{uid: uid, x: 0.0, y: 0.0, session: session}
				session.SendMessage([]byte(fmt.Sprintf("{\"type\": \"JOIN\"}")))
				// fmt.Printf("%+v", g.messages)
			}

		case <-g.Quit:
			t.Stop()
		}
	}
}

// Start game loop.
func (g *GameLoop) Start() {
	go g.startLoop()
}

// Stop game loop.
func (g *GameLoop) Stop() {
	g.Quit <- true
}

// Restart game loop.
func (g *GameLoop) Restart() {
	g.Stop()
	g.Start()
}
