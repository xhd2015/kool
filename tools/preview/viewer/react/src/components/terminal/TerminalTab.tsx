import { useEffect, useRef, useState } from 'react';
import { Terminal as XTerm } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css'; // Import xterm CSS only when this component is used
import './TerminalTab.css';

interface TerminalTabProps {
    tabId: string;
    isActive: boolean;
    isVisible: boolean;
}

interface TerminalSession {
    terminal: XTerm;
    fitAddon: FitAddon;
    websocket: WebSocket | null;
    connectionStatus: 'connecting' | 'connected' | 'disconnected';
}

interface WebSocketCallbacks {
    onConnectionStatusChange: (status: 'connecting' | 'connected' | 'disconnected') => void;
    onSendTerminalSize: () => void;
    onReconnect: () => void;
}

const TerminalTab: React.FC<TerminalTabProps> = ({ tabId, isActive, isVisible }) => {
    const terminalRef = useRef<HTMLDivElement>(null);
    const sessionRef = useRef<TerminalSession | null>(null);
    const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('disconnected');

    // Initialize terminal session (only once per tab)
    useEffect(() => {
        if (!terminalRef.current || sessionRef.current) return;

        console.log(`Initializing terminal session for tab ${tabId}...`);

        // Create new terminal instance
        const terminal = new XTerm({
            cursorBlink: true,
            theme: {
                background: '#1e1e1e',
                foreground: '#d4d4d4',
                cursor: '#ffffff',
            },
            fontSize: 14,
            fontFamily: 'Consolas, Monaco, "Courier New", monospace',
            screenReaderMode: false, // Disable screen reader mode to minimize helper elements
        });

        const fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);

        // Open terminal in container
        terminal.open(terminalRef.current);
        fitAddon.fit();

        // Store session
        sessionRef.current = {
            terminal,
            fitAddon,
            websocket: null,
            connectionStatus: 'disconnected'
        };

        console.log(`Terminal session initialized for tab ${tabId}`);

        // Handle terminal input - send to backend
        terminal.onData((data) => {
            console.log(`Terminal input data for tab ${tabId}:`, JSON.stringify(data));
            sendInput(data);
        });

        // Handle terminal resize
        terminal.onResize((size) => {
            console.log(`Terminal resized for tab ${tabId}:`, size);
            sendTerminalSize();
        });

        return () => {
            console.log(`Disposing terminal session for tab ${tabId}...`);
            if (sessionRef.current) {
                if (sessionRef.current.websocket) {
                    sessionRef.current.websocket.close();
                }
                sessionRef.current.terminal.dispose();
                sessionRef.current = null;
            }
        };
    }, [tabId]);

    // Handle WebSocket connection when tab becomes active and visible
    useEffect(() => {
        if (!sessionRef.current) return;

        if (isActive && isVisible) {
            // Connect WebSocket for active tab
            if (!sessionRef.current.websocket || sessionRef.current.websocket.readyState === WebSocket.CLOSED) {
                const callbacks: WebSocketCallbacks = {
                    onConnectionStatusChange: setConnectionStatus,
                    onSendTerminalSize: sendTerminalSize,
                    onReconnect: () => initializeWebSocket(sessionRef, tabId, callbacks, isActive, isVisible)
                };
                initializeWebSocket(sessionRef, tabId, callbacks, isActive, isVisible);
            }
            // Fit terminal when it becomes active
            setTimeout(() => {
                if (sessionRef.current?.fitAddon) {
                    sessionRef.current.fitAddon.fit();
                    sendTerminalSize();
                }
            }, 100);
        }
    }, [isActive, isVisible, tabId]);

    // Handle window resize for active tab
    useEffect(() => {
        const handleResize = () => {
            if (sessionRef.current && isActive && isVisible) {
                setTimeout(() => {
                    sessionRef.current?.fitAddon.fit();
                    sendTerminalSize();
                }, 100);
            }
        };

        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, [isActive, isVisible]);

    // Handle custom command execution events
    useEffect(() => {
        const handleExecuteCommand = (event: CustomEvent) => {
            // Only execute on the active terminal tab
            if (isActive && isVisible && event.detail?.command) {
                sendInput(event.detail.command);
            }
        };

        window.addEventListener('executeTerminalCommand', handleExecuteCommand as EventListener);
        return () => window.removeEventListener('executeTerminalCommand', handleExecuteCommand as EventListener);
    }, [isActive, isVisible]);



    const sendInput = (input: string) => {
        console.log(`Sending input to WebSocket for tab ${tabId}:`, JSON.stringify(input));
        if (sessionRef.current?.websocket && sessionRef.current.websocket.readyState === WebSocket.OPEN) {
            const message = JSON.stringify({ input: input });
            console.log(`WebSocket sending message for tab ${tabId}:`, message);
            sessionRef.current.websocket.send(message);
        } else {
            console.log(`WebSocket not ready for tab ${tabId}, state:`, sessionRef.current?.websocket ? sessionRef.current.websocket.readyState : 'null');
        }
    };

    const sendTerminalSize = () => {
        if (sessionRef.current?.websocket && sessionRef.current.websocket.readyState === WebSocket.OPEN && sessionRef.current.terminal) {
            const cols = sessionRef.current.terminal.cols;
            const rows = sessionRef.current.terminal.rows;
            console.log(`Sending terminal size for tab ${tabId}:`, { cols, rows });

            const message = JSON.stringify({
                resize: {
                    cols: cols,
                    rows: rows,
                },
            });
            sessionRef.current.websocket.send(message);
        }
    };

    const getConnectionStatusText = () => {
        switch (connectionStatus) {
            case 'connecting':
                return 'Connecting...';
            case 'connected':
                return 'Connected';
            case 'disconnected':
                return 'Disconnected';
        }
    };

    const getConnectionStatusClass = () => {
        switch (connectionStatus) {
            case 'connecting':
                return 'status-connecting';
            case 'connected':
                return 'status-connected';
            case 'disconnected':
                return 'status-disconnected';
        }
    };

    return (
        <div
            className={`terminal-tab-content ${isActive ? 'active' : 'inactive'}`}
            style={{ display: isActive ? 'flex' : 'none' }}
        >
            {isActive && (
                <div className="terminal-status-bar">
                    <span className={`connection-status ${getConnectionStatusClass()}`}>
                        {getConnectionStatusText()}
                    </span>
                </div>
            )}
            <div
                ref={terminalRef}
                className="terminal-container"
            />
        </div>
    );
};


