// Package gameloop implements a very simple game loop.
// This code is based on: github.com/kutase/go-gameloop
package gameloop

import (
	"fmt"
	"runtime"
	"time"

	"github.com/adriancable/webtransport-go"
)

// GameLoop implements a simple game loop.
type GameLoop struct {
	onUpdate  func(float64) // update function called by loop
	tickRate  time.Duration // tick interval
	Quit      chan bool
	myChannel chan *webtransport.Session // channel used for exiting the loop
	entities  []Entity
}

type Entity struct {
	x       float32
	y       float32
	session *webtransport.Session
}

// Create new game loop
func New(tickRate time.Duration, mychannel chan *webtransport.Session, onUpdate func(float64)) *GameLoop {
	return &GameLoop{
		onUpdate:  onUpdate,
		tickRate:  tickRate,
		Quit:      make(chan bool),
		myChannel: mychannel,
	}
}

// startLoop sets up and runs the loop until we exit.
func (g *GameLoop) startLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Set up ticker.
	t := time.NewTicker(g.tickRate)

	var now int64
	var delta float64
	start := time.Now().UnixNano()

	i := 0

	for {
		select {
		case <-t.C:
			// Calculate delta T in fractions of seconds.
			now = time.Now().UnixNano()
			delta = float64(now-start) / 1000000000
			start = now
			g.onUpdate(delta)
			for _, s := range g.entities {
				s.session.SendMessage([]byte(fmt.Sprintf("%2d", i)))
			}
		case s := <-g.myChannel:
			g.entities = append(g.entities, Entity{x: 0.0, y: 0.0, session: s})

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
