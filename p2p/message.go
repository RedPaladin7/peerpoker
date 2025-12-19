package p2p

type Message struct {
	Payload any 
	From string
}

type BroadcastTo struct {
	To []string 
	Payload any 
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From: from,
		Payload: payload,
	}
}

type Handshake struct {
	Version string 
	GameVariant GameVariant
	ListenAddr string 
}

type MessagePeerList struct {
	Peers []string 
}

type MessagePlayerAction struct {
	Action PlayerAction
	Value int 
	CurrentGameStatus GameStatus
}

type MessageReady struct {}

func (msg MessageReady) String() string {
	return "MSG: READY"
}

type MessageEncDeck struct {
	Deck [][]byte
}

type MessageGameState struct {
	Status GameStatus
	CommunityCards []int
}

type MessagePreShuffle struct {
	Deck [][]byte
}

type MessageShuffleStatus struct {
	Deck [][]byte
}