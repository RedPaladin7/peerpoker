package p2p

import "fmt"

type Suit int 

const (
	Spades Suit = iota 
	Hearts 
	Diamonds
	Clubs 
)

func (s Suit) String() string {
	return []string{"SPADES", "HEARTS", "DIAMONDS", "CLUBS"}[s]
}

func (s Suit) Unicode() string {
	return []string{"♠", "♥", "♦", "♣"}[s]
}

type Card struct {
	Suit Suit 
	Value int
}

func (c Card) String() string {
	valueStr := ""
	switch c.Value {
	case 1: valueStr = "ACE"
	case 11: valueStr = "JACK"
	case 12: valueStr = "QUEEN"
	case 13: valueStr = "KING"
	default: valueStr = fmt.Sprintf("%d", c.Value)
	}
	return fmt.Sprintf("%s of %s %s", valueStr, c.Suit, c.Suit.Unicode())
}

func NewCardFromByte(b byte) Card {
	return Card {
		Suit: Suit(b/13),
		Value: int(b%13) + 1,
	}
}

func CreatePlaceHolderDeck() [][]byte {
	deck := make([][]byte, 52)
	for i := 0; i < 52; i++ {
		deck[i] = []byte{byte(i)}
	}
	return deck
}