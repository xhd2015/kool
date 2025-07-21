import { useState, useEffect } from 'react';
import './Header.css';

const Header = () => {
    const [isDarkTheme, setIsDarkTheme] = useState(false);
    const [isStartingPlantuml, setIsStartingPlantuml] = useState(false);
    const [plantumlStatus, setPlantumlStatus] = useState({ isRunning: false, port: 0 });

    // Initialize theme from localStorage on mount
    useEffect(() => {
        const savedTheme = localStorage.getItem('theme');
        const isDark = savedTheme === 'dark';

        setIsDarkTheme(isDark);

        if (isDark) {
            document.body.classList.add('dark-theme');
        } else {
            document.body.classList.remove('dark-theme');
        }
    }, []);

    // Check PlantUML server status on mount
    useEffect(() => {
        checkPlantumlStatus();
    }, []);

    const toggleTheme = () => {
        const newTheme = !isDarkTheme;
        setIsDarkTheme(newTheme);

        if (newTheme) {
            document.body.classList.add('dark-theme');
            localStorage.setItem('theme', 'dark');
        } else {
            document.body.classList.remove('dark-theme');
            localStorage.setItem('theme', 'light');
        }
    };

    const checkPlantumlStatus = async () => {
        try {
            const response = await fetch('/api/plantuml-status');
            if (response.ok) {
                const status = await response.json();
                setPlantumlStatus(status);
            }
        } catch (error) {
            console.error('Failed to check PlantUML status:', error);
        }
    };

    const startPlantumlServer = async () => {
        if (isStartingPlantuml) return;

        setIsStartingPlantuml(true);

        try {
            const response = await fetch('/api/start-plantuml', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (data.success && data.command) {
                // Send command to terminal via custom event
                const event = new CustomEvent('executeTerminalCommand', {
                    detail: { command: data.command + '\n' }
                });
                window.dispatchEvent(event);

                // Update status
                setPlantumlStatus({ isRunning: true, port: data.port });
            }
        } catch (error) {
            console.error('Failed to start PlantUML server:', error);
        } finally {
            setIsStartingPlantuml(false);
        }
    };

    const stopPlantumlServer = async () => {
        if (isStartingPlantuml) return;

        setIsStartingPlantuml(true);

        try {
            const response = await fetch('/api/stop-plantuml', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            if (data.success) {
                // Update status - backend handles the Docker command directly
                setPlantumlStatus({ isRunning: false, port: 0 });
            }
        } catch (error) {
            console.error('Failed to stop PlantUML server:', error);
        } finally {
            setIsStartingPlantuml(false);
        }
    };

    const handlePlantumlButtonClick = () => {
        if (plantumlStatus.isRunning) {
            stopPlantumlServer();
        } else {
            startPlantumlServer();
        }
    };

    return (
        <div className="header">
            <span>Directory Preview</span>
            <div className="header-actions">
                <button
                    className={`plantuml-button ${plantumlStatus.isRunning ? 'stop' : 'start'}`}
                    onClick={handlePlantumlButtonClick}
                    disabled={isStartingPlantuml}
                >
                    {isStartingPlantuml
                        ? (plantumlStatus.isRunning ? 'Stopping...' : 'Starting...')
                        : (plantumlStatus.isRunning ? `Stop PlantUML (${plantumlStatus.port})` : 'Start PlantUML')
                    }
                </button>
                <div className="theme-switcher">
                    <span>Theme:</span>
                    <button className="theme-toggle" onClick={toggleTheme}>
                        {isDarkTheme ? 'Light' : 'Dark'}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default Header; 