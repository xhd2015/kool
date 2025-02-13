import { Alert } from 'antd'
import { useState } from 'react'
import { DownOutlined, UpOutlined } from '@ant-design/icons'

export interface ErrorMsgProps {
    error: string | null
}

export function ErrorMsg({ error }: ErrorMsgProps) {
    const [expanded, setExpanded] = useState(false)

    if (!error) {
        return null
    }

    const lines = error.split('\n')
    const hasMore = lines.length > 2
    const displayedLines = expanded ? lines : lines.slice(0, 2)
    const message = displayedLines.join('\n') + (hasMore && !expanded ? '...' : '')

    return (
        <Alert
            type="error"
            message={
                <div style={{ cursor: hasMore ? 'pointer' : 'default' }} onClick={() => hasMore && setExpanded(!expanded)}>
                    <div style={{ whiteSpace: 'pre-wrap' }}>{message}</div>
                    {hasMore && (
                        <div style={{ textAlign: 'center', marginTop: 4 }}>
                            {expanded ? <UpOutlined /> : <DownOutlined />}
                        </div>
                    )}
                </div>
            }
            style={{ marginBottom: 16 }}
        />
    )
}
