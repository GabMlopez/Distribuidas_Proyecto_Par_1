'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

interface UserContextState {
  token: string | null;
  nickname: string;
  deviceId: string;
  userId: string;
  isAdmin: boolean;
}

interface UserContextType {
  userContext: UserContextState;
  login: (token: string | null, nickname: string, userId: string, isAdmin: boolean) => void;
  logout: () => void;
  setUserId: (userId: string) => void;
  updateToken: (token: string) => void; // Nuevo método
}

const UserContext = createContext<UserContextType | undefined>(undefined);

const STORAGE_KEYS = {
  NICKNAME: 'chat_nickname',
  USER_ID: 'chat_user_id',
  IS_ADMIN: 'chat_is_admin',
  TOKEN: 'chat_token'
};

export function UserProvider({ children }: { children: ReactNode }) {
  const [userContext, setUserContext] = useState<UserContextState>({
    token: null,
    nickname: '',
    deviceId: '',
    userId: '',
    isAdmin: false
  });
  useEffect(() => {
    const loadPersistedData = () => {
      const nickname = sessionStorage.getItem(STORAGE_KEYS.NICKNAME) || '';
      const userId = sessionStorage.getItem(STORAGE_KEYS.USER_ID) || '';
      const isAdmin = sessionStorage.getItem(STORAGE_KEYS.IS_ADMIN) === 'true';
      const token = sessionStorage.getItem(STORAGE_KEYS.TOKEN) || null;
      
      if (nickname) {
        setUserContext(prev => ({
          ...prev,
          nickname,
          userId,
          isAdmin,
          token
        }));
      }
    };
    
    loadPersistedData();
  }, []);

  useEffect(() => {
    // Hidratación controlada post-render
    setUserContext(prev => ({
      ...prev,
      token: sessionStorage.getItem('token'),
      nickname: sessionStorage.getItem('nickname') || '',
      deviceId: sessionStorage.getItem('deviceId') || '',
      userId: sessionStorage.getItem('userId') || '',
      isAdmin: sessionStorage.getItem('isAdmin') === 'true'
    }));
  }, []);

  useEffect(() => {
    const fetchDeviceId = async () => {
      try {
        let hardwareInfo = '';
        
        try {
            const canvas = document.createElement('canvas');
            const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
            if (gl) {
                const debugInfo = (gl as any).getExtension('WEBGL_debug_renderer_info');
                const vendor = debugInfo ? (gl as any).getParameter(debugInfo.UNMASKED_VENDOR_WEBGL) : 'unk_v';
                const renderer = debugInfo ? (gl as any).getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : 'unk_r';
                hardwareInfo += `${vendor}_${renderer}_`;
            }
        } catch (e) {}

        const cores = navigator.hardwareConcurrency || 'unk_c';
        const ram = (navigator as any).deviceMemory || 'unk_m';
        const platform = navigator.platform || 'unk_p';
        const screen = `${window.screen.width}x${window.screen.height}_${window.screen.colorDepth}`;

        hardwareInfo += `${cores}_${ram}_${platform}_${screen}`;

        let hash = 0;
        for (let i = 0; i < hardwareInfo.length; i++) {
            const char = hardwareInfo.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash;
        }

        const finalDeviceId = `hw_${Math.abs(hash)}`;

        setUserContext(prev => ({
          ...prev,
          deviceId: finalDeviceId
        }));
      } catch (error) {
        console.error("Error generando Hardware Fingerprint", error);
        setUserContext(prev => ({
          ...prev,
          deviceId: 'dev_' + Math.random().toString(36).substring(2, 11)
        }));
      }
    };
    fetchDeviceId();
  }, []);

  const login = (token: string | null, nickname: string, userId: string, isAdmin: boolean) => {
    sessionStorage.setItem(STORAGE_KEYS.NICKNAME, nickname);
    sessionStorage.setItem(STORAGE_KEYS.USER_ID, userId);
    sessionStorage.setItem(STORAGE_KEYS.IS_ADMIN, String(isAdmin));
    if (token) {
      sessionStorage.setItem(STORAGE_KEYS.TOKEN, token);
    }
    
    setUserContext(prev => ({ 
      ...prev, 
      token, 
      nickname, 
      userId, 
      isAdmin 
    }));
  };

  const logout = () => {
    sessionStorage.removeItem(STORAGE_KEYS.NICKNAME);
    sessionStorage.removeItem(STORAGE_KEYS.USER_ID);
    sessionStorage.removeItem(STORAGE_KEYS.IS_ADMIN);
    sessionStorage.removeItem(STORAGE_KEYS.TOKEN);
    
    setUserContext(prev => ({
      ...prev,
      token: null,
      nickname: '',
      userId: '',
      isAdmin: false
    }));
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('nickname');
    sessionStorage.removeItem('userId');
    sessionStorage.removeItem('isAdmin');
    sessionStorage.removeItem('room_name');
    sessionStorage.removeItem('room_type');
    sessionStorage.removeItem('room_token');
  };

  const setUserId = (userId: string) => {
    sessionStorage.setItem(STORAGE_KEYS.USER_ID, userId);
    setUserContext(prev => ({ ...prev, userId }));
  };
  
  const updateToken = (token: string) => {
    sessionStorage.setItem(STORAGE_KEYS.TOKEN, token);
    setUserContext(prev => ({ ...prev, token }));
  };

  return (
    <UserContext.Provider value={{ userContext, login, logout, setUserId, updateToken }}>
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