import React, { useEffect, useState, useMemo } from 'react';
import { List, Button, Modal, message, Input } from 'antd';
import { debounce } from 'lodash';
import { bashApi } from '../api/bash';

const BashHistoryManager: React.FC = () => {
    const [history, setHistory] = useState<string[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(true);
    const [currentPage, setCurrentPage] = useState(1);
    const [search, setSearch] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');
    const pageSize = 10;

    const debouncedSetSearch = useMemo(
        () => debounce((val: string) => {
            setDebouncedSearch(val);
            setCurrentPage(1);
        }, 300),
        []
    );

    useEffect(() => {
        return () => {
            debouncedSetSearch.cancel();
        };
    }, [debouncedSetSearch]);

    const fetchHistory = async () => {
        try {
            setLoading(true);
            const data = await bashApi.list({
                page: currentPage,
                pageSize,
                search: debouncedSearch,
            });
            setHistory(data.list || []);
            setTotal(data.total);
        } catch (err: any) {
            message.error(err.message);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchHistory();
    }, [currentPage, debouncedSearch]);

    const handleDelete = (cmd: string) => {
        Modal.confirm({
            title: 'Delete Command',
            content: (
                <div style={{ whiteSpace: 'pre-wrap', maxHeight: '200px', overflow: 'auto' }}>
                    Are you sure you want to delete this command from history?
                    <br />
                    <code>{cmd}</code>
                </div>
            ),
            width: 600,
            onOk: async () => {
                try {
                    await bashApi.delete(cmd);
                    message.success('Command deleted');
                    fetchHistory();
                } catch (err: any) {
                    message.error(`Error deleting command: ${err.message}`);
                }
            },
        });
    };

    return (
        <div style={{ padding: '20px' }}>
            <h1>Bash History Manager</h1>
            <Input.Search
                placeholder="Search history..."
                value={search}
                onChange={(e) => {
                    setSearch(e.target.value);
                    debouncedSetSearch(e.target.value);
                }}
                onSearch={(value) => {
                    setSearch(value);
                    debouncedSetSearch.cancel();
                    setDebouncedSearch(value);
                    setCurrentPage(1);
                }}
                style={{ marginBottom: 16 }}
                allowClear
                enterButton
            />
            <List
                bordered
                loading={loading}
                dataSource={history}
                pagination={{
                    current: currentPage,
                    pageSize: pageSize,
                    total: total,
                    onChange: (page) => setCurrentPage(page),
                    showSizeChanger: false,
                    position: 'bottom',
                    align: 'center',
                }}
                renderItem={(item) => (
                    <List.Item
                        actions={[<Button danger onClick={() => handleDelete(item)}>Delete</Button>]}
                    >
                        <span style={{ wordBreak: 'break-all', fontFamily: 'monospace' }}>{item}</span>
                    </List.Item>
                )}
            />
        </div>
    );
};

export default BashHistoryManager;
