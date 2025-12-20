package p2p

import (
	"fmt"

	"github.com/chehsunliu/poker"
)

func EvaluateBestHand(hole []Card, community []Card) (int32, string) {
	allCards := append(hole, community...)
	libCards := make([]poker.Card, len(allCards))
	for i, c := range allCards {
		libCards[i] = translateToLibCard(c)
	}
	rank := int32(poker.Evaluate(libCards))
	return rank, poker.RankString(rank)
}

func translateToLibCard(c Card) poker.Card {
	rankMap := map[int]string{
		1: "a", 10: "t", 11: "j", 12: "q", 13: "k",
	}
	suitMap := map[Suit]string{
		Spades: "s", Hearts: "h", Diamonds: "d", Clubs: "c",
	}
	rankStr, ok := rankMap[c.Value]
	if !ok {
		rankStr = fmt.Sprintf("%d", c.Value)
	}
	suitStr := suitMap[c.Suit]
	cardStr := fmt.Sprintf("%s%s", rankStr, suitStr)
	return poker.NewCard(cardStr)
}