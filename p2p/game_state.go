package p2p

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	SmallBlind = 10
	BigBlind = 20
)

// #####################################
// UTILITY - STRUCTURES
// #####################################

type PlayerHand struct {
	Addr 		string 
	Hand 		[]Card 
	Rank 		int32 
	HandName 	string
}

type SidePot struct {
	Amount 			int 
	EligiblePlayers []string
}

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

type AtomicInt struct {
	value int32
}

func NewAtomicInt(value int32) *AtomicInt{
	return &AtomicInt{value: value}
}

func (a *AtomicInt) Get() int32 {return atomic.LoadInt32(&a.value)}
func (a *AtomicInt) Set(value int32) {atomic.StoreInt32(&a.value, value)}	

// #####################################
// GAME - LOGIC
// #####################################


type PlayerState struct {
	ListenAddr    	string 
	RotationID 		int 
	IsReady 		bool 
	IsActive 		bool 
	IsFolded 		bool 
	CurrentRoundBet int
	IsAllIn 		bool 
	Stack 			int
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
	deckKeys 			*CardKeys
	foldedPlayerKeys 	map[string]*CardKeys
	revealedKeys 		map[string]*CardKeys
	currentDeck 		[][]byte
	myHand 				[]Card
	communityCards 		[]Card
	sidePots 			[]SidePot
}

func NewGame(addr string, bc chan BroadcastTo) *Game {
	sharedPrime, _ := new(big.Int).SetString("C7970CEDCC5226685694605929849D3D", 16)
	keys, _ := GenerateCardKeys(sharedPrime)
	g := &Game{
		playersList: 			NewPlayersList(),
		broadcastch: 			bc,
		listenAddr: 			addr,
		currentStatus: 			NewAtomicInt(int32(GameStatusWaiting)),
		playerStates: 			make(map[string]*PlayerState),
		rotationMap: 			make(map[int]string),
		deckKeys: 				keys,
		foldedPlayerKeys: 		make(map[string]*CardKeys),
		revealedKeys: 			make(map[string]*CardKeys),
		myHand: 				make([]Card, 0, 2),
		communityCards: 		make([]Card, 0, 5),		
		sidePots:				[]SidePot{},	
	}
	g.playersList.add(addr)
	g.playerStates[addr] = &PlayerState{
		ListenAddr: addr, 
		IsActive: true, 
		Stack: 1000,
	}

	go g.loop()
	return g
}

func (g *Game) postBlinds() {
	activeCount := len(g.getReadyActivePlayers())
	
	if activeCount == 2 {
		sbID := g.currentDealerID 
		sbAddr := g.rotationMap[sbID]
		g.updatePlayerState(sbAddr, PlayerActionBet, SmallBlind)
		logrus.Infof("Player %s (dealer) posted small blind: %d", sbAddr, SmallBlind)

		bbID := g.getNextActivePlayerID(sbID)
		bbAddr := g.rotationMap[bbID]
		g.updatePlayerState(bbAddr, PlayerActionBet, BigBlind)
		logrus.Infof("Player %s posted big blind: %d", bbAddr, BigBlind)
		g.currentPlayerTurnID = sbID
		g.lastRaiserID = bbID
	} else {
		sbID := g.getNextActivePlayerID(g.currentDealerID)
		sbAddr := g.rotationMap[sbID]
		g.updatePlayerState(sbAddr, PlayerActionBet, SmallBlind)
		logrus.Infof("Player %s posted small blind: %d", sbAddr, SmallBlind)

		bbID := g.getNextActivePlayerID(sbID)
		bbAddr := g.rotationMap[bbID]
		g.updatePlayerState(bbAddr, PlayerActionBet, BigBlind)
		logrus.Infof("Player %s posted big blind: %d", bbAddr, BigBlind)

		g.currentPlayerTurnID = g.getNextActivePlayerID(bbID)
		g.lastRaiserID = bbID
	}
}

func (g *Game) AddPlayer(addr string) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, exists := g.playerStates[addr]; exists {
		g.playerStates[addr].IsActive = true 
		return 
	}
	g.playersList.add(addr)
	g.playerStates[addr] = &PlayerState{
		ListenAddr: addr, 
		IsActive: true,
		Stack: 1000,
	}
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

	g.sendToPlayers(MessageReady{}, g.getOtherPlayers()...)
	if len(g.getReadyPlayers()) >= 2 && GameStatus(g.currentStatus.Get()) == GameStatusWaiting {
		g.StartNewHand()
	}
}

