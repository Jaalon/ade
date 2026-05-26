import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Projects from './pages/Projects'
import ProjectDetail from './pages/ProjectDetail'
import Plugins from './pages/Plugins'
import PluginDetail from './pages/PluginDetail'
import Workflows from './pages/Workflows'
import WorkflowDetail from './pages/WorkflowDetail'
import Reports from './pages/Reports'
import ReportDetail from './pages/ReportDetail'

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/projects" element={<Projects />} />
        <Route path="/projects/:name" element={<ProjectDetail />} />
        <Route path="/plugins" element={<Plugins />} />
        <Route path="/plugins/:name" element={<PluginDetail />} />
        <Route path="/workflows" element={<Workflows />} />
        <Route path="/workflows/:id" element={<WorkflowDetail />} />
        <Route path="/reports" element={<Reports />} />
        <Route path="/reports/:id" element={<ReportDetail />} />
      </Routes>
    </Layout>
  )
}
