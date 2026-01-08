'use client';

import { useGameState } from '@/hooks/useGameState';
import { usePlayerAction } from '@/hooks/usePlayerAction';

export default function GameTable() {
  const { table, players, loading, connected, refreshState } = useGameState();
  const { executeAction, loading: actionLoading, error: actionError } = usePlayerAction(refreshState);

  if (loading && !table) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white"></div>
      </div>
    );
  }

  // Helper function to check if an action is valid
  const isActionValid = (action: string): boolean => {
    if (!table) return false;
    return table.valid_actions.includes(action);
  };

  // Helper to check if it's actually our turn for game actions (not READY)
  const canTakeGameAction = (): boolean => {
    return table?.is_my_turn || false;
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-7xl mx-auto flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">🃏 Table: {table?.status || 'Loading...'}</h1>
        <div className={`px-4 py-1 rounded-full text-sm ${connected ? 'bg-green-900 text-green-300' : 'bg-red-900 text-red-300'}`}>
          {connected ? '● Live' : '○ Disconnected'}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Table Area */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-gray-800 rounded-2xl p-8 border border-gray-700 min-h-[300px] flex flex-col justify-center items-center relative">
            {/* Community Cards */}
            <div className="flex gap-4 mb-8">
              {table?.community_cards && table.community_cards.length > 0 ? (
                table.community_cards.map((card, i) => (
                  <div key={i} className="bg-white text-black w-16 h-24 rounded-lg flex items-center justify-center font-bold text-xl shadow-lg">
                    {card.display.split(' of ')[0]}
                  </div>
                ))
              ) : (
                <div className="text-gray-600 italic">Waiting for cards...</div>
              )}
            </div>

            {/* Pot Info */}
            <div className="bg-black/40 px-6 py-2 rounded-full border border-gray-600">
              <span className="text-gray-400">Total Pot: </span>
              <span className="text-green-400 font-bold">${table?.pot || 0}</span>
            </div>

            {/* Turn Indicator */}
            {table?.is_my_turn && (
              <div className="mt-4 bg-green-500/20 border border-green-500 px-4 py-2 rounded-full text-green-400 font-bold animate-pulse">
                YOUR TURN
              </div>
            )}
          </div>

          {/* Player Hand & Actions */}
          <div className="bg-gray-800 rounded-2xl p-6 border border-gray-700">
            <div className="flex justify-between items-end mb-6">
              <div>
                <p className="text-sm text-gray-400 mb-2">Your Hand</p>
                <div className="flex gap-3">
                  {table?.my_hand && table.my_hand.length > 0 ? (
                    table.my_hand.map((card, i) => (
                      <div key={i} className="bg-blue-100 text-blue-900 w-20 h-28 rounded-xl flex items-center justify-center font-bold text-2xl shadow-inner border-2 border-blue-400">
                        {card.display.split(' of ')[0]}
                      </div>
                    ))
                  ) : (
                    <div className="text-gray-600 italic">No cards yet</div>
                  )}
                </div>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-400">Your Stack</p>
                <p className="text-3xl font-bold text-green-400">${table?.my_stack || 0}</p>
              </div>
            </div>

            {/* Valid Actions Info */}
            {table?.valid_actions && table.valid_actions.length > 0 && (
              <div className="mb-4 text-sm text-gray-400">
                Valid actions: {table.valid_actions.join(', ')}
              </div>
            )}

            {/* Action Buttons */}
            <div className="pt-6 border-t border-gray-700">
              <div className="grid grid-cols-2 gap-4 mb-4">
                {/* READY button - always available when not in a hand */}
                <button
                  onClick={() => executeAction('READY')}
                  disabled={actionLoading || (table?.status !== 'WAITING' && table?.status !== 'PLAYER-READY')}
                  className="bg-blue-600 hover:bg-blue-700 disabled:opacity-30 disabled:cursor-not-allowed py-3 rounded-xl font-bold transition"
                >
                  READY
                </button>

                {/* FOLD button */}
                <button
                  onClick={() => executeAction('FOLD')}
                  disabled={actionLoading || !canTakeGameAction() || !isActionValid('FOLD')}
                  className="bg-red-600 hover:bg-red-700 disabled:opacity-30 disabled:cursor-not-allowed py-3 rounded-xl font-bold transition"
                >
                  FOLD
                </button>

                {/* CHECK button */}
                <button
                  onClick={() => executeAction('CHECK')}
                  disabled={actionLoading || !canTakeGameAction() || !isActionValid('CHECK')}
                  className="bg-gray-700 hover:bg-gray-600 disabled:opacity-30 disabled:cursor-not-allowed py-3 rounded-xl font-bold transition"
                >
                  CHECK
                </button>

                {/* CALL button */}
                <button
                  onClick={() => executeAction('CALL')}
                  disabled={actionLoading || !canTakeGameAction() || !isActionValid('CALL')}
                  className="bg-green-700 hover:bg-green-600 disabled:opacity-30 disabled:cursor-not-allowed py-3 rounded-xl font-bold transition"
                >
                  CALL {table?.highest_bet && table.highest_bet > 0 ? `($${table.highest_bet})` : ''}
                </button>
              </div>

              {/* BET/RAISE with input */}
              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2 flex gap-2">
                  <input
                    type="number"
                    placeholder={`Min: $${table?.min_raise || 20}`}
                    className="flex-1 bg-gray-700 border border-gray-600 p-3 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-white"
                    min={table?.min_raise || 20}
                    id="bet-input"
                  />
                  <button
                    onClick={() => {
                      const input = document.getElementById('bet-input') as HTMLInputElement;
                      const value = parseInt(input.value);
                      if (value && value > 0) {
                        const action = isActionValid('BET') ? 'BET' : 'RAISE';
                        executeAction(action, value);
                        input.value = '';
                      }
                    }}
                    disabled={actionLoading || !canTakeGameAction() || (!isActionValid('BET') && !isActionValid('RAISE'))}
                    className="bg-yellow-600 hover:bg-yellow-700 disabled:opacity-30 disabled:cursor-not-allowed px-6 py-3 rounded-xl font-bold transition"
                  >
                    {isActionValid('BET') ? 'BET' : 'RAISE'}
                  </button>
                </div>
              </div>
            </div>

            {/* Error Display */}
            {actionError && (
              <div className="mt-4 p-3 bg-red-900/50 border border-red-500 rounded-lg text-red-200 text-sm">
                {actionError}
              </div>
            )}
          </div>
        </div>

        {/* Sidebar: Players List */}
        <div className="space-y-4">
          <h2 className="text-xl font-bold px-2">Players ({players?.total_players || 0})</h2>
          {players?.players && players.players.length > 0 ? (
            players.players.map((p) => (
              <div 
                key={p.player_id} 
                className={`p-4 rounded-xl border transition-all ${
                  p.is_current_turn 
                    ? 'border-green-500 bg-green-500/10 shadow-lg shadow-green-500/20' 
                    : 'border-gray-700 bg-gray-800'
                }`}
              >
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">Player {p.player_id}</span>
                    {p.is_dealer && <span>🎯</span>}
                    {p.is_current_turn && <span className="text-green-400">◄</span>}
                  </div>
                  <span className="text-green-400 font-bold">${p.stack}</span>
                </div>
                
                {/* Player Status */}
                <div className="mt-2 text-xs flex gap-2 flex-wrap">
                  {p.is_small_blind && <span className="px-2 py-1 bg-blue-500/20 text-blue-400 rounded">SB</span>}
                  {p.is_big_blind && <span className="px-2 py-1 bg-purple-500/20 text-purple-400 rounded">BB</span>}
                  {p.is_folded && <span className="px-2 py-1 bg-red-500/20 text-red-400 rounded">FOLDED</span>}
                  {p.is_all_in && <span className="px-2 py-1 bg-yellow-500/20 text-yellow-400 rounded">ALL-IN</span>}
                  {p.current_bet > 0 && <span className="px-2 py-1 bg-gray-700 text-gray-300 rounded">Bet: ${p.current_bet}</span>}
                </div>
                
                {/* Connection Status */}
                <div className="mt-2 text-xs text-gray-500 truncate">
                  {p.listen_addr}
                </div>
              </div>
            ))
          ) : (
            <div className="text-gray-600 italic p-4">No players connected</div>
          )}
        </div>
      </div>
    </div>
  );
}