func (g *Game) StartNewHand() {
	g.lock.Lock()
	defer g.lock.Unlock()

	activeReadyPlayers := g.getReadyActivePlayers()
	if len(activeReadyPlayers) < 2 {
		g.setStatus(GameStatusWaiting)
		logrus.Warn("Not enough players to start a hand")
		return 
	}
	g.rotationMap = make(map[int]string)
	g.nextRotationID = 0
	g.myHand = make([]Card, 0, 2)
	g.communityCards = make([]Card, 0, 5)

	sort.Strings(activeReadyPlayers)
	for _, addr := range activeReadyPlayers{
		state := g.playerStates[addr]
		state.RotationID = g.nextRotationID
		state.IsFolded = false 
		state.CurrentRoundBet = 0
		g.rotationMap[state.RotationID] = addr 
		g.nextRotationID++
	}
	g.advanceDealer()
	g.postBlinds()
	g.currentPot = 0
	g.highestBet = 0
	g.lastRaiserID = g.currentDealerID
	g.setStatus(GameStatusDealing)
	g.InitiateShuffleAndDeal()
}

func (g *Game) advanceDealer() {
	if g.nextRotationID == 0 {
		return 
	}
	startID := g.currentDealerID
	for {
		nextID := (startID + 1) % g.nextRotationID
		addr, ok := g.rotationMap[nextID]
		if ok && g.playerStates[addr].IsActive{
			g.currentDealerID = nextID 
			return 
		}
		startID = nextID
		if startID == g.currentDealerID {
			break
		}
	}
}

// #####################################
// MENTAL - POKER - LOGIC
// #####################################

func (g *Game) InitiateShuffleAndDeal(){
	logrus.Info("Starting mental poker shuffle cycle...")
	deck := CreatePlaceHolderDeck()
	encryptedDeck := g.shuffleAndEncrypt(deck)

	nextPlayerAddr := g.rotationMap[g.getNextPlayerID(g.currentDealerID)]
	g.sendToPlayers(MessageShuffleStatus{Deck: encryptedDeck}, nextPlayerAddr)
}

func (g *Game) shuffleAndEncrypt(deck [][]byte) [][]byte  {
	newDeck := make([][]byte, len(deck))
	for i, card := range deck {
		newDeck[i] = g.deckKeys.Encrypt(card)
	}
	for i := len(newDeck) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			logrus.Errorf("crypto rand failed: %s", err)
			continue
		}
		j := jBig.Int64()
		newDeck[i], newDeck[j] = newDeck[j], newDeck[i]
	}
	return newDeck
}

func (g *Game) ShuffleAndEncrypt(from string, deck [][]byte) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.listenAddr == g.rotationMap[g.currentDealerID]{
		logrus.Info("Deck fully encrypted by all players. Starting Pre-Flop.")
		g.currentDeck = deck 
		g.setStatus(GameStatusPreFlop)
		g.sendToPlayers(MessageGameState{
			Status: GameStatusPreFlop,
			CommunityCards: []int{},
		}, g.getOtherPlayers()...)
		g.SyncState(MessageGameState{Status: GameStatusPreFlop})
		return nil
	}
	nextDeck := g.shuffleAndEncrypt(deck)
	nextPlayerAddr := g.rotationMap[g.getNextPlayerID(g.playerStates[g.listenAddr].RotationID)]
	g.sendToPlayers(MessageShuffleStatus{Deck: nextDeck}, nextPlayerAddr)
	return nil
}

func (g *Game) SyncState(msg MessageGameState) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	logrus.Infof("Syncing game state: %s", msg.Status)
	g.setStatus(msg.Status)
	g.highestBet = 0
	for _, state := range g.playerStates {
		state.CurrentRoundBet = 0
	}
	if msg.Status == GameStatusPreFlop {
		go g.revealMyHoleCards()
	}
	if len(msg.CommunityCards) > 0 {
		go g.revealCommunityCards(msg.CommunityCards)
	}
	return nil
}

func (g *Game) HandleFoldKeyReveal(from string, msg MessageRevealKeys) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.foldedPlayerKeys[from] = msg.Keys 
	logrus.Infof("Received keys from folded player %s", from)
}

