package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
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

func (s *Server) Start() {
	go s.loop()
	logrus.WithFields(logrus.Fields{
		"p2p-port": s.ListenAddr,
		"variant": s.GameVariant,
		"maxPlayers": s.MaxPlayers,
	}).Info("Staring P2P game server...")
	s.transport.ListenAndAccept()
}

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

func (s *Server) Peers() []string {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	peers := make([]string, 0, len(s.peers))
	for addr := range s.peers {
		peers = append(peers, addr)
	}
	return peers
}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		GameVariant: s.GameVariant,
		Version: s.Version,
		ListenAddr: s.ListenAddr,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}
	return p.Send(buf.Bytes())
}

func (s *Server) Connect(addr string) error {
	if _, ok := s.GetPeer(addr); ok {
		logrus.Warnf("already connected to peer %s", addr)
		return nil
	}
	conn, err := net.DialTimeout("tcp", addr, handshakeTimeout)
	if err != nil {
		return err
	}
	// For outbound connections the listenaddr is not know yet
	// We will set it after the handshake is received
	peer := &Peer{
		conn: conn,
		outbound: true,
	}
	s.addPeer <- peer
	return s.SendHandshake(peer)
} 

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.broadcastch:
			if err := s.Broadcast(msg); err != nil {
				logrus.Errorf("broadcast error: %s", err)
			}
		case peer := <-s.delPeer:
			s.handleDelPeer(peer)
		case peer := <-s.addPeer:
			if err := s.handleNewPeer(peer); err != nil {
				logrus.Errorf("handle new peer error: %s", err)
			}
		case msg := <-s.msgch:
			go func(){
				if err := s.handleMessage(msg); err != nil {
					logrus.Errorf("message handler error: %s", err)
				}
			}()
		}
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	peer.conn.SetReadDeadline(time.Now().Add(handshakeTimeout))
	hs, err := s.handshake(peer)
	if err != nil {
		peer.conn.Close()
		return fmt.Errorf("handshake with incoming player failed: %s", err)
	}
	peer.conn.SetReadDeadline(time.Time{})
	peer.listenAddr = hs.ListenAddr
	s.AddPeer(peer)

	go peer.ReadLoop(s.msgch, s.delPeer)

	if !peer.outbound {
		if err := s.SendHandshake(peer); err !=  nil {
			peer.conn.Close()
			s.handleDelPeer(peer)
			return fmt.Errorf("failed to send handshake to peer: %s", err)
		}
		go func(){
			if err := s.sendPeerList(peer); err != nil{
				logrus.Errorf("error sending peer list: %s", err)
			}
		}()
	}
	logrus.WithFields(logrus.Fields{
		"peer_addr": peer.listenAddr,
		"outbound": peer.outbound,
		"version": hs.Version,
	})
	// s.gameState.AddPlayer(peer.listenAddr)
	return nil

 }

func (s *Server) handleDelPeer(peer *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	addrToDelete := peer.listenAddr
	// if addrToDelete == ""{
	// 	addrToDelete = peer.conn.RemoteAddr().String()
	// }
	delete(s.peers, addrToDelete)
	logrus.WithFields(logrus.Fields{
		"addr": addrToDelete,
	}).Info("Peer disconnected and removed")
	// s.gameState.RemovePlayer(addrToDelete)
}

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

func (s *Server) sendPeerList(p *Peer) error {
	peerListMsg := MessagePeerList{
		Peers: s.Peers(),
	}
	peers := []string{}
	for _, addr := range peerListMsg.Peers {
		if addr != p.listenAddr {
			peers = append(peers, addr)
		}
	}
	peerListMsg.Peers = peers
	if len(peerListMsg.Peers) == 0{
		return nil
	}
	msg := NewMessage(s.ListenAddr, peerListMsg)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	return p.Send(buf.Bytes())
}

func (s *Server) handlePeerList(msg MessagePeerList) error {
	logrus.WithFields(logrus.Fields{
		"we": s.ListenAddr,
		"list-size": len(msg.Peers),
	}).Info("received peer list message. Checking for new peers...")
	for _, addr := range msg.Peers{
		if _, ok := s.GetPeer(addr); !ok {
			if err := s.Connect(addr); err != nil {
				logrus.Errorf("Failed to dial peer %s: %s", addr, err)
				continue
			}
		}
	}
	return nil
}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
		case MessagePeerList:
			return s.handlePeerList(v)
		case MessageEncDeck:
			// return s.gameState.ShuffleAndEncrypt(msg.From, v.Deck)
			logrus.Infof("Received encrypted deck from %s", msg.From)
		case MessageReady:
			// return s.gameState.SetPlayerReady(msg.From)
			logrus.Infof("Received ready status from %s", msg.From)
		case MessagePreFlop:
			// return s.gameState.SetStatus(GameStatusPreFlop)
			logrus.Infof("Received preflop signal from %s", msg.From)
		case MessagePlayerAction:
			// return s.gameState.handlePlayerAction(msg.From, v)
			logrus.Infof("Recieved player action from %s: %s value %s", msg.From, v.Action, v.Value)
		default:
			logrus.Warnf("Received unhandled message type from %s", msg.From)
	}
	return nil
}

func init() {
	gob.Register(Handshake{})
	gob.Register(MessagePeerList{})
	gob.Register(MessagePlayerAction{})
	gob.Register(MessageReady{})
	gob.Register(MessageEncDeck{})
	gob.Register(MessagePreFlop{})
}