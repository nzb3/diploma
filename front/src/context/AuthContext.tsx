import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import authService from '../services/authService';

type AuthContextType = {
  isAuthenticated: boolean;
  isInitialized: boolean;
  username: string | undefined;
  login: () => void;
  logout: () => void;
  register: () => void;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

type AuthProviderProps = {
  children: ReactNode;
};

export function AuthProvider({ children }: AuthProviderProps) {
  const [isInitialized, setIsInitialized] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [username, setUsername] = useState<string | undefined>(undefined);

  useEffect(() => {
    const initAuth = async () => {
      try {
        const authenticated = await authService.init();
        setIsAuthenticated(authenticated);
        
        if (authenticated) {
          setUsername(authService.getUsername());
        }
      } catch (error) {
        console.error('Failed to initialize authentication', error);
      } finally {
        setIsInitialized(true);
      }
    };

    initAuth();
  }, []);

  useEffect(() => {
    if (isAuthenticated) {
      const interval = setInterval(() => {
        authService.updateToken(70).catch(() => {
          setIsAuthenticated(false);
        });
      }, 60000);

      return () => clearInterval(interval);
    }
  }, [isAuthenticated]);

  const login = () => {
    authService.login();
  };

  const logout = () => {
    authService.logout();
  };

  const register = () => {
    authService.register();
  };

  return (
    <AuthContext.Provider 
      value={{ 
        isAuthenticated, 
        isInitialized, 
        username, 
        login, 
        logout,
        register
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}; 