func (g *Game) revealMyHoleCards() {
	indices := g.getMyHoleCardIndices()
	myID := g.playerStates[g.listenAddr].RotationID
	nextPlayerAddr := g.rotationMap[g.getNextPlayerID(myID)]

	g.sendToPlayers(MessageGetRPC{
		CardIndices: indices,
		EncryptedData: [][]byte{g.currentDeck[indices[0]], g.currentDeck[indices[1]]},
		OriginalOwner: g.listenAddr,
	}, nextPlayerAddr)
}

func (g *Game) revealCommunityCards(indices []int) {
	encryptedCards := make([][]byte, len(indices))
	for i, idx := range indices {
		encryptedCards[i] = g.currentDeck[idx]
	}
	nextPlayerAddr := g.rotationMap[g.getNextPlayerID(g.playerStates[g.listenAddr].RotationID)]
	g.sendToPlayers(MessageGetRPC{
		CardIndices: indices,
		EncryptedData: encryptedCards,
		OriginalOwner: g.listenAddr,
	}, nextPlayerAddr)
}

func (g *Game) HandleRPCRequest(from string, msg MessageGetRPC) error {
	g.lock.RLock()
	defer g.lock.RUnlock()

	decryptedData := make([][]byte, len(msg.EncryptedData))
	for i, data := range msg.EncryptedData {
		decryptedData[i] = g.deckKeys.Decrypt(data)
	}
	nextID := g.getNextPlayerID(g.playerStates[g.listenAddr].RotationID)
	nextAddr := g.rotationMap[nextID]

	if nextAddr == msg.OriginalOwner {
		g.sendToPlayers(MessageRPCResponse{
			CardIndices: msg.CardIndices,
			DecryptedData: decryptedData,
		}, msg.OriginalOwner)
	} else {
		g.sendToPlayers(MessageGetRPC{
			CardIndices: msg.CardIndices,
			EncryptedData: decryptedData,
			OriginalOwner: msg.OriginalOwner,
		},  nextAddr)
	}
	return nil
}

func (g *Game) HandleRPCResponse(from string, msg MessageRPCResponse) {
	g.lock.Lock()
	defer g.lock.Unlock()
	myIndices := g.getMyHoleCardIndices()
	for i, idx := range msg.CardIndices {
		finalBytes := g.deckKeys.Decrypt(msg.DecryptedData[i])
		card := NewCardFromByte(finalBytes[0])
		isHole := false 
		for _, myIdx := range myIndices {
			if idx == myIdx {
				isHole = true 
				break
			}
		}
		if isHole {
			g.myHand = append(g.myHand, card)
			logrus.Infof("!!! HOLE CARD REVEALED: %s !!!", card.String())
		} else {
			g.communityCards = append(g.communityCards, card)
			logrus.Infof("!!! COMMUNITY CARD REVEALED: %s !!!", card.String())
		}
	}
}

// #####################################
// SHOWDOWN - LOGIC 
// #####################################

func (g *Game) InitiateShowdown() {
	logrus.Info("!!! SHOWDOWN REACHED: Broadcasting Private Keys !!!")
	g.sendToPlayers(MessageRevealKeys{
		Keys: g.deckKeys,
	}, g.getOtherPlayers()...)
	g.lock.Lock()
	g.revealedKeys[g.listenAddr] = g.deckKeys
	g.lock.Unlock()
}

