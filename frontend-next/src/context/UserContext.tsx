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
  setNickname: (nickname: string) => void;
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
    // Hidratación controlada post-render (arregla el error rojo "Hydration failed")
    setUserContext(prev => ({
      ...prev,
      token: localStorage.getItem('token'),
      nickname: localStorage.getItem('nickname') || '',
      deviceId: localStorage.getItem('deviceId') || '',
      userId: localStorage.getItem('userId') || '',
      isAdmin: localStorage.getItem('isAdmin') === 'true'
    }));
  }, []);

  useEffect(() => {
    const fetchDeviceId = async () => {
      try {
        // ============================================================
        // HUELLA DE DISPOSITIVO ESTABLE (funciona igual en Incógnito)
        // ============================================================
        // Chrome bloquea WEBGL_debug_renderer_info en modo incógnito,
        // por eso usamos Canvas Fingerprinting + propiedades de hardware
        // que son idénticas en ambos modos.
        // ============================================================
        let hardwareInfo = '';

        // 1. Canvas Fingerprint: dibujamos texto y formas en un canvas.
        //    El resultado depende de la GPU, motor de renderizado de fuentes
        //    y SO — pero NO cambia entre modo normal e incógnito.
        try {
          const canvas = document.createElement('canvas');
          canvas.width = 260;
          canvas.height = 60;
          const ctx = canvas.getContext('2d');
          if (ctx) {
            ctx.textBaseline = 'top';
            // Bloque de color
            ctx.font = '16px Arial';
            ctx.fillStyle = '#f60';
            ctx.fillRect(125, 1, 62, 20);
            // Texto con sombra
            ctx.fillStyle = '#069';
            ctx.fillText('DeviceFingerprint!@#', 2, 15);
            ctx.fillStyle = 'rgba(102, 204, 0, 0.7)';
            ctx.fillText('DeviceFingerprint!@#', 4, 17);
            // Arco
            ctx.beginPath();
            ctx.arc(50, 50, 10, 0, Math.PI * 2, true);
            ctx.closePath();
            ctx.fill();
            // Convertir a Data URL (la salida es determinista para el mismo HW)
            hardwareInfo += canvas.toDataURL();
          }
        } catch (e) { /* canvas no disponible */ }

        // 2. Propiedades de hardware (idénticas en incógnito)
        const cores = navigator.hardwareConcurrency || 0;
        const ram = (navigator as any).deviceMemory || 0;
        const platform = navigator.platform || '';
        const screenInfo = `${window.screen.width}x${window.screen.height}_${window.screen.colorDepth}`;
        const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || '';

        hardwareInfo += `|${cores}_${ram}_${platform}_${screenInfo}_${timezone}`;

        // 3. Generar hash estable de 53 bits (mejor distribución que 32-bit)
        let h1 = 0xdeadbeef;
        let h2 = 0x41c6ce57;
        for (let i = 0; i < hardwareInfo.length; i++) {
          const ch = hardwareInfo.charCodeAt(i);
          h1 = Math.imul(h1 ^ ch, 2654435761);
          h2 = Math.imul(h2 ^ ch, 1597334677);
        }
        h1 = Math.imul(h1 ^ (h1 >>> 16), 2246822507);
        h1 ^= Math.imul(h2 ^ (h2 >>> 13), 3266489909);
        h2 = Math.imul(h2 ^ (h2 >>> 16), 2246822507);
        h2 ^= Math.imul(h1 ^ (h1 >>> 13), 3266489909);
        const fullHash = 4294967296 * (2097151 & h2) + (h1 >>> 0);

        const finalDeviceId = `hw_${fullHash}`;
        localStorage.setItem('deviceId', finalDeviceId);
        sessionStorage.setItem('deviceId', finalDeviceId);

        // Log para depuración (puedes verificar que sea idéntico en incógnito)
        console.log('[DeviceFingerprint] ID generado:', finalDeviceId);

        setUserContext(prev => ({
          ...prev,
          deviceId: finalDeviceId,
        }));
      } catch (error) {
        console.error("Error generando Hardware Fingerprint", error);
        const fallbackId = 'dev_' + Math.random().toString(36).substring(2, 11);
        localStorage.setItem('deviceId', fallbackId);
        setUserContext(prev => ({
          ...prev,
          deviceId: fallbackId,
        }));
      }
    };
    fetchDeviceId();
  }, []);

  const login = (token: string | null, nickname: string, userId: string, isAdmin: boolean) => {
    setUserContext(prev => ({ ...prev, token, nickname: nickname , userId, isAdmin }));
    localStorage.setItem('token', token || '');
    localStorage.setItem('nickname', nickname);
    localStorage.setItem('userId', userId);
    localStorage.setItem('isAdmin', String(isAdmin));
  };

  const logout = () => {
    setUserContext(prev => ({
      ...prev,
      token: null,
      nickname: '',
      userId: '',
      isAdmin: false
    }));
    localStorage.removeItem('token');
    localStorage.removeItem('nickname');
    localStorage.removeItem('userId');
    localStorage.removeItem('isAdmin');
    localStorage.removeItem('room_name');
    localStorage.removeItem('room_type');
    localStorage.removeItem('room_token');
  };

  const setUserId = (userId: string) => {
    setUserContext(prev => ({ ...prev, userId }));
    localStorage.setItem('userId', userId);  
  };

  const setNickname = (nickname: string) => {
    setUserContext(prev => ({ ...prev, nickname }));
    localStorage.setItem('nickname', nickname);
  };

   return (
    <UserContext.Provider value={{ userContext, login, logout, setUserId, setNickname }}>
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