// Standalone WebSocket initialization function
function initializeWebSocket(
    sessionRef: React.MutableRefObject<TerminalSession | null>,
    tabId: string,
    callbacks: WebSocketCallbacks,
    isActive: boolean,
    isVisible: boolean
) {
    if (!sessionRef.current) return;

    if (sessionRef.current.websocket) {
        console.log(`Closing existing WebSocket for tab ${tabId}...`);
        sessionRef.current.websocket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/stream`;

    console.log(`Connecting to WebSocket for tab ${tabId}:`, wsUrl);
    callbacks.onConnectionStatusChange('connecting');

    const ws = new WebSocket(wsUrl);
    sessionRef.current.websocket = ws;
    sessionRef.current.connectionStatus = 'connecting';

    ws.onopen = () => {
        console.log(`Terminal WebSocket connected successfully for tab ${tabId}`);
        callbacks.onConnectionStatusChange('connected');
        if (sessionRef.current) {
            sessionRef.current.connectionStatus = 'connected';
        }
        // Send initial terminal size when connection is established
        callbacks.onSendTerminalSize();
    };

    ws.onmessage = (event) => {
        console.log(`WebSocket message received for tab ${tabId}:`, event.data);
        try {
            const data = JSON.parse(event.data);
            console.log(`Parsed WebSocket data for tab ${tabId}:`, data);

            if (data.output && sessionRef.current?.terminal) {
                console.log(`Writing output to terminal tab ${tabId}:`, JSON.stringify(data.output));
                sessionRef.current.terminal.write(data.output);
            }
            if (data.error && sessionRef.current?.terminal) {
                console.log(`Writing error to terminal tab ${tabId}:`, JSON.stringify(data.error));
                sessionRef.current.terminal.write(`\x1b[31m${data.error}\x1b[0m`);
            }
            if (data.keepalive) {
                console.log(`Received keepalive for tab ${tabId}`);
                // Just a keepalive message, ignore
            }
        } catch (error) {
            console.error(`Error parsing WebSocket message for tab ${tabId}:`, error);
        }
    };

    ws.onerror = (event) => {
        console.error(`Terminal WebSocket error for tab ${tabId}:`, event);
        callbacks.onConnectionStatusChange('disconnected');
        if (sessionRef.current) {
            sessionRef.current.connectionStatus = 'disconnected';
        }
    };

    ws.onclose = (event) => {
        console.log(`Terminal WebSocket closed for tab ${tabId}, code:`, event.code, 'reason:', event.reason);
        callbacks.onConnectionStatusChange('disconnected');
        if (sessionRef.current) {
            sessionRef.current.connectionStatus = 'disconnected';
        }

        // Try to reconnect after a delay only if tab is still active and visible
        if (isActive && isVisible) {
            setTimeout(() => {
                console.log(`Attempting to reconnect WebSocket for tab ${tabId}...`);
                callbacks.onReconnect();
            }, 5000);
        }
    };
};


export default TerminalTab; 