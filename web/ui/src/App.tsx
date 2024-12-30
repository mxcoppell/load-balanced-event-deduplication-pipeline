import { useState, useEffect } from 'react'
import './App.css'

interface TestConfig {
  num_keys: number
  key_delay: number  // in milliseconds
  key_ttl: number   // in milliseconds
  dedup_window: number  // in milliseconds
}

interface TestStatus {
  is_running: boolean
  generated: number
  consumed: number
}

function App() {
  const [config, setConfig] = useState<TestConfig>({
    num_keys: 1000,
    key_delay: 1,
    key_ttl: 100,
    dedup_window: 5000
  })

  const [status, setStatus] = useState<TestStatus>({
    is_running: false,
    generated: 0,
    consumed: 0
  })

  useEffect(() => {
    const interval = setInterval(fetchStatus, 1000)
    return () => clearInterval(interval)
  }, [])

  const fetchStatus = async () => {
    try {
      const response = await fetch('/api/status')
      const data = await response.json()
      setStatus(data)
    } catch (error) {
      console.error('Failed to fetch status:', error)
    }
  }

  const startTest = async () => {
    try {
      await fetch('/api/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(config)
      })
      fetchStatus()
    } catch (error) {
      console.error('Failed to start test:', error)
    }
  }

  const stopTest = async () => {
    try {
      await fetch('/api/stop', {
        method: 'POST'
      })
      fetchStatus()
    } catch (error) {
      console.error('Failed to stop test:', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100 py-8">
      <div className="container mx-auto px-4">
        <h1 className="text-3xl font-bold mb-8 text-gray-800">Load-Balanced Event Deduplication Pipeline</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          <div className="bg-white p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold mb-4 text-gray-800">Test Configuration</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Number of Keys</label>
                <input
                  type="number"
                  value={config.num_keys}
                  onChange={(e) => setConfig({ ...config, num_keys: parseInt(e.target.value) })}
                  className="w-full p-2 border rounded bg-white text-gray-800"
                  min="1"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Key Delay (ms)</label>
                <input
                  type="number"
                  value={config.key_delay}
                  onChange={(e) => setConfig({ ...config, key_delay: parseInt(e.target.value) })}
                  className="w-full p-2 border rounded bg-white text-gray-800"
                  min="1"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Key TTL (ms)</label>
                <input
                  type="number"
                  value={config.key_ttl}
                  onChange={(e) => setConfig({ ...config, key_ttl: parseInt(e.target.value) })}
                  className="w-full p-2 border rounded bg-white text-gray-800"
                  min="1"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Dedup Window (ms)</label>
                <input
                  type="number"
                  value={config.dedup_window}
                  onChange={(e) => setConfig({ ...config, dedup_window: parseInt(e.target.value) })}
                  className="w-full p-2 border rounded bg-white text-gray-800"
                  min="1"
                />
              </div>
              <div className="pt-4">
                {!status.is_running ? (
                  <button
                    onClick={startTest}
                    className="w-full bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 font-medium"
                  >
                    Start Test
                  </button>
                ) : (
                  <button
                    onClick={stopTest}
                    className="w-full bg-red-500 text-white py-2 px-4 rounded hover:bg-red-600 font-medium"
                  >
                    Stop Test
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold mb-4 text-gray-800">Test Metrics</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Status</label>
                <div className="text-lg font-semibold">
                  {status.is_running ? (
                    <span className="text-green-600">Running</span>
                  ) : (
                    <span className="text-gray-500">Stopped</span>
                  )}
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Generated Keys</label>
                <div className="text-lg font-semibold text-gray-800">{status.generated}</div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Consumed Keys</label>
                <div className="text-lg font-semibold text-gray-800">{status.consumed}</div>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1 text-gray-700">Progress</label>
                <div className="relative pt-1">
                  <div className="overflow-hidden h-2 text-xs flex rounded bg-gray-200">
                    <div
                      style={{ width: `${(status.consumed / config.num_keys) * 100}%` }}
                      className="shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center bg-blue-500 transition-all duration-300"
                    />
                  </div>
                  <div className="text-xs font-medium text-gray-600 mt-1">
                    {Math.round((status.consumed / config.num_keys) * 100)}%
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
