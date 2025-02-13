import { useState } from 'react'
import { Input, Space } from 'antd'
import { EditOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons'

export interface EditableTextProps {
    text: string
    onSave: (newText: string) => Promise<void>
    placeholder?: string
    maxRows?: number
}

export function EditableText({ text, onSave, placeholder = '-', maxRows = 5 }: EditableTextProps) {
    const [isEditing, setIsEditing] = useState(false)
    const [editText, setEditText] = useState(text)

    const handleEdit = () => {
        setIsEditing(true)
        setEditText(text)
    }

    const handleCancel = () => {
        setIsEditing(false)
        setEditText(text)
    }

    const handleSave = async () => {
        await onSave(editText)
        setIsEditing(false)
    }

    if (isEditing) {
        return (
            <Space.Compact style={{ width: '100%' }}>
                <Input.TextArea
                    value={editText}
                    onChange={e => setEditText(e.target.value)}
                    autoSize={{ minRows: 1, maxRows }}
                    style={{ width: 'calc(100% - 64px)' }}
                    onKeyDown={e => {
                        if (e.key === 'Escape') {
                            handleCancel()
                        } else if (e.key === 'Enter' && !e.shiftKey) {
                            e.preventDefault()
                            handleSave()
                        }
                    }}
                    autoFocus
                />
                <Space>
                    <CheckOutlined
                        style={{ color: '#52c41a', cursor: 'pointer', fontSize: '16px' }}
                        onClick={handleSave}
                    />
                    <CloseOutlined
                        style={{ color: '#ff4d4f', cursor: 'pointer', fontSize: '16px' }}
                        onClick={handleCancel}
                    />
                </Space>
            </Space.Compact>
        )
    }

    return (
        <Space>
            <span>{text || placeholder}</span>
            <EditOutlined
                style={{ color: '#1890ff', cursor: 'pointer' }}
                onClick={handleEdit}
            />
        </Space>
    )
} 