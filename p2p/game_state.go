package p2p

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

type PlayersList struct {
	lock sync.RWMutex
	list []string
}

func NewPlayersList() *PlayersList {
	return &PlayersList{list: []string{}}
}

func (p *PlayersList) add(addr string){
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, existingAddr := range p.list {
		if existingAddr == addr {
			return 
		}
	}
	p.list = append(p.list, addr)
	sort.Sort(p)
}

func (p *PlayersList) remove(addr string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for i, existingAddr := range p.list {
		if existingAddr == addr {
			p.list = append(p.list[:i], p.list[i+1:]...)
			return
		}
	}
}

func (p *PlayersList) get(index int) string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if len(p.list) - 1 < index || index < 0 {
		return ""
	}
	return p.list[index]
}

func (p *PlayersList) Len() int {return len(p.list)}

func (p *PlayersList) Swap(i, j int) {
	p.list[i], p.list[j] = p.list[j], p.list[i]
}

func (p *PlayersList) Less(i, j int) bool {
	return p.list[i] < p.list[j]
}

type AtomicInt struct {
	value int32
}

func NewAtomicInt(value int32) *AtomicInt{
	return &AtomicInt{value: value}
}

func (a *AtomicInt) String() string {return fmt.Sprintf("%d", a.Get())}
func (a *AtomicInt) Get() int32 {return atomic.LoadInt32(&a.value)}
func (a *AtomicInt) Set(value int32) {atomic.StoreInt32(&a.value, value)}
func (a *AtomicInt) Inc() {a.Set(a.Get()+1)}

type PlayerState struct {
	ListenAddr string 
	RotationID int 
	IsReady bool 
	IsActive bool 
	IsFolded bool 
	CurrentRoundBet int
}

type Game struct {
	lock sync.RWMutex
	listenAddr string 
	broadcastch chan BroadcastTo
	playersList *PlayersList
	
}