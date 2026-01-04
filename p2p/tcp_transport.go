package p2p

import (
	"encoding/gob"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

type Peer struct {
	conn 		net.Conn
	outbound 	bool 
	listenAddr 	string 
	encoderLock sync.Mutex
	encoder 	*gob.Encoder
}

func (p *Peer) Send(msg *Message) error {
	p.encoderLock.Lock()
	defer p.encoderLock.Unlock()
	return p.encoder.Encode(msg)
}

func (p *Peer) ReadLoop(msgch chan *Message, delPeerch chan *Peer){
	decoder := gob.NewDecoder(p.conn)
	for {
		msg := new(Message)
		if err := decoder.Decode(msg); err != nil {
			logrus.Errorf("Peer %s: decode message error: %s", p.listenAddr, err)
			break 
		}
		msgch <- msg 
	}
	delPeerch <- p
	logrus.Infof("Peer %s connection closed.", p.listenAddr)
	p.conn.Close()
}

type TCPTransport struct {
	listenAddr string 
	listener net.Listener
	AddPeer chan *Peer 
	DelPeer chan *Peer 
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err 
	}
	t.listener = ln 
	logrus.Infof("TCP Transport listening on %s", t.listenAddr)

	for {
		conn, err := ln.Accept()
		if err !=  nil {
			logrus.Error(err)
			continue 
		}
		peer := &Peer{
			conn: conn,
			outbound: false,
		}
		t.AddPeer <- peer 
	}
}