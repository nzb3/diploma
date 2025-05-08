import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ChatProvider, AuthProvider, useAuth } from '@/context';
import { MainLayout } from '@layouts';
import LandingPage from '@pages/Landing';
import NotFoundPage from '@pages/NotFound';
import UnauthorizedPage from '@pages/Unauthorized';
import SearchPage from '@pages/Search';
import ResourcesPage from '@pages/Resources';
import React from 'react';

function AuthRoute({ element }: { element: React.ReactElement }) {
  const { isAuthenticated, isInitialized } = useAuth();
  
  if (isInitialized) {
    if (!isAuthenticated) {
      return <UnauthorizedPage />;
    }

    return element;
  }
}

function RootRoute() {
  const { isAuthenticated, isInitialized } = useAuth();
  
  if (isInitialized) {
    return isAuthenticated ? (
        <MainLayout>
          <SearchPage />
        </MainLayout>
    ) : (
        <LandingPage />
    );
  }
}

function App() {
  return (
    <AuthProvider>
      <ChatProvider>
        <Router>
          <Routes>
            {/* Root path - show LandingPage or SearchPage based on auth status */}
            <Route path="/" element={<RootRoute />} />
            
            {/* Protected routes - show content if authenticated, otherwise show Unauthorized */}
            <Route 
              path="/resources" 
              element={
                <AuthRoute 
                  element={
                    <MainLayout>
                      <ResourcesPage />
                    </MainLayout>
                  } 
                />
              } 
            />
            
            {/* Any undefined route shows NotFound without redirect */}
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Router>
      </ChatProvider>
    </AuthProvider>
  );
}

export default App;
