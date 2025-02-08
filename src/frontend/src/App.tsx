import { Layout, Table } from "antd";
import { ColumnType } from "antd/es/table";
import { useEffect, useState } from "react";

const { Header, Content } = Layout;

type IpData = {
    ip: string,
    ping_ms: number,
    pinged_at: string
}

const API_PING_DATA_URL = "http://localhost:4242/events"
const API_WS_URL = "ws://localhost:4242/ws"

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
    const updateData = () => {
        fetch(API_PING_DATA_URL).then(resp => resp.json()).then(json => {
            setTableData(parseIpData(json))
        }).catch(e => console.log(e))
    }

    useEffect(() => {
        updateData()
    }, [])

    useEffect(() => {
        let ws: WebSocket | null = null;

        const connect = () => {
            if (ws) {
                ws.close()
            }

            ws = new WebSocket(API_WS_URL)

            ws.onopen = () => {
                console.log("Connected to WebSocket");
            };

            ws.onmessage = () => {
                console.log("Updating");
                updateData()
            }

            ws.onerror = (error) => {
                ws?.close()
                console.error("WebSocket error:", error);
            };

            ws.onclose = () => {
                console.log("WebSocket connection closed");
                ws = null;
                setTimeout(() => {
                    console.log("Reconnecting...");
                    connect()
                }, 10000)

            };
        }

        connect()

        return () => {
            if (ws) {
                ws.close()
            }
        };
    }, []);

    const columns: ColumnType[] = [
        { title: "IP", dataIndex: "ip" },
        { title: "Ping ms", dataIndex: "ping_ms" },
        { title: "Ping date", dataIndex: "pinged_at" },
    ];

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
