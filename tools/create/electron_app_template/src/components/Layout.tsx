import React from 'react';
import { Outlet } from 'react-router-dom';

export const Layout: React.FC = () => {
    return (
        <div style={{ display: 'flex', height: '100vh', flexDirection: 'column' }}>
            <header style={{ padding: '10px', borderBottom: '1px solid #ccc', backgroundColor: '#f0f0f0' }}>
                Template App
            </header>
            <main style={{ flex: 1, overflow: 'auto' }}>
                <Outlet />
            </main>
        </div>
    );
};
