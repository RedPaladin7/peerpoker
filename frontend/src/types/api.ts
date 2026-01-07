export interface CardResponse {
    suit: string 
    value: number 
    display: string
}

export interface TableStateResponse {
    status: string 
    my_hand: CardResponse[]
    community_cards: CardResponse[]
    pot: number 
    highest_bet: number 
    min_raise: number 
    valid_actions: string[]
    is_my_turn: boolean 
    my_stack: number 
    current_turn_id: number 
    my_player_id: number 
    dealer_id: number 
    small_blind: number 
    big_blind: number 
    time_bank?: number
}

export interface PlayerStateResponse {
  player_id: number;
  listen_addr: string;
  stack: number;
  current_bet: number;
  is_active: boolean;
  is_folded: boolean;
  is_all_in: boolean;
  is_dealer: boolean;
  is_small_blind: boolean;
  is_big_blind: boolean;
  is_current_turn: boolean;
}

export interface PlayersResponse {
  players: PlayerStateResponse[];
  total_players: number;
  active_players: number;
}

export interface HealthResponse {
  status: string;
  game_status: string;
}

export interface ActionResponse {
  status: string;
  value?: number;
  player: string;
}

export interface ActionRequest {
  value?: number;
}

export type PlayerAction = 
  | "FOLD" 
  | "CHECK" 
  | "CALL" 
  | "BET" 
  | "RAISE" 
  | "READY";

export type GameStatus = 
  | "WAITING"
  | "PLAYER-READY"
  | "DEALING"
  | "PREFLOP"
  | "FLOP"
  | "TURN"
  | "RIVER"
  | "SHOWDOWN"
  | "HAND-COMPLETE";