func (g *Game) ResolveWinner() {
	g.lock.Lock()
	defer g.lock.Unlock()
	
	logrus.Info("=== RESOLVING WINNER ===")
	
	activePlayers := g.getReadyActivePlayers()
	
	nonFoldedPlayers := []string{}
	for _, playerAddr := range activePlayers {
		state := g.playerStates[playerAddr]
		if !state.IsFolded {
			nonFoldedPlayers = append(nonFoldedPlayers, playerAddr)
		}
	}
	
	if len(nonFoldedPlayers) == 1 {
		winnerAddr := nonFoldedPlayers[0]
		g.playerStates[winnerAddr].Stack += g.currentPot
		logrus.Infof("🏆 WINNER BY DEFAULT: %s wins %d chips (everyone else folded)!", 
			winnerAddr, g.currentPot)
		g.currentPot = 0
		g.sidePots = []SidePot{}
		g.revealedKeys = make(map[string]*CardKeys)
		g.foldedPlayerKeys = make(map[string]*CardKeys)
		g.setStatus(GameStatusHandComplete)
		return
	}
	
	playerHands := make([]PlayerHand, 0, len(nonFoldedPlayers))
	
	for _, playerAddr := range nonFoldedPlayers {
		state := g.playerStates[playerAddr]
		
		indices := []int{state.RotationID * 2, (state.RotationID * 2) + 1}
		playerHand := []Card{}
		
		for _, cardIdx := range indices {
			encryptedCard := g.currentDeck[cardIdx]
			decryptedBytes := encryptedCard
			
			for _, keys := range g.revealedKeys {
				decryptedBytes = keys.Decrypt(decryptedBytes)
			}
		
			for _, keys := range g.foldedPlayerKeys {
				decryptedBytes = keys.Decrypt(decryptedBytes)
			}
			
			playerHand = append(playerHand, NewCardFromByte(decryptedBytes[0]))
		}
		
		rank, handName := EvaluateBestHand(playerHand, g.communityCards)
		logrus.Infof("Player %s: %v - %s (Rank: %d)", 
			playerAddr, playerHand, handName, rank)
		
		playerHands = append(playerHands, PlayerHand{
			Addr:     playerAddr,
			Hand:     playerHand,
			Rank:     rank,
			HandName: handName,
		})
	}

	if len(g.sidePots) > 0 {
		logrus.Info("Distributing side pots...")
		
		for i, pot := range g.sidePots {
			logrus.Infof("Side Pot #%d: %d chips", i+1, pot.Amount)
			
			bestRank := int32(999999)
			potWinners := []*PlayerHand{}
			
			for idx := range playerHands {
				ph := &playerHands[idx]
				isEligible := false
				for _, eligibleAddr := range pot.EligiblePlayers {
					if ph.Addr == eligibleAddr {
						isEligible = true
						break
					}
				}
				
				if isEligible {
					if ph.Rank < bestRank {
						bestRank = ph.Rank
						potWinners = []*PlayerHand{ph}
					} else if ph.Rank == bestRank {
						potWinners = append(potWinners, ph)
					}
				}
			}
			
			if len(potWinners) > 0 {
				splitAmount := pot.Amount / len(potWinners)
				remainder := pot.Amount % len(potWinners)
				
				if len(potWinners) > 1 {
					logrus.Infof("🤝 TIE in Side Pot #%d! %d players split %d chips", 
						i+1, len(potWinners), pot.Amount)
				}
				
				for j, winner := range potWinners {
					award := splitAmount
					if j == 0 {
						award += remainder
					}
					g.playerStates[winner.Addr].Stack += award
					logrus.Infof("🏆 Side Pot #%d Winner: %s receives %d chips with %s", 
						i+1, winner.Addr, award, winner.HandName)
				}
			}
		}
		
		if g.currentPot > 0 {
			bestRank := int32(999999)
			mainWinners := []*PlayerHand{}
			
			for idx := range playerHands {
				if playerHands[idx].Rank < bestRank {
					bestRank = playerHands[idx].Rank
					mainWinners = []*PlayerHand{&playerHands[idx]}
				} else if playerHands[idx].Rank == bestRank {
					mainWinners = append(mainWinners, &playerHands[idx])
				}
			}
			if len(mainWinners) > 0 {
				splitAmount := g.currentPot / len(mainWinners)
				remainder := g.currentPot % len(mainWinners)
				
				if len(mainWinners) > 1 {
					logrus.Infof("🤝 TIE in Main Pot! %d players split %d chips", 
						len(mainWinners), g.currentPot)
				}
				
				for j, winner := range mainWinners {
					award := splitAmount
					if j == 0 {
						award += remainder
					}
					g.playerStates[winner.Addr].Stack += award
					logrus.Infof("🏆 Main Pot Winner: %s receives %d chips with %s", 
						winner.Addr, award, winner.HandName)
				}
			}
		}
	} else {
		var winner *PlayerHand
		bestRank := int32(999999)
		tiePlayers := []*PlayerHand{}
		
		for idx := range playerHands {
			if playerHands[idx].Rank < bestRank {
				bestRank = playerHands[idx].Rank
				winner = &playerHands[idx]
				tiePlayers = []*PlayerHand{&playerHands[idx]}
			} else if playerHands[idx].Rank == bestRank {
				tiePlayers = append(tiePlayers, &playerHands[idx])
			}
		}
		
		if len(tiePlayers) > 1 {
			splitAmount := g.currentPot / len(tiePlayers)
			remainder := g.currentPot % len(tiePlayers)
			
			logrus.Infof("🤝 TIE DETECTED! %d players split the pot", len(tiePlayers))
			
			for i, ph := range tiePlayers {
				award := splitAmount
				if i == 0 {
					award += remainder
				}
				g.playerStates[ph.Addr].Stack += award
				logrus.Infof("🏆 TIE WINNER: %s receives %d chips with %s", 
					ph.Addr, award, ph.HandName)
			}
		} else if winner != nil {
			g.playerStates[winner.Addr].Stack += g.currentPot
			logrus.Infof("🏆 WINNER: %s wins %d chips with %s!", 
				winner.Addr, g.currentPot, winner.HandName)
		}
	}
	
	g.currentPot = 0
	g.sidePots = []SidePot{}
	g.revealedKeys = make(map[string]*CardKeys)
	g.foldedPlayerKeys = make(map[string]*CardKeys)
	g.setStatus(GameStatusHandComplete)
	
	logrus.Info("=== HAND COMPLETE ===")
}

