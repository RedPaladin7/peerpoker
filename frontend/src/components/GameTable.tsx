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

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-7xl mx-auto flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">🃏 Table: {table?.status}</h1>
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
              {table?.community_cards.map((card, i) => (
                <div key={i} className="bg-white text-black w-16 h-24 rounded-lg flex items-center justify-center font-bold text-xl shadow-lg">
                  {card.display.split(' ')[0]} {/* Simplified display */}
                </div>
              ))}
              {(!table?.community_cards || table.community_cards.length === 0) && (
                <div className="text-gray-600 italic">Waiting for cards...</div>
              )}
            </div>

            {/* Pot Info */}
            <div className="bg-black/40 px-6 py-2 rounded-full border border-gray-600">
              <span className="text-gray-400">Total Pot: </span>
              <span className="text-green-400 font-bold">${table?.pot}</span>
            </div>
          </div>

          {/* Player Hand & Actions */}
          <div className="bg-gray-800 rounded-2xl p-6 border border-gray-700">
            <div className="flex justify-between items-end">
              <div>
                <p className="text-sm text-gray-400 mb-2">Your Hand</p>
                <div className="flex gap-3">
                  {table?.my_hand.map((card, i) => (
                    <div key={i} className="bg-blue-100 text-blue-900 w-20 h-28 rounded-xl flex items-center justify-center font-bold text-2xl shadow-inner border-2 border-blue-400">
                      {card.display.split(' ')[0]}
                    </div>
                  ))}
                </div>
              </div>
              <div className="text-right">
                <p className="text-sm text-gray-400">Your Stack</p>
                <p className="text-3xl font-bold text-green-400">${table?.my_stack}</p>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="mt-8 pt-6 border-t border-gray-700 grid grid-cols-4 gap-4">
              {['FOLD', 'CHECK', 'CALL', 'READY'].map((action) => (
                <button
                  key={action}
                  onClick={() => executeAction(action as any)}
                  disabled={actionLoading || (action !== 'READY' && !table?.is_my_turn)}
                  className="bg-gray-700 hover:bg-gray-600 disabled:opacity-30 py-3 rounded-xl font-bold transition"
                >
                  {action}
                </button>
              ))}
            </div>
            {actionError && <p className="text-red-400 text-sm mt-4 text-center">{actionError}</p>}
          </div>
        </div>

        {/* Sidebar: Players List */}
        <div className="space-y-4">
          <h2 className="text-xl font-bold px-2">Players</h2>
          {players?.players.map((p) => (
            <div key={p.player_id} className={`p-4 rounded-xl border ${p.is_current_turn ? 'border-green-500 bg-green-500/10' : 'border-gray-700 bg-gray-800'}`}>
              <div className="flex justify-between items-center">
                <span className="font-medium">Player {p.player_id} {p.is_dealer && '🎯'}</span>
                <span className="text-green-400">${p.stack}</span>
              </div>
              <div className="mt-2 text-xs flex gap-2">
                {p.is_small_blind && <span className="text-blue-400">SB</span>}
                {p.is_big_blind && <span className="text-purple-400">BB</span>}
                {p.is_folded && <span className="text-red-500 font-bold text-[10px] uppercase">Folded</span>}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}