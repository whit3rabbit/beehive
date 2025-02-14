// contexts/AuthContext.tsx
import { createContext, useContext, useState } from 'react';

interface AuthContextType {
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState<string | null>(null);

  return (
    <AuthContext.Provider value={{
      token,
      login: (newToken) => setToken(newToken),
      logout: () => setToken(null)
    }}>
      {children}
    </AuthContext.Provider>
  );
};
