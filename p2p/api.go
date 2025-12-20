package p2p

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

type APIServer struct {
	listenAddr string 
	game *Game
}

func NewAPIServer(listenAddr string, game *Game) *APIServer {
	return &APIServer{
		game: game,
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/ready", makeHTTPHandlerFunc(s.handlePlayerReady)).Methods("GET")
	r.HandleFunc("/fold", makeHTTPHandlerFunc(s.handlePlayerFold)).Methods("GET")
	r.HandleFunc("/check", makeHTTPHandlerFunc(s.handlePlayerCheck)).Methods("GET")
	r.HandleFunc("/call", makeHTTPHandlerFunc(s.handlePlayerCall)).Methods("GET")
	r.HandleFunc("/bet/{value}", makeHTTPHandlerFunc(s.handlePlayerBet)).Methods("GET")
	r.HandleFunc("/raise/{value}", makeHTTPHandlerFunc(s.handlePlayerRaise)).Methods("GET")
	r.HandleFunc("/table", makeHTTPHandlerFunc(s.handleGetTable)).Methods("GET")

	http.ListenAndServe(s.listenAddr, r)
}

func (s *APIServer) handleGetTable(w http.ResponseWriter, r *http.Request) error {
	validActions := s.game.getValidActions()
	resp := map[string]any{
		"status": 			s.game.currentStatus.Get(),
		"my_hand":			s.game.GetPlayerHand(),
		"community_cards": 	s.game.GetCommunityCards(),
		"pot": 				s.game.GetCurrentPot(),
		"highest_bet": 		s.game.GetHighestBet(),
		"valid_actions": 	validActions,
		"is_my_turn": 		s.game.IsMyTurn(),
		"my_stack": 		s.game.GetMyStack(),
		"current_turn_id": 	s.game.GetCurrentTurnID(),
	}
	return JSON(w, http.StatusOK, resp)
}

func (s *APIServer) handlePlayerReady(w http.ResponseWriter, r *http.Request) error {
	s.game.SetReady(s.game.listenAddr)
	return JSON(w, http.StatusOK, map[string]string{"status": "READY"})
}

func (s *APIServer) handlePlayerFold(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionFold, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]string{"status": "FOLD"})
}

func (s *APIServer) handlePlayerCheck(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionCheck, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]string{"status": "CHECK"})
}

func (s *APIServer) handlePlayerCall(w http.ResponseWriter, r *http.Request) error {
	if err := s.game.TakeAction(PlayerActionCall, 0); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]string{"status": "CALL"})
}

func (s *APIServer) handlePlayerBet(w http.ResponseWriter, r *http.Request) error {
	value, err := s.parseValue(r)
	if err != nil {
		return err
	}
	if err := s.game.TakeAction(PlayerActionBet, value); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{"status": "BET", "value": value})
}

func (s *APIServer) handlePlayerRaise(w http.ResponseWriter, r *http.Request) error {
	value, err := s.parseValue(r)
	if err != nil {
		return err
	}
	if err := s.game.TakeAction(PlayerActionRaise, value); err != nil {
		return err
	}
	return JSON(w, http.StatusOK, map[string]any{"status": "RAISE", "value": value})
}

func (s *APIServer) parseValue(r *http.Request) (int, error) {
	valueStr := mux.Vars(r)["value"]
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid bet value: %s", valueStr)
	}
	return value, nil
}