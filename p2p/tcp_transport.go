package p2p

import "net"

type Peer struct {
	conn net.Conn
	outbound bool 
	listenAddr string 
}

func (p *Peer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err 
}

// func (p *Peer) ReadLoop(msgch chan *Message){
// 	for 
// }