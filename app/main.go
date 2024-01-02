package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/adriancable/webtransport-go"
)

func handleWebTransportStreams(session *webtransport.Session) {
	// Handle incoming datagrams
	go func() {
		for {
			msg, err := session.ReceiveMessage(session.Context())
			if err != nil {
				fmt.Println("Session closed, ending datagram listener:", err)
				break
			}
			fmt.Printf("Received datagram: %s\n", msg)

			sendMsg := bytes.ToUpper(msg)
			fmt.Printf("Sending datagram: %s\n", sendMsg)
			session.SendMessage(sendMsg)
		}
	}()
}

func main() {
	http.HandleFunc("/counter", func(rw http.ResponseWriter, r *http.Request) {
		session := r.Body.(*webtransport.Session)
		session.AcceptSession()
		// session.RejectSession(400)

		fmt.Println("Accepted incoming WebTransport session")
		handleWebTransportStreams(session)
	})

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
