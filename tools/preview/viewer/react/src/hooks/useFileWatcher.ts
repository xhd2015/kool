import { useEffect, useRef, useCallback } from 'react';

export interface FileChangeEvent {
    type: 'file_change';
    event: 'create' | 'delete' | 'modify' | 'rename';
    path: string;
    needsTreeRefresh: boolean;
}

interface UseFileWatcherProps {
    onFileChange?: (event: FileChangeEvent) => void;
    onTreeRefresh?: () => void;
    onFileModified?: (filePath: string) => void;
}

export const useFileWatcher = ({
    onFileChange,
    onTreeRefresh,
    onFileModified
}: UseFileWatcherProps) => {
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<number | null>(null);
    const reconnectAttempts = useRef(0);
    const maxReconnectAttempts = 5;

    const connect = useCallback(() => {
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/api/file-changes`;
            
            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log('File watcher WebSocket connected');
                reconnectAttempts.current = 0;
            };

            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data) as FileChangeEvent;
                    
                    if (data.type === 'file_change') {
                        console.log('File change detected:', data);
                        
                        // Call the general file change handler
                        onFileChange?.(data);
                        
                        // If it needs tree refresh, call the tree refresh handler
                        if (data.needsTreeRefresh) {
                            onTreeRefresh?.();
                        }
                        
                        // If it's a file modification, call the file modified handler
                        if (data.event === 'modify') {
                            onFileModified?.(data.path);
                        }
                    }
                } catch (error) {
                    console.error('Error parsing file change message:', error);
                }
            };

            ws.onclose = (event) => {
                console.log('File watcher WebSocket disconnected:', event.code, event.reason);
                wsRef.current = null;
                
                // Only attempt to reconnect if it wasn't a manual close
                if (event.code !== 1000 && reconnectAttempts.current < maxReconnectAttempts) {
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current), 10000);
                    reconnectAttempts.current++;
                    
                    console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttempts.current})`);
                    
                    reconnectTimeoutRef.current = setTimeout(() => {
                        connect();
                    }, delay);
                }
            };

            ws.onerror = (error) => {
                console.error('File watcher WebSocket error:', error);
            };

        } catch (error) {
            console.error('Failed to create file watcher WebSocket:', error);
        }
    }, [onFileChange, onTreeRefresh, onFileModified]);

    useEffect(() => {
        connect();

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close(1000, 'Component unmounting');
            }
        };
    }, [connect]);

    return {
        isConnected: wsRef.current?.readyState === WebSocket.OPEN
    };
};