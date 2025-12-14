package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultMaxPlayers = 6
	handshakeTimeout = 3 * time.Second
)

type GameVariant uint8

const (
	TexasHoldem GameVariant = iota 
	Other 
)

func (gv GameVariant) String() string {
	switch gv {
		case TexasHoldem:
			return "TEXAS-HOLDEM"
		case Other:
			return "OTHER"
		default:
			return "INVALID"
	}
}

type ServerConfig struct {
	Version string 
	ListenAddr string 
	APIListenAddr string 
	GameVariant GameVariant 
	MaxPlayers int 
}

type Server struct {
	ServerConfig
	transport *TCPTransport
	peerLock sync.RWMutex
	peers map[string]*Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgch chan *Message
	broadcastch chan BroadcastTo
	// gameState *Game
}

func NewServer(cfg ServerConfig) *Server {
	if cfg.MaxPlayers == 0{
		cfg.MaxPlayers = defaultMaxPlayers
	}
	s := &Server{
		ServerConfig: cfg,
		peers: make(map[string]*Peer),
		addPeer: make(chan *Peer, 10),
		delPeer: make(chan *Peer, 10),
		msgch: make(chan *Message, 100),
		broadcastch: make(chan BroadcastTo, 100),
	}
	// s.gameState = NewGame(s.listenAdr, s.broadcastch)
	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr 

	tr.AddPeer = s.addPeer
	tr.DelPeer = s.delPeer

	go func(s *Server){
		// apiServer := NewAPIServer(cfg.APIListenAddr, &Game{})
		// apiServer.Run()
		logrus.WithFields(logrus.Fields{
			"listenAddr": cfg.APIListenAddr,
		}).Info("API Server Placeholder running")
	}(s)
	return s
}

// func (s *Server) Start() {
// 	go s.loop()
// }

func (s *Server) AddPeer(p *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[p.listenAddr] = p
}

func (s *Server) GetPeer(addr string) (*Peer, bool){
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	peer, ok := s.peers[addr]
	return peer, ok
}

// func (s *Server) loop() {
// 	for {
// 		select {
// 		case msg := <- s.broadcastch:
// 			if err := 
// 		}
// 	}
// }

// func (s *Server) handleNewPeer(peer *Peer) error {
// 	// Returns an error if the read does not happen in the given time
// 	peer.conn.SetReadDeadline(time.Now().Add(handshakeTimeout))
// 	hs, err := s.handshake(peer)
// 	if err != nil {
// 		peer.conn.Close()
// 		return fmt.Errorf("Handshake with incoming player failed: %s", err)
// 	}
// 	// Resetting the deadline after handshake
// 	peer.conn.SetReadDeadline(time.Time{})
// 	peer.listenAddr = hs.ListenAddr
	
// }

func (s *Server) Broadcast(broadcastMsg BroadcastTo) error {
	msg := NewMessage(s.ListenAddr, broadcastMsg.Payload)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, addr := range broadcastMsg.To {
		peer, ok := s.GetPeer(addr)
		if ok {
			go func(peer *Peer){
				if err := peer.Send(buf.Bytes()); err != nil {
					logrus.Errorf("Broadcast to %s error: %s", addr, err)
				}
			}(peer)
		}
	}
	return nil
}

func (s *Server) handshake(p *Peer) (*Handshake, error) {
	if len(s.peers) >= s.MaxPlayers {
		return nil, fmt.Errorf("max players exceeded (%d)", s.MaxPlayers)
	}
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return nil, err
	}
	if s.GameVariant != hs.GameVariant {
		return nil, fmt.Errorf("gamevariant mismatch: want %s but got %s", s.GameVariant, hs.GameVariant)
	}
	if s.Version != hs.Version{
		return nil, fmt.Errorf("invalid version: want %s but got %s", s.Version, hs.Version)
	}
	return hs, nil
}