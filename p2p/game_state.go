package p2p

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type PlayersList struct {
	lock sync.RWMutex
	list []string
}

func NewPlayersList() *PlayersList {
	return &PlayersList{list: []string{}}
}

func (p *PlayersList) Len() int {return len(p.list)}

func (p *PlayersList) Swap(i, j int) {
	p.list[i], p.list[j] = p.list[j], p.list[i]
}

func (p *PlayersList) Less(i, j int) bool {
	return p.list[i] < p.list[j]
}

func (p *PlayersList) List() []string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	listCopy := make([]string, len(p.list))
	copy(listCopy, p.list)
	return listCopy
}

func (p *PlayersList) len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return len(p.list)
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
	ListenAddr    	string 
	RotationID 		int 
	IsReady 		bool 
	IsActive 		bool 
	IsFolded 		bool 
	CurrentRoundBet int
}

type Game struct {
	lock 				sync.RWMutex
	listenAddr 			string 
	broadcastch 		chan BroadcastTo
	playersList 		*PlayersList
	currentStatus 		*AtomicInt
	currentPot 			int 
	playerStates 		map[string]*PlayerState
	rotationMap 		map[int]string 
	nextRotationID 		int 
	currentDealerID 	int 
	currentPlayerTurnID int 
	highestBet 			int 
	lastRaiserID 		int 
}

func NewGame(addr string, bc chan BroadcastTo) *Game {
	g := &Game{
		playersList: 	NewPlayersList(),
		broadcastch: 	bc,
		listenAddr: 	addr,
		currentStatus: 	NewAtomicInt(int32(GameStatusWaiting)),
		playerStates: 	make(map[string]*PlayerState),
		rotationMap: 	make(map[int]string),
	}
	g.playersList.add(addr)
	g.playerStates[addr] = &PlayerState{ListenAddr: addr, IsActive: true}

	go g.loop()
	return g
}

func (g *Game) AddPlayer(addr string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, exists := g.playerStates[addr]; exists {
		g.playerStates[addr].IsActive = true 
		return 
	}
	g.playersList.add(addr)
	g.playerStates[addr] = &PlayerState{ListenAddr: addr, IsActive: true}
}

func (g *Game) RemovePlayer(addr string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if state, ok := g.playerStates[addr]; ok {
		state.IsActive = false 
		state.IsFolded = true 
		g.playersList.remove(addr)
	}
}

func (g *Game) SetReady(from string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	state, ok := g.playerStates[from]
	if !ok {
		return 
	}
	if !state.IsReady{
		state.RotationID = g.nextRotationID
		g.rotationMap[state.RotationID] = from 
		g.nextRotationID++
		state.IsReady = true 
	}

	// g.sendToPlayers(MessageReady{}. g.getOtherPlayers()...)
	// if len(g.getReadyPlayers()) >= 2 && GameStatus(g.currentStatus.Get()) == GameStatusWaiting {
	// 	g.startNewHand()
	// }
}

// func (g *Game) StartNewHand() {
// 	g.lock.Lock()
// 	defer g.lock.Unlock()
// }

func (g *Game) advanceDealer() {
	if g.nextRotationID == 0 {
		g.currentDealerID = 0
		return 
	}
	startID := g.currentDealerID
	for {
		nextID := (startID + 1) % g.nextRotationID
		addr, ok := g.rotationMap[nextID]
		if ok && g.playerStates[addr].IsActive{
			g.currentDealerID = nextID 
			// g.currentPlayerTurnID = g.getNextPlayerID(g.currentDealerID)
			return 
		}
		startID = nextID
		if startID == g.currentDealerID {
			break
		}
	}
}

// func (g *Game) TakeAction(action PlayerAction, value int) error {
// 	g.lock.Lock()
// 	defer g.lock.Unlock()

// 	if g.playerStates[g.listenAddr].RotationID != g.currentPlayerTurnID {
// 		return fmt.Errorf("it is not my turn to act: %s", g.listenAddr)
// 	}
// }

// func (g *Game) handlePlayerAction(from string, msg MessagePlayerAction) error {
// 	g.lock.Lock()
// 	defer g.lock.Unlock()

// 	if g.playerStates[from].RotationID != g.currentPlayerTurnID {
// 		return fmt.Errorf("player (%s) acting out of turn", from)
// 	}

// }

func (g *Game) updatePlayerState(addr string, action PlayerAction, value int) {
	state := g.playerStates[addr]
	switch action {

	case PlayerActionFold:
		state.IsFolded = true
	case PlayerActionBet, PlayerActionRaise:
		state.CurrentRoundBet += value 
		g.currentPot += value 
		if state.CurrentRoundBet > g.highestBet {
			g.highestBet = state.CurrentRoundBet
			g.lastRaiserID = state.RotationID
		}
	case PlayerActionCall:
		amountToCall := g.highestBet - state.CurrentRoundBet
		state.CurrentRoundBet += amountToCall
		g.currentPot += amountToCall
	case PlayerActionCheck:
		// no change in state
	}
}

// func (g *Game) advanceTurnAndCheckRoundEnd() {
	
// }

// func (g *Game) incNextPlayer(){
// 	startID := g.currentPlayerTurnID
// 	for {
// 		nextID := g.get
// 	}
// }

func (g *Game) getNextPlayerID(currentID int) int {
	return (currentID + 1) & g.nextRotationID
}

func (g *Game) getReadyActivePlayers() []string {
	active := []string{}
	for _, state := range g.playerStates {
		if state.IsReady && state.IsActive {
			active = append(active, state.ListenAddr)
		}
	}
	return active
}

func (g *Game) getReadyPlayers() []string {
	ready := []string{}
	for _, state := range g.playerStates {
		if state.IsReady {
			ready = append(ready, state.ListenAddr)
		}
	}
	return ready
}

func (g *Game) setStatus(s GameStatus) {
	g.currentStatus.Set(int32(s))
}

func (g *Game) getNextGameStatus() GameStatus {
	switch GameStatus(g.currentStatus.Get()){
	case GameStatusPreFlop: return GameStatusFlop
	case GameStatusFlop: return GameStatusTurn
	case GameStatusTurn: return GameStatusRiver
	case GameStatusRiver: return GameStatusShowdown
	default: return GameStatusHandComplete
	}
}

func (g *Game) sendToPlayers(payload any, addr ...string){
	g.broadcastch <- BroadcastTo{
		To: addr,
		Payload: payload,
	}
}

// func (g *Game) getOtherPlayers() []string {
// 	players := []string{}
// 	for _, addr := range g.playersList.List(){

// 	}
// }

func (g *Game) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<-ticker.C
		dealerAddr := g.rotationMap[g.currentDealerID]
		turnAddr := g.rotationMap[g.currentPlayerTurnID]
		logrus.WithFields(logrus.Fields{
			"status": GameStatus(g.currentStatus.Get()),
			"dealer-ID": g.currentDealerID,
			"dealer-addr": dealerAddr,
			"turn-ID": g.currentPlayerTurnID,
			"turn-player": turnAddr,
			"highest-bet": g.highestBet,
			"last-raised-ID": g.lastRaiserID,
			"rotation-size": g.nextRotationID,
			// "ready-active": len(g.getReadyActivePlayers()),
		}).Info("Game State Heartbeat")
	}
}