import { App, Tabs } from "antd"
import { useEffect, useState, ReactNode } from "react"
import { Outlet, Route, Routes, useLocation, useNavigate } from "react-router"

export function AppRoutes() {
    const navigate = useNavigate()
    const curPath = useLocation().pathname

    return <App>
        <Routes>
            <Route path="/" element={<AppIndex displayTabs={[
                { name: "Tab 1", route: "/tab1" },
                { name: "Tab 2", route: "/tab2" },
                { name: "Tab 3", route: "/tab3" },
            ]} />} >
                <Route path="tab1" element={<div>Tab 1</div>} />
                <Route path="tab2" element={<div>Tab 2</div>} />
                <Route path="tab3" element={<div>Tab 3</div>} />
            </Route>
            <Route path="*" element={<div>
                404 Not Found: {curPath}
                <button onClick={() => {
                    navigate("/")
                }}>Back</button>
            </div>} />
        </Routes>
    </App>
}


function AppIndex({ displayTabs: tabs }: { displayTabs: { name: string, route: string }[] }) {
    const appTabs = tabs || []

    const defaultTab = appTabs[0]?.route
    const navigate = useNavigate()

    const loc = useLocation()

    const [activeKey, setActiveKey] = useState(defaultTab)

    useEffect(() => {
        if (loc.pathname === "/") {
            navigate(defaultTab)
            return
        }
        const tab = appTabs.find((tab) => tab.route === loc.pathname)
        if (tab) {
            setActiveKey(tab.name)
        }
    }, [loc.pathname])

    return <div>
        <Tabs
            activeKey={activeKey}
            onChange={(key) => {
                setActiveKey(key)
                const tab = appTabs.find((tab) => tab.name === key)
                if (tab) {
                    navigate(tab.route)
                }
            }}
            items={appTabs.map((tab) => ({
                key: tab.name,
                label: tab.name,
                children: tab.element
            }))} />
        <Outlet />
    </div>
}