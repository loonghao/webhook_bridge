import { Routes, Route } from 'react-router-dom'
import { Dashboard } from './pages/Dashboard'
import { SystemStatus } from './pages/SystemStatus'
import { Configuration } from './pages/Configuration'
import { Logs } from './pages/Logs'
import { PythonInterpreters } from './pages/PythonInterpreters'
import { ConnectionStatus } from './pages/ConnectionStatus'
import { ApiTest } from './pages/ApiTest'
import { Plugins } from './pages/Plugins'
import { PluginManager } from './pages/PluginManager'
import { NotFound } from './pages/NotFound'
import { Layout } from './components/Layout'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/status" element={<SystemStatus />} />
        <Route path="/plugins" element={<Plugins />} />
        <Route path="/plugin-manager" element={<PluginManager />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/config" element={<Configuration />} />
        <Route path="/interpreters" element={<PythonInterpreters />} />
        <Route path="/connection" element={<ConnectionStatus />} />
        <Route path="/api-test" element={<ApiTest />} />
        <Route path="*" element={<NotFound />} />
      </Routes>
    </Layout>
  )
}

export default App
