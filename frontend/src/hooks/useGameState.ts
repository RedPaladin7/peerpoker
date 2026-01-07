import { apiClient } from "@/lib/api_client";
import { PlayersResponse, TableStateResponse } from "@/types/api";
import { useCallback, useEffect, useState } from "react";

interface GameState {
    table: TableStateResponse | null 
    players: PlayersResponse | null 
    loading: boolean 
    error: string | null 
    connected: boolean
}

interface UseGameStateReturn extends GameState {
    refreshState: ()=> Promise<void>
    setPollingInterval: (interval: number) => void 
}

export function useGameState(
    pollingInterval: number = 2000
): UseGameStateReturn {
    const [state, setState] = useState<GameState>({
        table: null,
        players: null,
        loading: true,
        error: null,
        connected: false
    })
    const [interval, setInterval] = useState(pollingInterval)
    const fetchGameState = useCallback(async() => {
        try {
            const [tableData, playersData] = await Promise.all([
                apiClient.getTableState(),
                apiClient.getPlayers(),
            ])

            setState({
                table: tableData,
                players: playersData,
                loading: false,
                error: null,
                connected: true
            })
        } catch (error) {
            console.error("Failed to fetch game state:", error)
            setState((prev)=>({
                ...prev,
                loading: false,
                error: error instanceof Error ? error.message : "Failed to fetch game state",
                connected: false
            }))
        }
    }, [])

    const refreshState = useCallback(async() => {
        setState((prev)=>({...prev, loading: true}))
        await fetchGameState()
    }, [fetchGameState])

    const setPollingInterval = useCallback((newInterval: number)=>{
        setInterval(newInterval)
    }, [])

    useEffect(()=>{
        fetchGameState()
    }, [fetchGameState])

    useEffect(()=> {
        if (!state.connected || interval <= 0) {
            return 
        }
        const timer = window.setInterval(fetchGameState, interval)
        return () => window.clearInterval(timer)
    }, [interval, state.connected, fetchGameState])

    return {
        ...state,
        refreshState,
        setPollingInterval
    }
}