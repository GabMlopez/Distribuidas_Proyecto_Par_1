// components/room/ChatMessage.tsx
/* eslint-disable @next/next/no-img-element */
'use client';

import { useState } from 'react';
import { ChatMessage as MsgType } from '@/types';
import { SystemMessage } from './SystemMessage';

interface ChatMessageProps {
  msg: MsgType;
  isOwn: boolean;
  formatTime: (ts?: number) => string;
  api: string;
  onDeleteFile?: (fileUrl: string) => void;
}

export function ChatMessage({ msg, isOwn, formatTime, api, onDeleteFile }: ChatMessageProps) {
  const [imageError, setImageError] = useState(false);
  const [isImageLoaded, setIsImageLoaded] = useState(false);
  const [showFullImage, setShowFullImage] = useState(false);

  // Si es mensaje de sistema, mostrar componente especial
  if (msg.type === 'system' || msg.tipo === 'system') {
    return (
      <SystemMessage 
        text={msg.texto || ''} 
        timestamp={msg.timestamp} 
        formatTime={formatTime} 
      />
    );
  }

  const isMultimedia = msg.tipo === 'multimedia' || msg.type === 'multimedia';
  const fileUrl = msg.file_url;
  
  const isImage = fileUrl && /\.(jpg|jpeg|png|gif|webp|bmp|svg)$/i.test(fileUrl);
  const isPDF = fileUrl && /\.pdf$/i.test(fileUrl);

  const fullImageUrl = fileUrl ? `${api}${fileUrl}` : '';

  // Si es mensaje de join/leave antiguo (por si acaso)
  if (msg.tipo === 'join' || msg.type === 'join') {
    return (
      <SystemMessage 
        text={msg.texto || `${msg.nickname} se ha unido a la sala`} 
        timestamp={msg.timestamp} 
        formatTime={formatTime} 
      />
    );
  }

  if (msg.tipo === 'leave' || msg.type === 'leave') {
    return (
      <SystemMessage 
        text={msg.texto || `${msg.nickname} ha salido de la sala`} 
        timestamp={msg.timestamp} 
        formatTime={formatTime} 
      />
    );
  }

  // Resto del código para mensajes normales...
  return (
    <div className={`flex ${isOwn ? 'justify-end' : 'justify-start'} animate-fade-in`}>
      <div className={`max-w-[75%] ${isOwn ? 'items-end' : 'items-start'} flex flex-col`}>
        
        {!isOwn && (
          <span className="text-xs text-indigo-300/70 mb-1 ml-2">{msg.nickname}</span>
        )}

        <div className={`relative ${isOwn ? 'bg-indigo-600/30 border-indigo-400/30' : 'bg-white/5 border-white/15'} border rounded-2xl ${isOwn ? 'rounded-tr-md' : 'rounded-tl-md'} overflow-hidden`}>
          
          <div className="px-4 py-2.5">
            {msg.texto && (
              <p className="text-sm text-slate-100 whitespace-pre-wrap break-words">
                {msg.texto}
              </p>
            )}

            {fileUrl && (
              <div className="mt-2 relative group">
                {/* El botón global absoluto fue eliminado en favor de botones específicos por tipo de archivo */}

                {isImage && !imageError && (
                  <div className="relative group/img inline-block">
                    {isOwn && onDeleteFile && (
                      <button 
                        onClick={() => onDeleteFile(fileUrl)}
                        title="Eliminar imagen"
                        className="absolute top-2 right-2 z-10 w-6 h-6 bg-black/50 hover:bg-red-500/90 text-white rounded-full flex items-center justify-center opacity-0 group-hover/img:opacity-100 transition-all backdrop-blur-sm shadow-sm"
                      >
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                          <line x1="18" y1="6" x2="6" y2="18"></line>
                          <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                      </button>
                    )}
                    {!isImageLoaded && (
                      <div className="w-full h-40 bg-white/5 rounded-lg animate-pulse flex items-center justify-center">
                        <svg className="w-8 h-8 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                      </div>
                    )}
                    
                    <img
                      src={fullImageUrl}
                      alt={msg.texto || 'Imagen compartida'}
                      className={`rounded-lg max-w-full max-h-[300px] object-contain cursor-pointer transition-all hover:opacity-90 ${isImageLoaded ? 'block' : 'hidden'}`}
                      onLoad={() => setIsImageLoaded(true)}
                      onError={() => setImageError(true)}
                      onClick={() => setShowFullImage(true)}
                    />
                  </div>
                )}

                {isImage && imageError && (
                  <a
                    href={fullImageUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-2 px-3 py-2 bg-white/5 rounded-lg text-sm text-indigo-300 hover:bg-white/10 transition-colors"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    Ver imagen ({msg.texto || 'imagen'})
                  </a>
                )}

                {(isPDF || (!isImage && fileUrl)) && (
                  <div className="flex items-center gap-2 mt-1">
                    <a
                      href={fullImageUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-2 px-3 py-2 bg-white/5 rounded-lg text-sm text-indigo-300 hover:bg-white/10 transition-colors max-w-full truncate"
                    >
                      <svg className="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                      </svg>
                      <span className="truncate">{msg.texto || 'Descargar archivo'}</span>
                    </a>
                    
                    {isOwn && onDeleteFile && (
                      <button 
                        onClick={() => onDeleteFile(fileUrl)}
                        title="Eliminar archivo"
                        className="shrink-0 w-8 h-8 flex items-center justify-center rounded-full text-slate-400 hover:bg-red-500/20 hover:text-red-400 transition-colors"
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <polyline points="3 6 5 6 21 6"></polyline>
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                        </svg>
                      </button>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>

          <div className={`text-[0.6rem] text-slate-500 px-3 pb-1.5 ${isOwn ? 'text-right' : 'text-left'}`}>
            {formatTime(msg.timestamp)}
          </div>
        </div>
      </div>

      {showFullImage && (
        <div
          className="fixed inset-0 z-[100] bg-black/90 flex items-center justify-center cursor-pointer backdrop-blur-md"
          onClick={() => setShowFullImage(false)}
        >
          <img
            src={fullImageUrl}
            alt="Vista completa"
            className="max-w-[90vw] max-h-[90vh] object-contain"
          />
          <button
            className="absolute top-4 right-4 w-10 h-10 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center transition-colors"
            onClick={() => setShowFullImage(false)}
          >
            <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}
    </div>
  );
}