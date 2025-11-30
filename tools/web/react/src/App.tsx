import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { useState } from 'react';
import './App.css';

function Home() {
    const [count, setCount] = useState(0);

    return (
        <div style={{ textAlign: 'center', padding: '50px' }}>
            <h1>Welcome to Kool Go-React</h1>
            <p>
                Edit <code>src/App.tsx</code> and save to test HMR
            </p>
            <div className="card">
                <button onClick={() => setCount((count) => count + 1)}>
                    count is {count}
                </button>
            </div>
            <div style={{ marginTop: '20px' }}>
                <Link to="/about" style={{ fontSize: '18px', color: '#646cff', textDecoration: 'none' }}>
                    Go to About Page
                </Link>
            </div>
        </div>
    );
}

function About() {
    return (
        <div style={{ textAlign: 'center', padding: '50px' }}>
            <h1>About</h1>
            <p>This is a generic about page.</p>
            <Link to="/" style={{ fontSize: '18px', color: '#646cff', textDecoration: 'none' }}>
                Back to Home
            </Link>
        </div>
    );
}

function App() {
    return (
        <Router>
            <nav style={{ padding: '10px 20px', borderBottom: '1px solid #eee', display: 'flex', gap: '20px' }}>
                <Link to="/">Home</Link>
                <Link to="/about">About</Link>
            </nav>

            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/about" element={<About />} />
            </Routes>
        </Router>
    );
}

export default App;
