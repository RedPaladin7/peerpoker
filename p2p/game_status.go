package p2p

type PlayerAction byte 

const (
	PlayerActionIdle PlayerAction = iota 
	PlayerActionFold 
	PlayerActionCheck 
	PlayerActionCall
	PlayerActionBet 
	PlayerActionRaise 
	PlayerActionAllIn
)

func (pa PlayerAction) String() string {
	switch pa {
		case PlayerActionIdle:
			return "IDLE"
		case PlayerActionFold:
			return "FOLD"
		case PlayerActionCheck:
			return "CHECK"
		case PlayerActionCall:
			return "CALL"
		case PlayerActionBet:
			return "BET"
		case PlayerActionRaise:
			return "RAISE"
		case PlayerActionAllIn:
			return "ALL-IN"
		default:
			return "INVALID"
	}
}

type GameStatus int32 

const (
	GameStatusWaiting GameStatus = iota 
	GameStatusPlayerReady 
	GameStatusDealing 
	GameStatusPreFlop 
	GameStatusFlop 
	GameStatusTurn 
	GameStatusRiver 
	GameStatusShowdown
	GameStatusHandComplete 
)

func (g GameStatus) String() string {
	switch g {
		case GameStatusWaiting:
			return "WAITING"
		case GameStatusPlayerReady:
			return "PLAYER-READY"
		case GameStatusDealing:
			return "DEALING"
		case GameStatusPreFlop:
			return "PREFLOP"
		case GameStatusFlop:
			return "FLOP"
		case GameStatusTurn:
			return "TURN"
		case GameStatusRiver:
			return "RIVER"
		case GameStatusShowdown:
			return "SHOWDOWN"
		case GameStatusHandComplete:
			return "HAND-COMPLETE"
		default:
			return "INVALID"
	}
}