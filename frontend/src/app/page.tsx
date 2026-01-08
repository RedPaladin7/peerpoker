'use client';

import { useState } from 'react';
import GameTable from "@/components/GameTable"

export default function EntranceScreen() {
  const [view, setView] = useState<'entrance' | 'game'>('entrance');
  const [joinAddr, setJoinAddr] = useState('');
  const [connecting, setConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreateGame = () => {
    // In P2P, "creating" just means staying as the seed node
    console.log('Creating game as seed node');
    setView('game');
  };

  const handleJoinGame = async () => {
    if (!joinAddr.trim()) {
      setError('Please enter a peer address');
      return;
    }

    setConnecting(true);
    setError(null);

    try {
      // Get the port from URL params or default to 8080
      const params = new URLSearchParams(window.location.search);
      const port = params.get('port') || '8080';
      const apiUrl = `http://localhost:${port}/api/connect`;

      console.log(`Attempting to connect to peer ${joinAddr} via ${apiUrl}`);

      const response = await fetch(apiUrl, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ addr: joinAddr }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Connection failed: ${response.status}`);
      }

      const data = await response.json();
      console.log('Connection response:', data);

      // Give the backend a moment to establish the connection
      setTimeout(() => {
        setView('game');
      }, 1000);

    } catch (e) {
      const errorMsg = e instanceof Error ? e.message : 'Failed to connect to peer';
      console.error('Connection error:', errorMsg);
      setError(errorMsg);
    } finally {
      setConnecting(false);
    }
  };

  if (view === 'game') {
    return <GameTable />;
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">
      <div className="bg-gray-800 p-8 rounded-xl shadow-2xl w-96 border border-gray-700">
        <h1 className="text-3xl font-bold mb-2 text-center">🃏 PeerPoker</h1>
        <p className="text-gray-400 text-sm text-center mb-8">Decentralized Texas Hold'em</p>
        
        <div className="space-y-6">
          <button 
            onClick={handleCreateGame}
            className="w-full bg-green-600 hover:bg-green-700 py-3 rounded-lg font-bold transition transform hover:scale-105"
          >
            Create New Game
          </button>

          <div className="relative flex items-center">
            <div className="flex-grow border-t border-gray-600"></div>
            <span className="flex-shrink mx-4 text-gray-500 text-sm">OR</span>
            <div className="flex-grow border-t border-gray-600"></div>
          </div>

          <div className="space-y-3">
            <div>
              <label className="block text-sm text-gray-400 mb-2">
                Peer Address
              </label>
              <input 
                type="text" 
                placeholder="localhost:3000" 
                value={joinAddr}
                onChange={(e) => {
                  setJoinAddr(e.target.value);
                  setError(null);
                }}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    handleJoinGame();
                  }
                }}
                disabled={connecting}
                className="w-full bg-gray-700 border border-gray-600 p-3 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              />
              <p className="text-xs text-gray-500 mt-1">
                Example: localhost:3000
              </p>
            </div>
            
            <button 
              onClick={handleJoinGame}
              disabled={connecting || !joinAddr.trim()}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed py-3 rounded-lg font-bold transition transform hover:scale-105 flex items-center justify-center gap-2"
            >
              {connecting ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  Connecting...
                </>
              ) : (
                'Join Existing Game'
              )}
            </button>
          </div>

          {/* Error Display */}
          {error && (
            <div className="bg-red-900/50 border border-red-500 rounded-lg p-3 text-red-200 text-sm">
              {error}
            </div>
          )}

          {/* Info Box */}
          <div className="bg-blue-900/30 border border-blue-500/50 rounded-lg p-4 text-sm">
            <p className="text-blue-300 font-semibold mb-2">How to play:</p>
            <ul className="text-blue-200 space-y-1 text-xs">
              <li>• Start a game on one machine (Create)</li>
              <li>• Join from other machines using host:port</li>
              <li>• Press READY when 2+ players join</li>
              <li>• Game starts automatically!</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}