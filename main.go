package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RedPaladin7/peerpoker/p2p"
	"github.com/sirupsen/logrus"
)

const (
	defaultVersion = "1.0.0"
	defaultP2PPort = "3000"
	defaultAPIPort = "8080"
)

func main() {
	var (
		p2pPort = flag.String("p2p-port", defaultP2PPort, "P2P network port")
		apiPort = flag.String("api-port", defaultAPIPort, "HTTP API port")
		connectTo = flag.String("connect", "", "Connect to existing peer (e.g., localhost: 3000)")
		maxPlayers = flag.Int("max-players", 6, "Maximum number of players")
		logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		version = flag.Bool("version", false, "Print version and exit")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Decentralized Poker Engine v%s\n", defaultVersion)
		os.Exit(0)
	}

	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatalf("Invalid log level: %s", *logLevel)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	p2pAddr := fmt.Sprintf("localhost:%s", *p2pPort)
	apiAddr := fmt.Sprintf("localhost:%s", *apiPort)

	cfg := p2p.ServerConfig{
		Version: defaultVersion,
		ListenAddr: p2pAddr,
		APIListenAddr: apiAddr,
		MaxPlayers: *maxPlayers,
		GameVariant: p2p.TexasHoldem,
	}

	server := p2p.NewServer(cfg)

	if *connectTo != "" {
		logrus.Infof("Connecting to peer: %s", *connectTo)
		go func() {
			if err := server.Connect(*connectTo); err != nil {
				logrus.Errorf("Failed to connect to peer %s: %s", *connectTo, err)
			} else {
				logrus.Infof("Successfully connected to peer %s", *connectTo)
			}
		}()
	}

	logrus.Info("===========================================")
	logrus.Info("  Decentralized Poker Engine")
	logrus.Info("===========================================")
	logrus.Infof("Version:        %s", defaultVersion)
	logrus.Infof("P2P Address:    %s", p2pAddr)
	logrus.Infof("API Address:    http://%s", apiAddr)
	logrus.Infof("Game Variant:   %s", cfg.GameVariant)
	logrus.Infof("Max Players:    %d", *maxPlayers)
	logrus.Info("===========================================")
	logrus.Info("")
	logrus.Info("API Endpoints:")
	logrus.Infof("  Health:       GET  http://%s/api/health", apiAddr)
	logrus.Infof("  Table State:  GET  http://%s/api/table", apiAddr)
	logrus.Infof("  Players:      GET  http://%s/api/players", apiAddr)
	logrus.Infof("  Ready:        POST http://%s/api/ready", apiAddr)
	logrus.Infof("  Fold:         POST http://%s/api/fold", apiAddr)
	logrus.Infof("  Check:        POST http://%s/api/check", apiAddr)
	logrus.Infof("  Call:         POST http://%s/api/call", apiAddr)
	logrus.Infof("  Bet:          POST http://%s/api/bet", apiAddr)
	logrus.Infof("  Raise:        POST http://%s/api/raise", apiAddr)
	logrus.Info("===========================================")
	logrus.Info("")

	if *connectTo == "" {
		logrus.Info("💡 Starting as initial node. To connect other players, run:")
		logrus.Infof("   go run main.go -p2p-port=3001 -api-port=8081 -connect=%s", p2pAddr)
	}

	logrus.Info("")
	logrus.Info("🎮 Server starting... Press Ctrl+C to stop")
	logrus.Info("")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go server.Start()

	<-sigChan
	logrus.Info("")
	logrus.Info("🛑 Shutdown signal received. Cleaning up...")
	
	// TODO: Add graceful shutdown logic here
	// - Save game state
	// - Notify peers
	// - Close connections

	logrus.Info("✅ Server stopped successfully")
}