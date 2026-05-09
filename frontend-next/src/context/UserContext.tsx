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
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: ReactNode }) {
  const [userContext, setUserContext] = useState<UserContextState>({
    token: null,
    nickname: '',
    deviceId: '',
    userId: '',
    isAdmin: false
  });

  useEffect(() => {
    const fetchDeviceId = async () => {
      try {
        // Generar una Huella de Hardware (Hard Fingerprint) estricta
        // Esto NO depende del almacenamiento del navegador, por lo que será idéntico en Incógnito.
        let hardwareInfo = '';
        
        // 1. Extraer el modelo exacto de la Tarjeta Gráfica (GPU) mediante WebGL
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

        // 2. Extraer CPU, Memoria RAM y Sistema Operativo
        const cores = navigator.hardwareConcurrency || 'unk_c';
        const ram = (navigator as any).deviceMemory || 'unk_m';
        const platform = navigator.platform || 'unk_p';
        
        // 3. Extraer Resolución de Pantalla y Profundidad de Color
        const screen = `${window.screen.width}x${window.screen.height}_${window.screen.colorDepth}`;

        // Unir toda la información del hardware físico
        hardwareInfo += `${cores}_${ram}_${platform}_${screen}`;

        // Convertir la cadena de hardware en un Hash numérico simple (32-bit)
        let hash = 0;
        for (let i = 0; i < hardwareInfo.length; i++) {
            const char = hardwareInfo.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash;
        }

        const finalDeviceId = `hw_${Math.abs(hash)}`;

        setUserContext(prev => ({
          ...prev,
          deviceId: finalDeviceId,
          isAdmin: true
        }));
      } catch (error) {
        console.error("Error generando Hardware Fingerprint", error);
        setUserContext(prev => ({
          ...prev,
          deviceId: 'dev_' + Math.random().toString(36).substring(2, 11),
          isAdmin: true  
        }));
      }
    };
    fetchDeviceId();
  }, []);

  const login = (token: string | null, nickname: string, userId: string, isAdmin: boolean) => {
    setUserContext(prev => ({ ...prev, token, nickname: nickname , userId, isAdmin }));
  };

  const logout = () => {
    setUserContext(prev => ({
      ...prev,
      token: null,
      nickname: '',
      userId: '',
      isAdmin: false
    }));
  };

  const setUserId = (userId: string) => {
    setUserContext(prev => ({ ...prev, userId }));
    sessionStorage.setItem('user_id', userId);  
  };

   return (
    <UserContext.Provider value={{ userContext, login, logout, setUserId }}>
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
