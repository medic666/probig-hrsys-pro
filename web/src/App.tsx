import { Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import PersonList from './pages/Person';
import PersonDetail from './pages/Person/Detail';
import OrganizationList from './pages/Organization';
import OrganizationDetail from './pages/Organization/Detail';
import Attendance from './pages/Attendance';
import Salary from './pages/Salary';
import FileManagement from './pages/File';
import Audit from './pages/Audit';
import { useAuth } from './hooks/useAuth';

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  if (loading) return null;
  if (!user) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

export default function App() {
  const { user, loading } = useAuth();

  if (loading) return null;

  return (
    <Routes>
      <Route path="/login" element={user ? <Navigate to="/" replace /> : <Login />} />
      <Route
        path="/"
        element={
          <PrivateRoute>
            <MainLayout />
          </PrivateRoute>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="person" element={<PersonList />} />
        <Route path="person/:id" element={<PersonDetail />} />
        <Route path="organization" element={<OrganizationList />} />
        <Route path="organization/:id" element={<OrganizationDetail />} />
        <Route path="attendance" element={<Attendance />} />
        <Route path="salary" element={<Salary />} />
        <Route path="file" element={<FileManagement />} />
        <Route path="audit" element={<Audit />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
