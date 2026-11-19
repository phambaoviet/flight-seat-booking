import { createContext, useContext, useState } from 'react';
import { loginUser, logoutUser } from '../services/authService';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() =>
    localStorage.getItem('access_token')
  );

  const login = async (credentials) => {
    const data = await loginUser(credentials);
    setToken(data.access_token);
    return data;
  };

  const logout = () => {
    logoutUser();
    setToken(null);
  };

  const isAuthenticated = Boolean(token);

  return (
    <AuthContext.Provider value={{ token, isAuthenticated, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
