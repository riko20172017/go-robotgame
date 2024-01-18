package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-robotgame/gameloop"

	"github.com/adriancable/webtransport-go"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Data struct {
	x string
	y string
}

func main() {
	connectionChannel := make(chan *webtransport.Session)
	dataChannel := make(chan []byte)
	http.HandleFunc("/counter", func(rw http.ResponseWriter, r *http.Request) {
		session := r.Body.(*webtransport.Session)
		session.AcceptSession()
		connectionChannel <- session
		// session.RejectSession(400)

		fmt.Println("Accepted incoming WebTransport session")

		go func() {
			for {
				msg, err := session.ReceiveMessage(session.Context())
				if err != nil {
					fmt.Println("Session closed, ending datagram listener:", err)
					break
				}
				fmt.Printf("Received datagram: %s\n", msg)

				sendMsg := bytes.ToUpper(msg)
				// fmt.Printf("Sending datagram: %s\n", sendMsg)
				dataChannel <- sendMsg
				// session.SendMessage(sendMsg)
			}
		}()
	})

	g := gameloop.New(time.Second/1, connectionChannel, dataChannel, func(delta float64) {
		// log.Println(delta)
	})

	g.Start()

	// Note: "new-tab-page" in AllowedOrigins lets you access the server from a blank tab (via DevTools Console).
	// "" in AllowedOrigins lets you access the server from JavaScript loaded from disk (i.e. via a file:// URL)
	server := &webtransport.Server{
		ListenAddr:     ":4433",
		TLSCert:        webtransport.CertFile{Path: "localhost.pem"},
		TLSKey:         webtransport.CertFile{Path: "localhost-key.pem"},
		AllowedOrigins: []string{"googlechrome.github.io", "127.0.0.1:8000", "localhost:8000", "new-tab-page", ""},
		QuicConfig: &webtransport.QuicConfig{
			KeepAlive:      true,
			MaxIdleTimeout: 30 * time.Second,
		},
	}

	fmt.Println("Launching WebTransport server at", server.ListenAddr)
	ctx, cancel := context.WithCancel(context.Background())
	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
		cancel()
	}
}
