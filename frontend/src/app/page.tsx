'use client';

import { useGameState } from '@/hooks/useGameState';
import { usePlayerAction } from '@/hooks/usePlayerAction';

export default function Home() {
  const { table, players, loading, error, connected, refreshState } = useGameState();
  const { executeAction, loading: actionLoading, error: actionError } = usePlayerAction(refreshState);

  if (loading && !table) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
          <p>Connecting to poker server...</p>
        </div>
      </div>
    );
  }

  if (error && !connected) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">
        <div className="text-center max-w-md">
          <div className="text-red-500 text-6xl mb-4">⚠️</div>
          <h2 className="text-2xl font-bold mb-2">Connection Failed</h2>
          <p className="text-gray-400 mb-4">{error}</p>
          <button
            onClick={refreshState}
            className="bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded-lg transition"
          >
            Retry Connection
          </button>
          <div className="mt-6 text-sm text-gray-500">
            <p>Make sure your Go server is running:</p>
            <code className="block mt-2 bg-gray-800 p-2 rounded">go run main.go</code>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      {/* Header */}
      <div className="max-w-7xl mx-auto mb-8">
        <div className="flex justify-between items-center">
          <h1 className="text-4xl font-bold">🃏 Decentralized Poker</h1>
          <div className="flex items-center gap-4">
            <div className={`flex items-center gap-2 ${connected ? 'text-green-400' : 'text-red-400'}`}>
              <div className={`w-3 h-3 rounded-full ${connected ? 'bg-green-400' : 'bg-red-400'} animate-pulse`}></div>
              {connected ? 'Connected' : 'Disconnected'}
            </div>
            <button
              onClick={refreshState}
              disabled={loading}
              className="bg-gray-700 hover:bg-gray-600 px-4 py-2 rounded-lg transition disabled:opacity-50"
            >
              🔄 Refresh
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Table State */}
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-2xl font-bold mb-4">🎲 Table State</h2>
          {table ? (
            <div className="space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-400">Status:</span>
                <span className="font-semibold text-yellow-400">{table.status}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Pot:</span>
                <span className="font-semibold text-green-400">${table.pot}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Your Stack:</span>
                <span className="font-semibold">${table.my_stack}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Current Bet:</span>
                <span className="font-semibold">${table.highest_bet}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Min Raise:</span>
                <span className="font-semibold">${table.min_raise}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-400">Your Turn:</span>
                <span className={`font-semibold ${table.is_my_turn ? 'text-green-400' : 'text-gray-500'}`}>
                  {table.is_my_turn ? 'YES ✓' : 'NO'}
                </span>
              </div>
              
              {/* Your Cards */}
              {table.my_hand.length > 0 && (
                <div className="border-t border-gray-700 pt-3 mt-3">
                  <p className="text-gray-400 mb-2">Your Cards:</p>
                  <div className="flex gap-2">
                    {table.my_hand.map((card, idx) => (
                      <div key={idx} className="bg-white text-gray-900 px-3 py-2 rounded font-mono text-sm">
                        {card.display}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Community Cards */}
              {table.community_cards.length > 0 && (
                <div className="border-t border-gray-700 pt-3 mt-3">
                  <p className="text-gray-400 mb-2">Community Cards:</p>
                  <div className="flex gap-2">
                    {table.community_cards.map((card, idx) => (
                      <div key={idx} className="bg-white text-gray-900 px-3 py-2 rounded font-mono text-sm">
                        {card.display}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Valid Actions */}
              <div className="border-t border-gray-700 pt-3 mt-3">
                <p className="text-gray-400 mb-2">Valid Actions:</p>
                <div className="flex flex-wrap gap-2">
                  {table.valid_actions.map((action) => (
                    <span key={action} className="bg-blue-600 px-3 py-1 rounded text-sm">
                      {action}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            <p className="text-gray-400">No table data available</p>
          )}
        </div>

        {/* Players */}
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-2xl font-bold mb-4">👥 Players ({players?.total_players || 0})</h2>
          {players && players.players.length > 0 ? (
            <div className="space-y-3">
              {players.players.map((player) => (
                <div
                  key={player.player_id}
                  className={`p-4 rounded-lg border-2 ${
                    player.is_current_turn
                      ? 'border-green-400 bg-gray-700'
                      : 'border-gray-700 bg-gray-800'
                  }`}
                >
                  <div className="flex justify-between items-center mb-2">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold">Player {player.player_id}</span>
                      {player.is_dealer && <span className="text-yellow-400">🎲 D</span>}
                      {player.is_small_blind && <span className="text-blue-400">SB</span>}
                      {player.is_big_blind && <span className="text-purple-400">BB</span>}
                    </div>
                    <span className="text-green-400 font-semibold">${player.stack}</span>
                  </div>
                  <div className="flex gap-2 text-sm">
                    {player.current_bet > 0 && (
                      <span className="bg-blue-600 px-2 py-1 rounded">Bet: ${player.current_bet}</span>
                    )}
                    {player.is_folded && <span className="bg-red-600 px-2 py-1 rounded">Folded</span>}
                    {player.is_all_in && <span className="bg-yellow-600 px-2 py-1 rounded">All-In</span>}
                    {!player.is_active && <span className="bg-gray-600 px-2 py-1 rounded">Inactive</span>}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-400">No players connected</p>
          )}
        </div>

        {/* Action Controls */}
        <div className="bg-gray-800 rounded-lg p-6 lg:col-span-2">
          <h2 className="text-2xl font-bold mb-4">🎮 Actions</h2>
          
          {actionError && (
            <div className="bg-red-900/50 border border-red-500 text-red-200 px-4 py-3 rounded mb-4">
              {actionError}
            </div>
          )}

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <button
              onClick={() => executeAction('READY')}
              disabled={actionLoading}
              className="bg-green-600 hover:bg-green-700 disabled:bg-gray-600 px-6 py-3 rounded-lg font-semibold transition"
            >
              Ready
            </button>
            <button
              onClick={() => executeAction('FOLD')}
              disabled={actionLoading || !table?.is_my_turn}
              className="bg-red-600 hover:bg-red-700 disabled:bg-gray-600 px-6 py-3 rounded-lg font-semibold transition"
            >
              Fold
            </button>
            <button
              onClick={() => executeAction('CHECK')}
              disabled={actionLoading || !table?.is_my_turn}
              className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 px-6 py-3 rounded-lg font-semibold transition"
            >
              Check
            </button>
            <button
              onClick={() => executeAction('CALL')}
              disabled={actionLoading || !table?.is_my_turn}
              className="bg-yellow-600 hover:bg-yellow-700 disabled:bg-gray-600 px-6 py-3 rounded-lg font-semibold transition"
            >
              Call
            </button>
            <button
              onClick={() => executeAction('BET', 50)}
              disabled={actionLoading || !table?.is_my_turn}
              className="bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 px-6 py-3 rounded-lg font-semibold transition"
            >
              Bet $50
            </button>
            <button
              onClick={() => executeAction('RAISE', table?.min_raise || 40)}
              disabled={actionLoading || !table?.is_my_turn}
              className="bg-orange-600 hover:bg-orange-700 disabled:bg-gray-600 px-6 py-3 rounded-lg font-semibold transition"
            >
              Raise Min
            </button>
          </div>

          <p className="text-gray-400 text-sm mt-4">
            Note: This is a test interface. Full UI with bet slider coming next!
          </p>
        </div>
      </div>
    </div>
  );
}