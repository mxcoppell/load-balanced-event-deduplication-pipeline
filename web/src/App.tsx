import { useState, useEffect } from 'react';
import axios from 'axios';

interface TestConfig {
    num_keys: number;
    key_delay: number;
    key_ttl: number;
    dedup_window: number;
}

interface TestStatus {
    is_running: boolean;
    generated: number;
    consumed: number;
}

interface TestMetrics {
    generated: number;
    consumed: number;
    consumers: { [key: string]: number };
}

export default function App() {
    const [config, setConfig] = useState<TestConfig>({
        num_keys: 1000,
        key_delay: 10,
        key_ttl: 5000,
        dedup_window: 5000,
    });

    const [status, setStatus] = useState<TestStatus>({
        is_running: false,
        generated: 0,
        consumed: 0,
    });

    const [metrics, setMetrics] = useState<TestMetrics>({
        generated: 0,
        consumed: 0,
        consumers: {},
    });

    useEffect(() => {
        const interval = setInterval(() => {
            fetchStatus();
            fetchMetrics();
        }, 1000);
        return () => clearInterval(interval);
    }, []);

    const fetchStatus = async () => {
        try {
            const response = await axios.get<TestStatus>('/api/status');
            setStatus(response.data);
        } catch (error) {
            console.error('Failed to fetch status:', error);
        }
    };

    const fetchMetrics = async () => {
        try {
            const response = await axios.get<TestMetrics>('/api/metrics');
            setMetrics(response.data);
        } catch (error) {
            console.error('Failed to fetch metrics:', error);
        }
    };

    const startTest = async () => {
        try {
            await axios.post('/api/start', config);
        } catch (error) {
            console.error('Failed to start test:', error);
        }
    };

    const stopTest = async () => {
        try {
            await axios.post('/api/stop');
        } catch (error) {
            console.error('Failed to stop test:', error);
        }
    };

    return (
        <div className="container mx-auto p-4">
            <h1 className="text-2xl font-bold mb-4">Key Expiration Test</h1>

            <div className="bg-white rounded-lg shadow p-4 mb-4">
                <h2 className="text-xl font-semibold mb-2">About This Test</h2>
                <p className="text-gray-700 mb-2">
                    This test simulates key expiration events in a distributed system using Redis. The system consists of:
                </p>
                <ul className="list-disc list-inside text-gray-700 mb-2 ml-4">
                    <li>A Generator service that creates keys with specified TTLs</li>
                    <li>Multiple Consumer services that process key expiration events</li>
                    <li>Redis for key storage and expiration notifications</li>
                    <li>NATS for message distribution</li>
                </ul>
                <p className="text-gray-700">
                    The goal is to test the system's ability to handle key expiration events efficiently and reliably across multiple consumers,
                    while preventing duplicate processing through a deduplication mechanism.
                </p>
            </div>

            <div className="bg-white rounded-lg shadow p-4 mb-4">
                <h2 className="text-xl font-semibold mb-2">Configuration</h2>
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <label className="block text-gray-700 mb-1">Number of Keys:</label>
                        <input
                            type="number"
                            value={config.num_keys}
                            onChange={(e) => setConfig({ ...config, num_keys: parseInt(e.target.value) })}
                            className="w-full px-3 py-2 border rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-gray-700 mb-1">Key Delay (ms):</label>
                        <input
                            type="number"
                            value={config.key_delay}
                            onChange={(e) => setConfig({ ...config, key_delay: parseInt(e.target.value) })}
                            className="w-full px-3 py-2 border rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-gray-700 mb-1">Key TTL (ms):</label>
                        <input
                            type="number"
                            value={config.key_ttl}
                            onChange={(e) => setConfig({ ...config, key_ttl: parseInt(e.target.value) })}
                            className="w-full px-3 py-2 border rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-gray-700 mb-1">Dedup Window (ms):</label>
                        <input
                            type="number"
                            value={config.dedup_window}
                            onChange={(e) => setConfig({ ...config, dedup_window: parseInt(e.target.value) })}
                            className="w-full px-3 py-2 border rounded"
                        />
                    </div>
                </div>
            </div>

            <div className="bg-white rounded-lg shadow p-4 mb-4">
                <h2 className="text-xl font-semibold mb-2">Status</h2>
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <p className="text-gray-700">Status: <span className="font-semibold">{status.is_running ? 'Running' : 'Stopped'}</span></p>
                        <p className="text-gray-700">Generated Keys: <span className="font-semibold">{status.generated}</span></p>
                        <p className="text-gray-700">Consumed Keys: <span className="font-semibold">{status.consumed}</span></p>
                    </div>
                    <div>
                        <button
                            onClick={status.is_running ? stopTest : startTest}
                            className={`px-4 py-2 rounded ${status.is_running ? 'bg-red-500 hover:bg-red-600' : 'bg-blue-500 hover:bg-blue-600'} text-white`}
                        >
                            {status.is_running ? 'Stop Test' : 'Start Test'}
                        </button>
                    </div>
                </div>
            </div>

            <div className="bg-white rounded-lg shadow p-4">
                <h2 className="text-xl font-semibold mb-2">Consumer Metrics</h2>
                <div className="overflow-x-auto">
                    <table className="min-w-full table-auto">
                        <thead>
                            <tr className="bg-gray-100">
                                <th className="px-4 py-2 text-left">Consumer ID</th>
                                <th className="px-4 py-2 text-left">Events Processed</th>
                            </tr>
                        </thead>
                        <tbody>
                            {Object.entries(metrics.consumers).map(([id, count]) => (
                                <tr key={id} className="border-t">
                                    <td className="px-4 py-2">{id}</td>
                                    <td className="px-4 py-2">{count}</td>
                                </tr>
                            ))}
                            <tr className="border-t bg-gray-50 font-semibold">
                                <td className="px-4 py-2">Total Consumed Events</td>
                                <td className="px-4 py-2">{Object.values(metrics.consumers).reduce((sum, count) => sum + count, 0)}</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
} 