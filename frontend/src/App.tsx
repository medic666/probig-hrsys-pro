import { Routes, Route, Navigate } from 'react-router-dom'
import AppLayout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Persons from './pages/Persons'
import Policies from './pages/Policies'
import Attendance from './pages/Attendance'
import Salary from './pages/Salary'
import Assets from './pages/Assets'
import Events from './pages/Events'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/*"
        element={
          <PrivateRoute>
            <AppLayout>
              <Routes>
                <Route path="/" element={<Dashboard />} />
                <Route path="/persons" element={<Persons />} />
                <Route path="/policies" element={<Policies />} />
                <Route path="/attendance" element={<Attendance />} />
                <Route path="/salary" element={<Salary />} />
                <Route path="/assets" element={<Assets />} />
                <Route path="/events" element={<Events />} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </AppLayout>
          </PrivateRoute>
        }
      />
    </Routes>
  )
}

export default App
