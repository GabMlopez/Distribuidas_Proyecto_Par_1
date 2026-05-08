'use client';

import { useEffect, useState } from 'react';
import { UploadTask } from '@/types';

interface UploadToastProps {
  uploads: UploadTask[];
  onRemove: (id: string) => void;
}

export function UploadToast({ uploads, onRemove }: UploadToastProps) {
  const [visibleToasts, setVisibleToasts] = useState<UploadTask[]>([]);

  useEffect(() => {
    const newUploads = uploads.filter(u => !visibleToasts.find(v => v.id === u.id));
    if (newUploads.length > 0) {
      setVisibleToasts(prev => [...prev, ...newUploads]);
    }
    
    const updated = visibleToasts.map(toast => {
      const updatedUpload = uploads.find(u => u.id === toast.id);
      return updatedUpload || toast;
    });
    
    if (JSON.stringify(visibleToasts) !== JSON.stringify(updated)) {
      setVisibleToasts(updated);
    }
  }, [uploads]);

  useEffect(() => {
    const timeouts: NodeJS.Timeout[] = [];
    
    visibleToasts.forEach(toast => {
      if (toast.done && !toast.error) {
        const timeout = setTimeout(() => {
          setVisibleToasts(prev => prev.filter(t => t.id !== toast.id));
          onRemove(toast.id);
        }, 2000);
        timeouts.push(timeout);
      } else if (toast.error) {
        const timeout = setTimeout(() => {
          setVisibleToasts(prev => prev.filter(t => t.id !== toast.id));
          onRemove(toast.id);
        }, 4000);
        timeouts.push(timeout);
      }
    });
    
    return () => timeouts.forEach(clearTimeout);
  }, [visibleToasts]);

  if (visibleToasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {visibleToasts.map((toast) => (
        <div
          key={toast.id}
          className="animate-slide-up bg-slate-900/95 backdrop-blur-md border border-white/20 rounded-xl shadow-2xl p-3 min-w-[280px] max-w-[320px]"
        >
          <div className="flex items-start gap-3">
            {/* Icono */}
            <div className="shrink-0">
              {toast.error ? (
                <div className="w-8 h-8 rounded-lg bg-rose-500/20 flex items-center justify-center">
                  <svg className="w-4 h-4 text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="12" y1="8" x2="12" y2="12"/>
                    <line x1="12" y1="16" x2="12.01" y2="16"/>
                  </svg>
                </div>
              ) : toast.done ? (
                <div className="w-8 h-8 rounded-lg bg-emerald-500/20 flex items-center justify-center">
                  <svg className="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
              ) : (
                <div className="w-8 h-8 rounded-lg bg-indigo-500/20 flex items-center justify-center">
                  <div className="w-4 h-4 border-2 border-indigo-400 border-t-transparent rounded-full animate-spin" />
                </div>
              )}
            </div>
            
            {/* Contenido */}
            <div className="flex-1">
              <p className="text-sm font-medium text-white">
                {toast.error 
                  ? 'Error al subir archivo'
                  : toast.done 
                  ? 'Archivo subido'
                  : `Subiendo... ${toast.progress}%`
                }
              </p>
              <p className="text-xs text-slate-400 truncate">{toast.name}</p>
              {toast.error && (
                <p className="text-xs text-rose-400/80 mt-1 break-words">{toast.error}</p>
              )}
              {!toast.done && !toast.error && toast.progress > 0 && (
                <div className="mt-1 h-1 bg-white/10 rounded-full overflow-hidden">
                  <div 
                    className="h-full rounded-full bg-gradient-to-r from-indigo-500 to-purple-600 transition-all duration-300"
                    style={{ width: `${toast.progress}%` }}
                  />
                </div>
              )}
            </div>
            
            {/* Botón cerrar */}
            <button 
              onClick={() => {
                setVisibleToasts(prev => prev.filter(t => t.id !== toast.id));
                onRemove(toast.id);
              }}
              className="shrink-0 w-6 h-6 rounded-lg bg-white/5 hover:bg-white/10 text-slate-400 hover:text-white transition-all flex items-center justify-center"
            >
              ✕
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}