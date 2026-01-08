// frontend/app/page.tsx
'use client';

import { useState } from 'react';
import { useGameState } from '@/hooks/useGameState';
import GameTable from "@/components/GameTable"

export default function EntranceScreen() {
  const [view, setView] = useState<'entrance' | 'game'>('entrance');
  const [joinAddr, setJoinAddr] = useState('');
  const { connected, error, refreshState } = useGameState();

  const handleCreateGame = () => {
    // In P2P, "creating" just means staying as the seed node
    setView('game');
  };

  const handleJoinGame = async () => {
    try {
      const port = new URLSearchParams(window.location.search).get('port') || '8080';
      await fetch(`http://localhost:${port}/api/connect`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ addr: joinAddr }),
      });
      setView('game');
      refreshState();
    } catch (e) {
      alert("Failed to connect to peer");
    }
  };

  if (view === 'game') {
    return <GameTable />; // Your existing Home() UI logic
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">
      <div className="bg-gray-800 p-8 rounded-xl shadow-2xl w-96 border border-gray-700">
        <h1 className="text-3xl font-bold mb-8 text-center">🃏 PeerPoker</h1>
        
        <div className="space-y-6">
          <button 
            onClick={handleCreateGame}
            className="w-full bg-green-600 hover:bg-green-700 py-3 rounded-lg font-bold transition"
          >
            Create New Game
          </button>

          <div className="relative flex items-center">
            <div className="flex-grow border-t border-gray-600"></div>
            <span className="flex-shrink mx-4 text-gray-500">OR</span>
            <div className="flex-grow border-t border-gray-600"></div>
          </div>

          <div className="space-y-2">
            <input 
              type="text" 
              placeholder="Peer Address (localhost:3000)" 
              value={joinAddr}
              onChange={(e) => setJoinAddr(e.target.value)}
              className="w-full bg-gray-700 border border-gray-600 p-3 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button 
              onClick={handleJoinGame}
              className="w-full bg-blue-600 hover:bg-blue-700 py-3 rounded-lg font-bold transition"
            >
              Join Existing Game
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}