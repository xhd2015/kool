import { useEffect, useState } from 'react';
import { Button, Card, Space, Statistic, Typography } from 'antd';
import T from '@pagemeta/CounterCard.json';
import { getHomePage, sectionOf, type CounterItems, type PageSection } from '../api/page';
import { nextCounter } from '../api/counter';

/**
 * The demo card. Its title, hint and empty state come from the part the server
 * owns (@pagemeta/CounterCard.json is server/pagemeta/parts/CounterCard.json);
 * its rows come from the page document, so the browser and an agent reading
 * GET /api/pages/home see the same card.
 */
export function CounterCard() {
    const [section, setSection] = useState<PageSection<CounterItems> | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [nonce, setNonce] = useState(0);

    useEffect(() => {
        let cancelled = false;
        getHomePage()
            .then((doc) => {
                if (cancelled) {
                    return;
                }
                const counter = sectionOf<CounterItems>(doc, 'counter');
                if (counter) {
                    setSection(counter);
                    setError(null);
                }
            })
            .catch((err: unknown) => {
                if (cancelled) {
                    return;
                }
                setError(err instanceof Error ? err.message : String(err));
            });
        return () => {
            cancelled = true;
        };
    }, [nonce]);

    const increment = () => {
        void nextCounter().then(() => setNonce((n) => n + 1));
    };

    return (
        <Card title={T.title}>
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Typography.Text type="secondary">{T.hint}</Typography.Text>
                {error != null && <Typography.Text type="danger">{error}</Typography.Text>}
                {section == null && error == null && (
                    <Typography.Text type="secondary">Loading…</Typography.Text>
                )}
                {section != null && section.empty && (
                    <Typography.Text type="secondary">{T.empty}</Typography.Text>
                )}
                {section != null && (
                    <Space size="large">
                        <Statistic title="count" value={section.items?.last ?? 0} />
                        <Button type="primary" onClick={increment}>
                            increment
                        </Button>
                    </Space>
                )}
            </Space>
        </Card>
    );
}
