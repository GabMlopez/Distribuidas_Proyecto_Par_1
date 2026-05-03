'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface UserContextState {
  token: string | null;
  nickname: string;
  deviceId: string;
}

interface UserContextType {
  userContext: UserContextState;
  login: (token: string | null, nickname: string) => void;
  logout: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: ReactNode }) {
  const [userContext, setUserContext] = useState<UserContextState>({
    token: null,
    nickname: '',
    deviceId: ''
  });

  useEffect(() => {
    setUserContext(prev => ({
      ...prev,
      deviceId: 'dev_' + Math.random().toString(36).substring(2, 11)
    }));
  }, []);

  const login = (token: string | null, nickname: string) => {
    setUserContext(prev => ({ ...prev, token, nickname: nickname || 'Admin' }));
  };

  const logout = () => {
    setUserContext({
      token: null,
      nickname: '',
      deviceId: 'dev_' + Math.random().toString(36).substring(2, 11)
    });
  };

  return (
    <UserContext.Provider value={{ userContext, login, logout }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}
