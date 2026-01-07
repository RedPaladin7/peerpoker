import { ActionRequest, ActionResponse, HealthResponse, PlayersResponse, TableStateResponse } from "@/types/api";
import axios, { AxiosInstance } from "axios";

class PokerAPIClient {
    private client: AxiosInstance

    constructor(baseURL: string = "http://localhost:8080") {
        this.client = axios.create({
            baseURL,
            timeout: 10000,
            headers: {
                "Content-type": "application/json",
            },
        })

        this.client.interceptors.response.use(
            (response) => response, 
            (error) => {
                console.error("API Error: ", error.response?.data || error.message)
                return Promise.reject(error)
            }
        )
    }

    async health(): Promise<HealthResponse> {
        const response = await this.client.get<HealthResponse>("/api/health")
        return response.data
    }

    async getTableState(): Promise<TableStateResponse> {
        const response = await this.client.get<TableStateResponse>("/api/table")
        return response.data
    }

    async getPlayers(): Promise<PlayersResponse> {
        const response = await this.client.get<PlayersResponse>("/api/players")
        return response.data
    }

    async ready(): Promise<ActionResponse> {
        const response = await this.client.post<ActionResponse>("/api/ready")
        return response.data
    }

    async fold(): Promise<ActionResponse> {
        const response = await this.client.post<ActionResponse>("/api/fold")
        return response.data
    }

    async check(): Promise<ActionResponse> {
        const response = await this.client.post<ActionResponse>("/api/check")
        return response.data
    }
    
    async call(): Promise<ActionResponse> {
        const response = await this.client.post<ActionResponse>("/api/call")
        return response.data
    }

    async bet(value: number): Promise<ActionResponse> {
        const response = await this.client.post<ActionResponse>(
            "/api/bet",
            {value} as ActionRequest
        )
        return response.data
    }

    async raise(value: number): Promise<ActionResponse> {
        const response = await this.client.post<ActionResponse>(
            "/api/raise",
            {value} as ActionRequest
        )
        return response.data
    }

    setBaseURL(baseURL: string): void {
        this.client.defaults.baseURL = baseURL
    }
}

export const apiClient = new PokerAPIClient()

export default PokerAPIClient