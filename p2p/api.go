package p2p

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error 

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			JSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
	}
}

func JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type APIServer struct {
	listenAddr string 
	game 	   *Game
}

func NewAPIServer(listenAddr string, game *Game) *APIServer {
	return &APIServer{
		game: game,
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	r := mux.NewRouter()
	r.Use(enableCORS)

	r.HandleFunc("/api/ready", makeHTTPHandlerFunc(s.handlePlayerReady)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/fold", makeHTTPHandlerFunc(s.handlePlayerFold)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/check", makeHTTPHandlerFunc(s.handlePlayerCheck)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/call", makeHTTPHandlerFunc(s.handlePlayerCall)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/bet", makeHTTPHandlerFunc(s.handlePlayerBet)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/raise", makeHTTPHandlerFunc(s.handlePlayerRaise)).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/table", makeHTTPHandlerFunc(s.handleGetTable)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/players", makeHTTPHandlerFunc(s.handleGetPlayers)).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/health", makeHTTPHandlerFunc(s.handleHealth)).Methods("GET", "OPTIONS")

	logrus.WithFields(logrus.Fields{
		"addr": s.listenAddr,
	}).Info("API Server starting...")

	http.ListenAndServe(s.listenAddr, r)
}

type TableStateResponse struct {
	Status 			string 				`json:"status"`
	MyHand 			[]CardResponse 		`json:"myHand"`
	CommunityCards 	[]CardResponse 		`json:"community_cards"`
	Pot 			int 				`json:"pot"`
	HighestBet 		int 				`json:"highest_bet"`
	MinRaise 		int 				`json:"min_raise"`
	ValidActions 	[]string 			`json:"valid_actions"`
	IsMyTurn 		bool 				`json:"is_my_turn"`
	MyStack 		int 				`json:"my_stack"`
	CurrentTurnID 	int 				`json:"current_turn_id"`
	MyPlayerID 		int 				`json:"my_player_id"`
	DealerID 		int 				`json:"dealer_id"`
	SmallBlind 		int 				`json:"small_blind"`
	BigBlind 		int 				`json:"big_blind"`
	TimeBank 		int 				`json:"time_bank,omitempty"`
}

type CardResponse struct {
	Suit 	string 	`json:"suit"`
	Value 	int 	`json:"value"`
	Display string 	`json:"display"`
}

type PlayerStateResponse struct {
	PlayerID 		int 		`json:"player_id"`
	ListenAddr 		string 		`json:"listen_addr"`
	Stack 			int 		`json:"stack"`
	CurrentBet 		int 		`json:"current_bet"`
	IsActive 		bool 		`json:"is_active"`
	IsFolded 		bool 		`json:"is_folded"`
	IsAllIn 		bool 		`json:"is_all_in"`
	IsDealer 		bool 		`json:"is_dealer"`
	IsSmallBlind 	bool 		`json:"is_small_blind"`
	IsBigBlind 		bool 		`json:"is_big_blind"`
	IsCurrentTurn 	bool 		`json:"is_current_turn"`
}

type PlayerResponse struct {
	Players 		[]PlayerStateResponse 	`json:"players"`
	TotalPlayers 	int 					`json:"total_players"`
	ActivePlayers 	int 					`json:"active_players"`
}

type ActionRequest struct {
	Value	int		`json:"value,omitempty"`
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) error {
	return JSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"game_status": s.game.GetStatus().String(),
	})
}

