import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './features/auth/useAuth';
import { ProtectedRoute } from './app/ProtectedRoute';
import { LoginPage } from './pages/Login';
import { HomePage } from './pages/Home';
import { CreateInitiative } from './pages/CreateInitiative';
import { InitiativeDetails } from './pages/InitiativeDetails';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          
          {/* Защищённые маршруты */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <HomePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/initiatives/new"
            element={
              <ProtectedRoute>
                <CreateInitiative />
              </ProtectedRoute>
            }
          />
          <Route
            path="/initiatives/:id"
            element={
              <ProtectedRoute>
                <InitiativeDetails />
              </ProtectedRoute>
            }
          />
          
          {/* Fallback */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;