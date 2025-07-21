import './GitDiffViewer.css';

interface GitDiffViewerProps {
    diff: string;
    title: string;
}

interface DiffLine {
    type: 'header' | 'hunk' | 'added' | 'removed' | 'context' | 'file';
    content: string;
    lineNumber?: {
        old?: number;
        new?: number;
    };
}

const GitDiffViewer = ({ diff, title }: GitDiffViewerProps) => {
    const parseDiff = (diffText: string): DiffLine[] => {
        const lines = diffText.split('\n');
        const parsedLines: DiffLine[] = [];
        let oldLineNumber = 0;
        let newLineNumber = 0;

        for (const line of lines) {
            if (line.startsWith('diff --git') || line.startsWith('index ')) {
                parsedLines.push({ type: 'file', content: line });
            } else if (line.startsWith('---') || line.startsWith('+++')) {
                parsedLines.push({ type: 'header', content: line });
            } else if (line.startsWith('@@')) {
                // Parse hunk header like @@ -1,4 +1,4 @@
                const hunkMatch = line.match(/@@ -(\d+),?\d* \+(\d+),?\d* @@/);
                if (hunkMatch) {
                    oldLineNumber = parseInt(hunkMatch[1], 10);
                    newLineNumber = parseInt(hunkMatch[2], 10);
                }
                parsedLines.push({ type: 'hunk', content: line });
            } else if (line.startsWith('+')) {
                parsedLines.push({
                    type: 'added',
                    content: line.substring(1),
                    lineNumber: { new: newLineNumber }
                });
                newLineNumber++;
            } else if (line.startsWith('-')) {
                parsedLines.push({
                    type: 'removed',
                    content: line.substring(1),
                    lineNumber: { old: oldLineNumber }
                });
                oldLineNumber++;
            } else if (line.startsWith(' ') || line === '') {
                parsedLines.push({
                    type: 'context',
                    content: line.substring(1),
                    lineNumber: {
                        old: oldLineNumber,
                        new: newLineNumber
                    }
                });
                oldLineNumber++;
                newLineNumber++;
            } else {
                // Other lines (usually metadata or empty)
                parsedLines.push({ type: 'context', content: line });
            }
        }

        return parsedLines;
    };

    const diffLines = parseDiff(diff);

    if (!diff.trim()) {
        return (
            <div className="git-diff-viewer">
                <h4>{title}</h4>
                <div className="diff-empty">No changes detected</div>
            </div>
        );
    }

    return (
        <div className="git-diff-viewer">
            <h4>{title}</h4>
            <div className="diff-content">
                {diffLines.map((line, index) => (
                    <div key={index} className={`diff-line diff-line-${line.type}`}>
                        {line.type !== 'file' && line.type !== 'header' && line.type !== 'hunk' && (
                            <div className="line-numbers">
                                <span className="old-line-number">
                                    {line.lineNumber?.old || ''}
                                </span>
                                <span className="new-line-number">
                                    {line.lineNumber?.new || ''}
                                </span>
                            </div>
                        )}
                        <div className="line-content">
                            {line.type === 'added' && <span className="diff-prefix">+</span>}
                            {line.type === 'removed' && <span className="diff-prefix">-</span>}
                            {line.type === 'context' && <span className="diff-prefix"> </span>}
                            <span className="line-text">{line.content}</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default GitDiffViewer; 