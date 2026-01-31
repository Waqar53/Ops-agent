import { io, Socket } from 'socket.io-client';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'http://localhost:8080';

export class WebSocketClient {
    private socket: Socket | null = null;
    private listeners: Map<string, Set<Function>> = new Map();

    connect() {
        if (this.socket?.connected) return;

        this.socket = io(WS_URL, {
            path: '/api/v1/ws',
            transports: ['websocket'],
        });

        this.socket.on('connect', () => {
            console.log('WebSocket connected');
        });

        this.socket.on('disconnect', () => {
            console.log('WebSocket disconnected');
        });

        // Handle different event types
        this.socket.on('deployment:update', (data) => {
            this.emit('deployment:update', data);
        });

        this.socket.on('metrics:update', (data) => {
            this.emit('metrics:update', data);
        });

        this.socket.on('logs:new', (data) => {
            this.emit('logs:new', data);
        });

        this.socket.on('alert:new', (data) => {
            this.emit('alert:new', data);
        });
    }

    disconnect() {
        this.socket?.disconnect();
        this.socket = null;
    }

    subscribe(projectId: string) {
        this.socket?.emit('subscribe', { projectId });
    }

    unsubscribe(projectId: string) {
        this.socket?.emit('unsubscribe', { projectId });
    }

    on(event: string, callback: Function) {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, new Set());
        }
        this.listeners.get(event)!.add(callback);
    }

    off(event: string, callback: Function) {
        this.listeners.get(event)?.delete(callback);
    }

    private emit(event: string, data: any) {
        this.listeners.get(event)?.forEach(callback => callback(data));
    }
}

export const wsClient = new WebSocketClient();
