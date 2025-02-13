import React from 'react'
import Editor from '@monaco-editor/react'

export interface JSONViewerProps {
    data: any
    height?: string
    width?: string
    theme?: 'vs-dark' | 'light'
    style?: React.CSSProperties
}

export function JSONViewer({
    data,
    height = '70vh',
    width = '100%',
    theme = 'light',
    style,
}: JSONViewerProps) {
    const formattedData = typeof data === 'string' ? data : JSON.stringify(data, null, 2)

    return (
        <div style={{ ...style }}>
            <Editor
                height={height}
                width={width}
                language="json"
                theme={theme}
                value={formattedData}
                options={{
                    readOnly: true,
                    minimap: { enabled: false },
                    folding: true,
                    lineNumbers: 'on',
                    wordWrap: 'on',
                    formatOnPaste: true,
                    automaticLayout: true,
                }}
            />
        </div>
    )
} 