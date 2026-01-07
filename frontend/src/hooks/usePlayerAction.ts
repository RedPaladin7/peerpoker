import { apiClient } from "@/lib/api_client";
import { PlayerAction } from "@/types/api";
import { useCallback, useState } from "react";

interface usePlayerActionReturn {
    executeAction: (action: PlayerAction, value?: number) => Promise<void>
    loading: boolean
    error: string | null
    lastAction: PlayerAction | null
}

export function usePlayerAction(
    onSuccess?: () => void
): usePlayerActionReturn {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [lastAction, setLastAction] = useState<PlayerAction | null>(null)

    const executeAction = useCallback(
        async (action: PlayerAction, value?: number) => {
            setLoading(true)
            setError(null)
            try {
                let response 
                switch (action){
                    case "READY":
                        response = await apiClient.ready()
                        break
                    case "FOLD":
                        response = await apiClient.fold()
                        break
                    case "CALL":
                        response = await apiClient.call()
                        break
                    case "CHECK":
                        response = await apiClient.check()
                        break
                    case "BET":
                        if (value == undefined) {
                            throw new Error("Value is required for BET action")
                        }
                        response = await apiClient.bet(value)
                        break
                    case "RAISE":
                        if (value == undefined) {
                            throw new Error("Value is required for RAISE action")
                        }
                        response = await apiClient.raise(value)
                        break
                    default:
                        throw new Error("Invalid action")
                }
                console.log("Action response:", response)
                setLastAction(action)
                if(onSuccess){
                    onSuccess()
                }
            } catch(err) {
                const errorMessage = err instanceof Error ? err.message : "Action failed"
                console.error("Action error: ", errorMessage)
                setError(errorMessage)
            } finally {
                setLoading(false)
            }
        },
        [onSuccess]
    )

    return {
        executeAction,
        loading,
        error,
        lastAction
    }
}