func (s *APIServer) handleGetTable(w http.ResponseWriter, r *http.Request) error {
	s.game.lock.RLock()
	defer s.game.lock.RUnlock()

	validActions := s.game.getValidActions()
	actionStrings := make([]string, len(validActions))
	for i, action := range validActions {
		actionStrings[i] = action.String()
	}

	myHandResp := make([]CardResponse, 0) 
	if len(s.game.myHand) > 0 {
		myHandResp = make([]CardResponse, len(s.game.myHand))
		for i, card := range s.game.myHand {
			myHandResp[i] = CardResponse{
				Suit: card.Suit.String(),
				Value: card.Value,
				Display: card.String(),
			}
		}
	}
	communityCardResp := make([]CardResponse, len(s.game.communityCards))
	for i, card := range s.game.communityCards {
		communityCardResp[i] = CardResponse{
			Suit: card.Suit.String(),
			Value: card.Value,
			Display: card.String(),
		}
	}

	minRaise := s.game.highestBet + s.game.lastRaiseAmount
	if s.game.highestBet == 0 {
		minRaise = BigBlind
	}

	myState := s.game.playerStates[s.game.listenAddr]

	resp := TableStateResponse{
		Status: 		s.game.GetStatus().String(),
		MyHand: 		myHandResp,
		CommunityCards: 	communityCardResp,
		Pot: 			s.game.currentPot,
		HighestBet: 	s.game.highestBet,
		MinRaise: 		minRaise,
		ValidActions: 	actionStrings,
		IsMyTurn: 		myState.RotationID == s.game.currentPlayerTurnID,
		MyStack: 		myState.Stack,
		CurrentTurnID: 	s.game.currentPlayerTurnID,
		MyPlayerID: 	myState.RotationID,
		DealerID: 		s.game.currentDealerID,
		SmallBlind: 	SmallBlind,
		BigBlind: 		BigBlind,
	}

	return JSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleGetPlayers(w http.ResponseWriter, r *http.Request) error {
	s.game.lock.RLock()
	defer s.game.lock.RUnlock()

	players := make([]PlayerStateResponse, 0)
	activeCount := 0 

	var sbID, bbID int
	activeReadyPlayers := s.game.getReadyActivePlayers()
	if len(activeReadyPlayers) == 2{
		sbID = s.game.currentDealerID
		bbID = s.game.getNextPlayerID(sbID)
	} else if len(activeReadyPlayers) > 2 {
		sbID = s.game.getNextActivePlayerID(s.game.currentDealerID)
		bbID = s.game.getNextActivePlayerID(sbID)
	}

	for i := 0; i < s.game.nextRotationID; i++ {
		addr, ok := s.game.rotationMap[i]
		if !ok {
			continue
		}
		state, ok := s.game.playerStates[addr]
		if !ok {
			continue
		}
		if state.IsActive {
			activeCount++
		}
		players = append(players, PlayerStateResponse{
			PlayerID: 		state.RotationID,
			ListenAddr: 	state.ListenAddr,
			Stack: 			state.Stack,
			CurrentBet: 	state.CurrentRoundBet,
			IsActive: 		state.IsActive,
			IsFolded: 		state.IsFolded,
			IsAllIn: 		state.IsAllIn,
			IsDealer: 		state.RotationID == s.game.currentDealerID,
			IsSmallBlind: 	state.RotationID == sbID,
			IsBigBlind: 	state.RotationID == bbID,
			IsCurrentTurn: 	state.RotationID == s.game.currentPlayerTurnID,
		})
	}

	resp := PlayerResponse{
		Players: players,
		TotalPlayers: len(players),
		ActivePlayers: activeCount,
	}

	return JSON(w, http.StatusOK, resp)
}

func (s *APIServer) handlePlayerReady(w http.ResponseWriter, r *http.Request) error {
	s.game.SetReady(s.game.listenAddr)
	return JSON(w, http.StatusOK, map[string]string{
		"status": "READY",
		"player": s.game.listenAddr,
	})
}

func (s *APIServer) handlePlayerFold(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionFold, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{
		"status": "FOLD",
		"player": s.game.listenAddr,
	})
}

func (s *APIServer) handlePlayerCheck(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionCheck, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{
		"status": "CHECK",
		"player": s.game.listenAddr,
	})
}

func (s *APIServer) handlePlayerCall(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionCall, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{
		"status": "CALL",
		"player": s.game.listenAddr,
	})
}

func (s *APIServer) handlePlayerBet(w http.ResponseWriter, r *http.Request) error {
	value, err := parseActionValue(r, "bet")
	if err != nil {
		return err
	}
	if err := s.game.TakeAction(PlayerActionBet, value); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{
		"status": "BET",
		"player": s.game.listenAddr,
		"value": value,
	})
}

func (s *APIServer) handlePlayerRaise(w http.ResponseWriter, r *http.Request) error {
	value, err := parseActionValue(r, "raise")
	if err != nil {
		return err
	}
	if err := s.game.TakeAction(PlayerActionRaise, value); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{
		"status": "RAISE",
		"player": s.game.listenAddr,
		"value": value,
	})
}

func parseActionValue(r *http.Request, actionName string) (int, error) {
	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return 0, fmt.Errorf("invalid request body: %s", err)
	}
	if req.Value <= 0 {
		return 0, fmt.Errorf("%s value must be positive, got: %d", actionName, req.Value)
	}
	return req.Value, nil
}