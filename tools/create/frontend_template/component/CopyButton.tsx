import { CopyOutlined, CheckOutlined } from '@ant-design/icons';
import { Button, message } from 'antd';
import "./DailyReportContent.css";
import { useState } from 'react';

export interface CopyButtonProps {
    content: string | (() => string | Promise<void> | undefined);
    id?: string
    children?: React.ReactNode;
    style?: React.CSSProperties;
}

export function CopyButton({ content, id, children, style }: CopyButtonProps) {
    const [copied, setCopied] = useState(false);

    const handleCopy = async () => {
        let hasError = false
        try {
            if (typeof content === 'function') {
                await content()
                return
            }
            const textToCopy = content;
            if (!textToCopy) {
                return; // Function-based copy handled internally
            }
            await navigator.clipboard.writeText(textToCopy);
            message.success('Copied to clipboard');

        } catch (err) {
            hasError = true
            message.error('Failed to copy');
        } finally {
            if (!hasError) {
                setCopied(true);
                setTimeout(() => setCopied(false), 500); // Reset after 1.5s
            }
        };
    }

    return (
        <Button
            id={id}
            type="link"
            icon={copied ? <CheckOutlined /> : <CopyOutlined />}
            onClick={handleCopy}
            style={style}
        >
            {children}
        </Button>
    );
}