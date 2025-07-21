import GitDiffViewer from './GitDiffViewer';
import './DiffModal.css';

interface DiffModalProps {
    isOpen: boolean;
    onClose: () => void;
    onReload: () => void;
    userDiff: string;
    currentDiff: string;
    filePath: string;
}



const DiffModal = ({
    isOpen,
    onClose,
    onReload,
    userDiff,
    currentDiff,
    filePath
}: DiffModalProps) => {
    if (!isOpen) return null;

    return (
        <div className="diff-modal-overlay">
            <div className="diff-modal">
                <div className="diff-modal-header">
                    <h2>Save Conflict Detected</h2>
                    <button className="close-button" onClick={onClose}>×</button>
                </div>

                <div className="diff-modal-content">
                    <p className="conflict-message">
                        The file <strong>{filePath}</strong> has been modified by another process.
                        Your changes cannot be saved until you reload the file.
                    </p>

                    <div className="diff-sections">
                        <div className="diff-section">
                            <GitDiffViewer diff={userDiff} title="Your Changes" />
                        </div>

                        <div className="diff-section">
                            <GitDiffViewer diff={currentDiff} title="External Changes" />
                        </div>
                    </div>
                </div>

                <div className="diff-modal-actions">
                    <button className="reload-button" onClick={onReload}>
                        Reload File
                    </button>
                    <button className="cancel-button" onClick={onClose}>
                        Cancel
                    </button>
                </div>
            </div>
        </div>
    );
};

export default DiffModal; 