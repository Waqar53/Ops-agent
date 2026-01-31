const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http:
export interface Project {
    id: string;
    name: string;
    status: string;
    language: string;
    framework: string;
    gitRepo: string;
    createdAt: string;
    lastDeployment?: Deployment;
}
export interface Deployment {
    id: string;
    projectId: string;
    environmentId: string;
    version: string;
    gitCommit: string;
    strategy: string;
    status: 'pending' | 'building' | 'deploying' | 'success' | 'failed';
    deployedBy: string;
    deployedAt: string;
    durationSeconds?: number;
}
export interface Environment {
    id: string;
    projectId: string;
    name: string;
    type: 'production' | 'staging' | 'development' | 'preview';
    status: string;
    url?: string;
    createdAt: string;
}
export interface Metric {
    timestamp: number;
    value: number;
}
export interface LogEntry {
    timestamp: string;
    level: string;
    message: string;
    service: string;
}
class ApiClient {
    private baseUrl: string;
    constructor() {
        this.baseUrl = API_BASE;
    }
    private getAuthHeaders(): HeadersInit {
        const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
        return token ? { 'Authorization': `Bearer ${token}` } : {};
    }
    private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...this.getAuthHeaders(),
            ...options.headers,
        };
        const response = await fetch(url, { ...options, headers });
        if (!response.ok) {
            throw new Error(`API error: ${response.statusText}`);
        }
        return response.json();
    }
    async getProjects(): Promise<Project[]> {
        return this.request<Project[]>('/api/v1/projects');
    }
    async getProject(id: string): Promise<Project> {
        return this.request<Project>(`/api/v1/projects/${id}`);
    }
    async createProject(data: Partial<Project>): Promise<Project> {
        return this.request<Project>('/api/v1/projects', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }
    async analyzeProject(id: string): Promise<any> {
        return this.request(`/api/v1/projects/${id}/analyze`, {
            method: 'POST',
        });
    }
    async getDeployments(projectId: string): Promise<Deployment[]> {
        return this.request<Deployment[]>(`/api/v1/projects/${projectId}/deployments`);
    }
    async deploy(projectId: string, data: {
        environment: string;
        strategy?: string;
        branch?: string;
    }): Promise<Deployment> {
        return this.request<Deployment>(`/api/v1/projects/${projectId}/deploy`, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }
    async rollback(deploymentId: string): Promise<void> {
        return this.request(`/api/v1/deployments/${deploymentId}/rollback`, {
            method: 'POST',
        });
    }
    async getEnvironments(projectId: string): Promise<Environment[]> {
        return this.request<Environment[]>(`/api/v1/projects/${projectId}/environments`);
    }
    async getMetrics(projectId: string, metricName: string, timeRange: string): Promise<Metric[]> {
        return this.request<Metric[]>(
            `/api/v1/projects/${projectId}/metrics/${metricName}?range=${timeRange}`
        );
    }
    async getLogs(projectId: string, params?: {
        limit?: number;
        filter?: string;
        service?: string;
    }): Promise<LogEntry[]> {
        const query = new URLSearchParams(params as any).toString();
        return this.request<LogEntry[]>(`/api/v1/projects/${projectId}/logs?${query}`);
    }
    async getCost(projectId: string): Promise<any> {
        return this.request(`/api/v1/projects/${projectId}/cost`);
    }
    async getCostForecast(projectId: string): Promise<any> {
        return this.request(`/api/v1/projects/${projectId}/cost/forecast`);
    }
    async getCostRecommendations(projectId: string): Promise<any> {
        return this.request(`/api/v1/cost/recommendations?project_id=${projectId}`);
    }
    async applyCostRecommendation(recommendationId: string): Promise<void> {
        return this.request('/api/v1/cost/apply', {
            method: 'POST',
            body: JSON.stringify({ recommendation_id: recommendationId }),
        });
    }
    async getDashboardStats(projectId?: string): Promise<{
        cpu: number;
        memory: number;
        requests: number;
        deployments: number;
        uptime: number;
        latency: number;
        errors: number;
        alerts: number;
    }> {
        const query = projectId ? `?project_id=${projectId}` : '';
        return this.request(`/api/v1/stats${query}`);
    }
    async getAlerts(projectId: string, status?: string): Promise<any[]> {
        const query = status ? `&status=${status}` : '';
        return this.request(`/api/v1/alerts?project_id=${projectId}${query}`);
    }
    async resolveAlert(alertId: string): Promise<void> {
        return this.request(`/api/v1/alerts/resolve?alert_id=${alertId}`, {
            method: 'POST',
        });
    }
    async login(email: string, password: string): Promise<{ user: any; token: string }> {
        return this.request('/api/v1/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
        });
    }
    async register(name: string, email: string, password: string): Promise<{ user: any; token: string }> {
        return this.request('/api/v1/auth/register', {
            method: 'POST',
            body: JSON.stringify({ name, email, password }),
        });
    }
    async getCurrentUser(): Promise<any> {
        return this.request('/api/v1/auth/me');
    }
    async createApiKey(name: string): Promise<{ key: string }> {
        return this.request('/api/v1/auth/api-keys', {
            method: 'POST',
            body: JSON.stringify({ name }),
        });
    }
}
export const api = new ApiClient();