func (g *Game) HandleShowdownKeyReveal(from string, msg MessageRevealKeys) {
	g.lock.Lock()

	g.revealedKeys[from] = msg.Keys
	expectedKeys := 0 
	for _, state := range g.playerStates {
		if state.IsActive && !state.IsFolded {
			expectedKeys++
		}
	}
	g.lock.Unlock()
	if len(g.revealedKeys) == expectedKeys {
		go g.ResolveWinner()
	}
}

// #####################################
// INTERNAL - HELPER - FUNCTIONS
// #####################################

func (g *Game) getMyHoleCardIndices() []int {
	myID := g.playerStates[g.listenAddr].RotationID
	return []int{myID*2, (myID*2)+1}
}

func (g *Game) getNextPlayerID(id int) int {
	if g.nextRotationID == 0 {return 0}
	return (id + 1) % g.nextRotationID
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

func (g *Game) getOtherPlayers() []string {
	all := g.playersList.List()
	others := []string{}
	for _, a := range all {
		if a != g.listenAddr {
			others = append(others, a)
		}
	}
	return others
}


func (g *Game) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		g.lock.RLock()
		logrus.WithFields(logrus.Fields{
			"status": GameStatus(g.currentStatus.Get()),
			"dealer": g.currentDealerID,
			"turn": g.currentPlayerTurnID,
			"pot": g.currentPot,
			"hand_size": len(g.myHand),
		}).Info("Game State Heartbeat")
		g.lock.RUnlock()
	}
}

func (g *Game) createSidePot(allinPlayer string, allinAmount int) {
	eligible := []string{}
	for addr, state := range g.playerStates {
		if !state.IsFolded && state.CurrentRoundBet >= allinAmount {
			eligible = append(eligible, addr)
		}
	}
	g.sidePots = append(g.sidePots, SidePot{
		Amount: 0,
		EligiblePlayers: eligible,
	})
}

func (g *Game) updatePlayerState(addr string, action PlayerAction, value int) {
	state := g.playerStates[addr]
	switch action {

	case PlayerActionFold:
		state.IsFolded = true
	case PlayerActionBet, PlayerActionRaise:
		actualBet := value 
		if actualBet > state.Stack {
			actualBet = state.Stack
			state.IsAllIn = true 
			logrus.Infof("Player %s is ALL-IN!", addr)
			g.createSidePot(addr, state.CurrentRoundBet+actualBet)
		}
		state.CurrentRoundBet += actualBet
		g.currentPot += actualBet
		state.Stack -= actualBet 
		if state.CurrentRoundBet > g.highestBet {
			g.highestBet = state.CurrentRoundBet
			g.lastRaiserID = state.RotationID
		}
	case PlayerActionCall:
		amountNeeded := g.highestBet - state.CurrentRoundBet
		actualCall := amountNeeded
		if actualCall > state.Stack {
			actualCall = state.Stack 
			state.IsAllIn = true 
			logrus.Infof("Player %s is ALL-IN!", addr)
			g.createSidePot(addr, state.CurrentRoundBet+actualCall)
		}
		state.CurrentRoundBet += actualCall
		g.currentPot += actualCall
		state.Stack -= actualCall
	case PlayerActionCheck:
		// no change in state
	}
}

