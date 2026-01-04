package p2p

import (
	"encoding/json"
	"os"
)

type GameSnapshot struct {
	CurrentStatus 	int32 
	CurrentPot 		int 
	PlayerStates 	map[string]*PlayerState
	RotationMap 	map[int]string 
	CurrentDealerID int 
	HighestBet 		int 
	CommunityCards 	[]Card
}

func (g *Game) SaveSnapshot(filename string) error {
	g.lock.RLock()
	defer g.lock.RUnlock()

	snapshot := GameSnapshot{
		CurrentStatus: g.currentStatus.Get(),
		CurrentPot: g.currentPot,
		PlayerStates: g.playerStates,
		RotationMap: g.rotationMap,
		CurrentDealerID: g.currentDealerID,
		HighestBet: g.highestBet,
		CommunityCards: g.communityCards,
	}

	data, err := json.MarshalIndent(snapshot, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (g *Game) LoadSnapshot(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var snapshot GameSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	g.lock.Lock()
	defer g.lock.Unlock()

	g.currentStatus.Set(snapshot.CurrentStatus)
	g.currentPot = snapshot.CurrentPot
	g.playerStates = snapshot.PlayerStates
	g.rotationMap = snapshot.RotationMap
	g.currentDealerID = snapshot.CurrentDealerID
	g.highestBet = snapshot.HighestBet
	g.communityCards = snapshot.CommunityCards
	
	return nil
}