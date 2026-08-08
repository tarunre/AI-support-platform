import { Routes, Route, Navigate } from 'react-router-dom'
import MainLayout from '../layouts/MainLayout'
import AppLayout from '../layouts/AppLayout'
import AgentLayout from '../layouts/AgentLayout'
import Home from '../pages/Home'
import Login from '../pages/Login'
import Register from '../pages/Register'
import AgentRegister from '../pages/AgentRegister'
import Dashboard from '../pages/Dashboard'
import TicketList from '../pages/tickets/TicketList'
import CreateTicket from '../pages/tickets/CreateTicket'
import TicketDetails from '../pages/tickets/TicketDetails'
import EditTicket from '../pages/tickets/EditTicket'
import MyTickets from '../pages/tickets/MyTickets'
import UnassignedTickets from '../pages/tickets/UnassignedTickets'
import KnowledgeBase from '../pages/KnowledgeBase'
import JobMonitor from '../pages/JobMonitor'
import EmailAccounts from '../pages/EmailAccounts'
import EmailMonitor from '../pages/EmailMonitor'
import AnalyticsDashboard from '../pages/analytics/AnalyticsDashboard'
import AgentPerformance from '../pages/analytics/AgentPerformance'
import Reports from '../pages/analytics/Reports'
import Integrations from '../pages/Integrations'
import TenantSettings from '../pages/TenantSettings'
import SuperAdminDashboard from '../pages/superadmin/SuperAdminDashboard'
import SLAManagement from '../pages/SLAManagement'
import SLADashboard from '../pages/SLADashboard'
import AgentDashboard from '../pages/agent/AgentDashboard'
import AgentQueue from '../pages/agent/AgentQueue'
import AgentAvailable from '../pages/agent/AgentAvailable'
import AgentTicketDetail from '../pages/agent/AgentTicketDetail'
import CustomerPortal from '../pages/CustomerPortal'

function AppRoutes() {
  return (
    <Routes>
      <Route element={<MainLayout />}>
        <Route path="/" element={<Home />} />
      </Route>

      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/agent-register" element={<AgentRegister />} />
      <Route path="/portal" element={<CustomerPortal />} />

      {/* Agent portal — only accessible to SupportAgent role */}
      <Route element={<AgentLayout />}>
        <Route path="/agent" element={<AgentDashboard />} />
        <Route path="/agent/queue" element={<AgentQueue />} />
        <Route path="/agent/available" element={<AgentAvailable />} />
        <Route path="/agent/tickets/:id" element={<AgentTicketDetail />} />
      </Route>

      {/* Customer portal — no auth wrapper, magic-link token only */}
      <Route path="/portal" element={<CustomerPortal />} />

      <Route element={<AppLayout />}>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/my-tickets" element={<MyTickets />} />
        <Route path="/tickets/unassigned" element={<UnassignedTickets />} />
        <Route path="/knowledge-base" element={<KnowledgeBase />} />
        <Route path="/jobs" element={<JobMonitor />} />
        <Route path="/email/accounts" element={<EmailAccounts />} />
        <Route path="/email/monitor" element={<EmailMonitor />} />
        <Route path="/analytics" element={<AnalyticsDashboard />} />
        <Route path="/analytics/agents" element={<AgentPerformance />} />
        <Route path="/analytics/reports" element={<Reports />} />
        <Route path="/tickets" element={<TicketList />} />
        <Route path="/tickets/new" element={<CreateTicket />} />
        <Route path="/tickets/:id" element={<TicketDetails />} />
        <Route path="/tickets/:id/edit" element={<EditTicket />} />
        <Route path="/integrations" element={<Integrations />} />
        <Route path="/settings" element={<TenantSettings />} />
        <Route path="/admin" element={<SuperAdminDashboard />} />
        <Route path="/sla" element={<SLADashboard />} />
        <Route path="/sla-management" element={<SLAManagement />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default AppRoutes
