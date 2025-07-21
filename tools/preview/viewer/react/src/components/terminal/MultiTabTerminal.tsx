import { useState, useCallback } from 'react';
import TerminalTab from './TerminalTab';
import './MultiTabTerminal.css';

interface MultiTabTerminalProps {
    isVisible: boolean;
    onToggle: () => void;
}

interface TabData {
    id: string;
    title: string;
    isActive: boolean;
}

const MultiTabTerminal: React.FC<MultiTabTerminalProps> = ({ isVisible, onToggle }) => {
    const [tabs, setTabs] = useState<TabData[]>([
        { id: 'tab-1', title: 'Terminal 1', isActive: true }
    ]);
    const [nextTabId, setNextTabId] = useState(2);

    const addNewTab = useCallback(() => {
        const newTabId = `tab-${nextTabId}`;
        setTabs(prevTabs => [
            ...prevTabs.map(tab => ({ ...tab, isActive: false })),
            { id: newTabId, title: `Terminal ${nextTabId}`, isActive: true }
        ]);
        setNextTabId(prev => prev + 1);
    }, [nextTabId]);

    const switchTab = useCallback((tabId: string) => {
        setTabs(prevTabs =>
            prevTabs.map(tab => ({ ...tab, isActive: tab.id === tabId }))
        );
    }, []);

    const closeTab = useCallback((tabId: string) => {
        setTabs(prevTabs => {
            const filteredTabs = prevTabs.filter(tab => tab.id !== tabId);

            // If we're closing the active tab, make another tab active
            if (filteredTabs.length > 0) {
                const wasActiveTab = prevTabs.find(tab => tab.id === tabId)?.isActive;
                if (wasActiveTab) {
                    filteredTabs[Math.max(0, filteredTabs.length - 1)].isActive = true;
                }
            }

            return filteredTabs;
        });
    }, []);

    return (
        <div className={`multi-tab-terminal ${!isVisible ? 'collapsed' : ''}`}>
            <div className="terminal-header">
                {isVisible ? (
                    <>
                        <div className="tab-bar">
                            <div className="tabs-container">
                                {tabs.map(tab => (
                                    <div
                                        key={tab.id}
                                        className={`terminal-tab ${tab.isActive ? 'active' : ''}`}
                                        onClick={() => switchTab(tab.id)}
                                    >
                                        <span className="tab-title">{tab.title}</span>
                                        {tabs.length > 1 && (
                                            <button
                                                className="tab-close"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    closeTab(tab.id);
                                                }}
                                            >
                                                ×
                                            </button>
                                        )}
                                    </div>
                                ))}
                                <button className="add-tab-button" onClick={addNewTab} title="Add new terminal">
                                    +
                                </button>
                            </div>
                        </div>
                        <div className="terminal-toggle-container">
                            <button className="terminal-toggle" onClick={onToggle}>
                                Hide Terminal
                            </button>
                        </div>
                    </>
                ) : (
                    <div className="terminal-toggle-container-collapsed">
                        <button className="terminal-toggle" onClick={onToggle}>
                            Show Terminal
                        </button>
                    </div>
                )}
            </div>

            <div className={`terminal-content ${!isVisible ? 'hidden' : ''}`}>
                {tabs.map(tab => (
                    <TerminalTab
                        key={tab.id}
                        tabId={tab.id}
                        isActive={tab.isActive}
                        isVisible={isVisible}
                    />
                ))}
            </div>
        </div>
    );
};

export default MultiTabTerminal; 