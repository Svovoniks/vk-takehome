import { Layout, Table } from "antd";
import { ColumnType } from "antd/es/table";
import { useCallback, useEffect, useRef, useState } from "react";

const { Header, Content } = Layout;

type IpData = {
    ip: string,
    ping_ms: number,
    pinged_at: string
}

const columns: ColumnType<IpData>[] = [
    { title: "IP", dataIndex: "ip" },
    { title: "Ping ms", dataIndex: "ping_ms" },
    { title: "Last ping date", dataIndex: "pinged_at" },
];

const API_PING_DATA_URL = "http://localhost:4242/events";
const API_WS_URL = "ws://localhost:4242/ws";

function parseIpData(data: []): IpData[] {
    return data.map((a, i) => ({
        key: i,
        ip: a['ip'],
        ping_ms: a['ping_ms'],
        pinged_at: (new Date(a['pinged_at'])).toLocaleString('ru'),
    }) as IpData).sort((a, b) => a['ip'].localeCompare(b['ip']))
}

function App() {
    const [tableData, setTableData] = useState<IpData[]>([])

    const updateData = useCallback(async () => {
        try {
            const resp = await fetch(API_PING_DATA_URL);
            const json = await resp.json();
            setTableData(parseIpData(json));
        } catch (e) {
            console.error("Error fetching data:", e);
        }
    }, [])

    useEffect(() => {
        updateData()
    }, [updateData])

    const ws = useRef<WebSocket | null>(null);
    useEffect(() => {
        const connect = () => {
            if (ws.current) {
                ws.current.close()
            }

            ws.current = new WebSocket(API_WS_URL)

            ws.current.onopen = () => console.log("Connected to WebSocket");

            ws.current.onmessage = () => {
                console.log("Updating");
                updateData()
            }

            ws.current.onerror = (error) => {
                ws.current?.close()
                console.error("WebSocket error:", error);
            };

            ws.current.onclose = () => {
                console.log("WebSocket connection closed");
                ws.current = null;
                setTimeout(() => {
                    console.log("Reconnecting...");
                    connect()
                }, 10000)

            };
        }

        connect()

        return () => {
            if (ws.current) {
                ws.current.close()
            }
        };
    }, [updateData]);

    return (
        <Layout style={{ width: '100vw', height: '100vh' }}>
            <Header style={{ background: 'black', color: 'white', textAlign: "center", fontSize: "24px" }}>
                IP Table
            </Header>
            <Content style={{ width: '100vw', height: '100vh' }}>
                <Table
                    dataSource={tableData}
                    columns={columns}
                />
            </Content>
        </Layout>
    );
}

export default App;