func (g *Game) advanceTurnAndCheckRoundEnd() {
	g.incNextPlayer()
	if g.checkRoundEnd(){
		g.advanceToNextRound()
	}
}

func (g *Game) incNextPlayer() {
	startID := g.currentPlayerTurnID
	for {
		nextID := g.getNextPlayerID(startID)
		addr := g.rotationMap[nextID]
		state, ok := g.playerStates[addr]
		if ok && !state.IsFolded && !state.IsAllIn {
			g.currentPlayerTurnID = nextID
			return 
		}
		startID = nextID 
		if startID == g.currentPlayerTurnID {
			break
		}
	}
}

func (g *Game) checkRoundEnd() bool {
	activeCount := 0
	allinCount := 0 
	for _, state := range g.playerStates {
		if state.IsActive && !state.IsFolded {
			if state.IsAllIn {
				allinCount++
			} else {
				activeCount++
			}
		}
	}
	if activeCount <= 1 {
		return true 
	}
	allMatched := true 
	for _, state := range g.playerStates {
		if state.IsActive && !state.IsFolded && !state.IsAllIn {
			if state.CurrentRoundBet < g.highestBet {
				allMatched = false
				break
			}
		}
	}
	nextToAct := g.getNextActivePlayerID(g.lastRaiserID)
	if g.currentPlayerTurnID == nextToAct && allMatched {
		return true 
	}
	return false
}

func (g *Game) getValidActions() []PlayerAction {
	state := g.playerStates[g.listenAddr]
	actions := []PlayerAction{PlayerActionFold}
	if g.highestBet == 0 || state.CurrentRoundBet == g.highestBet {
		actions = append(actions, PlayerActionCheck)
	}
	if g.highestBet > state.CurrentRoundBet {
		actions = append(actions, PlayerActionCall)
	}
	actions = append(actions, PlayerActionRaise)
	return actions
}

func (g *Game) TakeAction(action PlayerAction, value int) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	myState := g.playerStates[g.listenAddr]

	if myState.RotationID != g.currentPlayerTurnID {
		return fmt.Errorf("it is not my turn to act: %s", g.listenAddr)
	}

	valid := false 
	for _, a := range g.getValidActions() {
		if a == action {
			valid = true 
			break
		}
	}
	if !valid {
		return fmt.Errorf("illegal action: you cannot %s right now", action)
	}
	switch action {
	case PlayerActionBet:
		if value < BigBlind {
			return fmt.Errorf("bet must be atleast the big blind (%d)", BigBlind)
		}
		if value > myState.Stack {
			return fmt.Errorf("bet (%d) exceeds your stack (%d)", value, myState.Stack)
		}
	case PlayerActionRaise:
		minRaise := g.highestBet * 2 
		if value < minRaise {
			return fmt.Errorf("raise must be at least %d (double current bet)", minRaise)
		}
		if value > myState.Stack {
			return fmt.Errorf("raise (%d) exceeds your stack (%d)", value, myState.Stack)
		}
	case PlayerActionCall:
		amountNeeded := g.highestBet - myState.CurrentRoundBet
		if amountNeeded > myState.Stack {
			logrus.Infof("Call will be all in for %d", myState.Stack)
		}
	}
	if action == PlayerActionFold {
		g.sendToPlayers(MessageRevealKeys{
			Keys: g.deckKeys,
		}, g.getOtherPlayers()...)
		myState.IsFolded = true
	}
	g.updatePlayerState(g.listenAddr, action, value)
	g.sendToPlayers(MessagePlayerAction{
		Action: action,
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Value: value,
	}, g.getOtherPlayers()...)
	g.advanceTurnAndCheckRoundEnd()
	go func(){
		if err := g.SaveSnapshot("game_snapshot.json"); err != nil {
			logrus.Errorf("Failed to save snapshot: %s", err)
		}
	}()
	return nil
}

func (g *Game) handlePlayerAction(from string, msg MessagePlayerAction) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.playerStates[from].RotationID != g.currentPlayerTurnID {
		return fmt.Errorf("player (%s) acting out of turn", from)
	}

	g.updatePlayerState(from, msg.Action, msg.Value)
	g.advanceTurnAndCheckRoundEnd()
	return nil
}

