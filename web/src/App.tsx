import { Routes, Route } from 'react-router-dom'
import { Dashboard } from './pages/Dashboard'
import { SystemStatus } from './pages/SystemStatus'
import { Configuration } from './pages/Configuration'
import { Layout } from './components/Layout'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/status" element={<SystemStatus />} />
        <Route path="/config" element={<Configuration />} />
      </Routes>
    </Layout>
  )
}

export default App
