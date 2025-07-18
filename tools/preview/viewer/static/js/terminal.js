import { state } from './state.js';

// Initialize terminal with WebSocket streaming
export function initializeTerminal() {
    if (typeof Terminal === 'undefined' || typeof FitAddon === 'undefined') {
        console.error('xterm libraries not loaded');
        return;
    }

    state.terminal = new Terminal({
        cursorBlink: true,
        theme: {
            background: '#1e1e1e',
            foreground: '#d4d4d4',
            cursor: '#ffffff',
            selection: '#264f78'
        },
        fontSize: 14,
        fontFamily: 'Consolas, Monaco, "Courier New", monospace'
    });

    state.fitAddon = new FitAddon.FitAddon();
    state.terminal.loadAddon(state.fitAddon);

    const terminalContainer = document.getElementById('terminal-container');
    state.terminal.open(terminalContainer);
    state.fitAddon.fit();

    // Handle terminal input - send each character directly to bash
    state.terminal.onData((data) => {
        console.log('Terminal input data:', JSON.stringify(data));

        // Don't echo locally - let bash handle all display
        // Just send the character directly to the bash session
        sendInput(data);
    });

    // Handle window resize
    window.addEventListener('resize', () => {
        if (state.fitAddon && state.terminalVisible) {
            state.fitAddon.fit();
        }
    });

    // Don't initialize WebSocket here - wait until terminal is shown
    // initializeTerminalWebSocket();
}

// Initialize WebSocket for terminal streaming
function initializeTerminalWebSocket() {
    if (state.terminalWebSocket) {
        console.log('Closing existing WebSocket...');
        state.terminalWebSocket.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/stream`;

    console.log('Connecting to WebSocket:', wsUrl);
    console.log('Current location:', window.location.href);

    state.terminalWebSocket = new WebSocket(wsUrl);

    state.terminalWebSocket.onopen = function (event) {
        console.log('Terminal WebSocket connected successfully');
    };

    state.terminalWebSocket.onmessage = function (event) {
        console.log('WebSocket message received:', event.data);
        try {
            const data = JSON.parse(event.data);
            console.log('Parsed WebSocket data:', data);

            if (data.output) {
                console.log('Writing output to terminal:', JSON.stringify(data.output));
                state.terminal.write(data.output);
            }
            if (data.error) {
                console.log('Writing error to terminal:', JSON.stringify(data.error));
                state.terminal.write(`\x1b[31m${data.error}\x1b[0m`);
            }
            if (data.keepalive) {
                console.log('Received keepalive');
                // Just a keepalive message, ignore
            }
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    state.terminalWebSocket.onerror = function (event) {
        console.error('Terminal WebSocket error:', event);
    };

    state.terminalWebSocket.onclose = function (event) {
        console.log('Terminal WebSocket closed, code:', event.code, 'reason:', event.reason);
        console.log('Was clean close:', event.wasClean);

        // Try to reconnect after a delay only if terminal is still visible
        if (state.terminalVisible) {
            setTimeout(() => {
                console.log('Attempting to reconnect WebSocket...');
                initializeTerminalWebSocket();
            }, 5000);
        }
    };
}

// Send input to terminal via WebSocket
function sendInput(input) {
    console.log('Sending input to WebSocket:', JSON.stringify(input));
    if (state.terminalWebSocket && state.terminalWebSocket.readyState === WebSocket.OPEN) {
        const message = JSON.stringify({ input: input });
        console.log('WebSocket sending message:', message);
        state.terminalWebSocket.send(message);
    } else {
        console.log('WebSocket not ready, state:', state.terminalWebSocket ? state.terminalWebSocket.readyState : 'null');
    }
}

// Initialize terminal toggle
export function initializeTerminalToggle() {
    const toggleButton = document.getElementById('terminal-toggle');
    const terminalContainer = document.getElementById('terminal-container');
    const verticalResizer = document.querySelector('.vertical-resizer');
    const previewSection = document.querySelector('.preview-section');

    toggleButton.addEventListener('click', () => {
        state.terminalVisible = !state.terminalVisible;
        console.log('Terminal visibility toggled to:', state.terminalVisible);

        if (state.terminalVisible) {
            terminalContainer.classList.remove('hidden');
            verticalResizer.style.display = 'block';
            toggleButton.textContent = 'Hide Terminal';

            // Reset to split layout (80% preview, 20% terminal)
            previewSection.style.flex = '4';
            previewSection.style.height = '';
            terminalContainer.style.flex = '1';

            if (state.fitAddon) {
                setTimeout(() => state.fitAddon.fit(), 100);
            }

            // Start terminal streaming when terminal becomes visible
            if (!state.terminalWebSocket || state.terminalWebSocket.readyState === WebSocket.CLOSED) {
                console.log('Starting WebSocket connection...');
                initializeTerminalWebSocket();
            } else {
                console.log('WebSocket already connected, state:', state.terminalWebSocket.readyState);
            }
        } else {
            terminalContainer.classList.add('hidden');
            verticalResizer.style.display = 'none';
            toggleButton.textContent = 'Show Terminal';

            // Make preview take more space
            previewSection.style.flex = '1';
            previewSection.style.height = '';

            // Close terminal streaming when terminal is hidden
            if (state.terminalWebSocket) {
                console.log('Closing WebSocket connection...');
                state.terminalWebSocket.close();
                state.terminalWebSocket = null;
            }
        }
    });
}

// Clean up WebSocket on page unload
export function cleanupTerminal() {
    if (state.terminalWebSocket) {
        state.terminalWebSocket.close();
    }
} 