func (g *Game) advanceToNextRound() {
	nonFoldedCount := 0 
	var lastPlayerAddr string 
	for addr, state := range g.playerStates {
		if state.IsActive && !state.IsFolded{
			nonFoldedCount++
			lastPlayerAddr = addr
		}
	}
	if nonFoldedCount == 1{
		logrus.Infof("Only one player remains!, %s wins by default", lastPlayerAddr)
		g.playerStates[lastPlayerAddr].Stack += g.currentPot

		g.currentPot = 0
		g.sidePots = []SidePot{}
		g.revealedKeys = make(map[string]*CardKeys)
		g.foldedPlayerKeys = make(map[string]*CardKeys)
		g.setStatus(GameStatusHandComplete)

		go g.StartNewHand()
		return 
	}

	if GameStatus(g.currentStatus.Get()) == GameStatusHandComplete{
		logrus.Infof("Hand is complete. Starting next round.")
		go g.StartNewHand()
		return 
	}

	newStatus := g.getNextGameStatus()
	g.setStatus(newStatus)
	g.highestBet = 0 
	for _, state := range g.playerStates {
		state.CurrentRoundBet = 0
	}
	if newStatus == GameStatusShowdown {
		logrus.Infof("Advancing to %s", newStatus)
		g.InitiateShowdown()
		return 
	}

	if g.listenAddr == g.rotationMap[g.currentDealerID] {
		communityIndices := []int{}
		numPlayers := len(g.getReadyActivePlayers())
		switch newStatus {
		case GameStatusFlop:
			start := numPlayers * 2
			communityIndices = []int{start, start+1, start+2}
		case GameStatusTurn:
			communityIndices = []int{numPlayers*2 + 3}
		case GameStatusRiver:
			communityIndices = []int{numPlayers*2 + 4}
		}
		g.sendToPlayers(MessageGameState{
			Status: newStatus,
			CommunityCards: communityIndices,
		}, g.getOtherPlayers()...)
		g.SyncState(MessageGameState{
			Status: newStatus, 
			CommunityCards: communityIndices,
		})
	}
	g.currentPlayerTurnID = g.getNextActivePlayerID(g.currentDealerID)
	logrus.Infof("Advancing to next round: %s, Turn: %d", newStatus, g.currentPlayerTurnID)
}
// #####################################
// HELPER - FUNCTIONS
// #####################################

func (g *Game) getNextActivePlayerID(currentID int) int {
	startID := currentID
	for {
		nextID := g.getNextPlayerID(startID)
		addr, ok := g.rotationMap[nextID]
		if ok {
			state := g.playerStates[addr]
			if state.IsActive && !state.IsFolded{
				return nextID
			}
		}
		startID = nextID
		if startID == currentID {
			return currentID
		}
	}
}

func (g *Game) setStatus(s GameStatus) {
	g.currentStatus.Set(int32(s))
}

func (g *Game) getNextGameStatus() GameStatus {
	switch GameStatus(g.currentStatus.Get()){
	case GameStatusPreFlop: return GameStatusFlop
	case GameStatusFlop: 	return GameStatusTurn
	case GameStatusTurn: 	return GameStatusRiver
	case GameStatusRiver: 	return GameStatusShowdown
	default: 				return GameStatusHandComplete
	}
}

func (g *Game) sendToPlayers(payload any, addr ...string){
	g.broadcastch <- BroadcastTo{
		To: addr,
		Payload: payload,
	}
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

// #####################################
// GETTER - FUNCTIONS
// #####################################

func (g *Game) GetStatus() GameStatus {
	return GameStatus(g.currentStatus.Get())
}

func (g *Game) GetPlayerHand() []Card {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.myHand
}

func (g *Game) GetCommunityCards() []Card {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.communityCards
}

func (g *Game) GetCurrentPot() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.currentPot
}

func (g *Game) GetHighestBet() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.highestBet
}

func (g *Game) IsMyTurn() bool {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.playerStates[g.listenAddr].RotationID == g.currentPlayerTurnID
}

func (g *Game) GetMyStack() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.playerStates[g.listenAddr].Stack
}

func (g *Game) GetCurrentTurnID() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.currentPlayerTurnID
}