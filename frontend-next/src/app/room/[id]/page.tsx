'use client';

import { useState, useEffect, useRef, use } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/context/UserContext';
import { ChatMessage as MsgType, UploadTask } from '@/types';
import { ChatMessage } from '@/components/room/ChatMessage';
import { ChatSidebar } from '@/components/room/ChatSidebar';
import { UploadToast } from '@/components/room/ProgressModal';

const API = process.env.NEXT_PUBLIC_API_URL;

export default function RoomPage({ params }: { params: Promise<{ id: string }> }) {
  const unwrappedParams = use(params);
  const roomId = unwrappedParams.id;
  const router = useRouter();
  const { userContext } = useUser();

  const [roomName, setRoomName] = useState(() => typeof window !== 'undefined' ? sessionStorage.getItem('room_name') || roomId : roomId);
  const [roomType, setRoomType] = useState(() => typeof window !== 'undefined' ? sessionStorage.getItem('room_type') || 'texto' : 'texto');
  const [roomToken, setRoomToken] = useState(() => typeof window !== 'undefined' ? sessionStorage.getItem('room_token') || '' : '');

  const [ws, setWs] = useState<WebSocket | null>(null);
  const [messages, setMessages] = useState<MsgType[]>([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [users, setUsers] = useState<string[]>([]);
  const [wsStatus, setWsStatus] = useState('connecting');
  const [uploads, setUploads] = useState<UploadTask[]>([]);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const chatBodyRef = useRef<HTMLDivElement>(null);

  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  useEffect(() => {
    if (!userContext?.nickname) {
      router.push('/'); 
      return;
    }

    // Cargar historial de mensajes
    fetch(`${API}/rooms/${roomId}/messages`)
      .then(res => res.json())
      .then(data => {
        if (Array.isArray(data)) {
          setMessages(data);
          setTimeout(() => { 
            if (chatBodyRef.current) chatBodyRef.current.scrollTop = chatBodyRef.current.scrollHeight; 
          }, 100);
        }
      })
      .catch(err => console.error('Error fetching message history:', err));

    setWsStatus('connecting');
    const wsProtocol = API.startsWith('https') ? 'wss' : 'ws';
    const wsHost = API.replace(/^https?:\/\//, '');
    const url = `${wsProtocol}://${wsHost}/ws/${roomId}?nickname=${encodeURIComponent(userContext.nickname)}&sala_id=${roomId}&device_id=${userContext.deviceId}`;
    
    const websocket = new WebSocket(url);
    websocket.onopen = () => setWsStatus('open');
    websocket.onclose = () => setWsStatus('closed');
    websocket.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        
        if (data.type === 'user_list' || data.tipo === 'user_list') {
          try {
            const parsed = typeof data.texto === 'string' ? JSON.parse(data.texto) : data;
            setUsers(parsed.users || data.users || []);
          } catch { 
            setUsers(data.users || []); 
          }
        } 
        else if (data.tipo === 'join' || data.type === 'join') {
          setMessages(prev => [...prev, {
            type: 'system',
            texto: data.texto || `${data.nickname} se unió a la sala`,
            nickname: 'Sistema',
            timestamp: data.timestamp || Date.now() / 1000
          }]);
          setTimeout(() => { 
            if (chatBodyRef.current) chatBodyRef.current.scrollTop = chatBodyRef.current.scrollHeight; 
          }, 50);
        }
        else if (data.tipo === 'leave' || data.type === 'leave') {
          setMessages(prev => [...prev, {
            type: 'system',
            texto: data.texto || `${data.nickname} salió de la sala`,
            nickname: 'Sistema',
            timestamp: data.timestamp || Date.now() / 1000
          }]);
          setTimeout(() => { 
            if (chatBodyRef.current) chatBodyRef.current.scrollTop = chatBodyRef.current.scrollHeight; 
          }, 50);
        }
        else {
          setMessages(prev => [...prev, data]);
          setTimeout(() => { 
            if (chatBodyRef.current) chatBodyRef.current.scrollTop = chatBodyRef.current.scrollHeight; 
          }, 50);
        }
      } catch (err) {
        console.error('Error parsing message:', err);
      }
    };

    setWs(websocket);

    return () => { 
      websocket.close(); 
    };
  }, [roomId, userContext, router]);

  const sendMessage = () => {
    if (!currentMessage.trim() || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ tipo: 'chat', texto: currentMessage.trim() }));
    setCurrentMessage('');
  };

  const uploadFile = (file: File) => {
  return new Promise((resolve, reject) => {
    const id = Math.random().toString(36).slice(2);
    const MAX_SIZE = 10 * 1024 * 1024; // 10MB
    
    if (file.size > MAX_SIZE) {
      setUploads(prev => [...prev, { 
        id, 
        name: file.name, 
        progress: 0, 
        done: true, 
        error: `Archivo demasiado grande (máx. 10MB)` 
      }]);
      resolve(false);
      return;
    }
    setUploads(prev => [...prev, { id, name: file.name, progress: 0, done: false, error: null }]);

    const formData = new FormData();
    formData.append('sala_id', roomId);
    formData.append('file', file);

    const xhr = new XMLHttpRequest();
    
    xhr.upload.addEventListener('progress', (ev) => {
      if (ev.lengthComputable) {
        const percentComplete = Math.round((ev.loaded / ev.total) * 100);
        console.log(`Progreso ${file.name}: ${percentComplete}%`);
        setUploads(prev => prev.map(u => 
          u.id === id ? { ...u, progress: percentComplete } : u
        ));
      }
    });

    xhr.addEventListener('load', () => {
      console.log(`Respuesta ${file.name}: Status ${xhr.status}`);
      
      if (xhr.status >= 200 && xhr.status < 300) {
        setUploads(prev => prev.map(u => 
          u.id === id ? { ...u, progress: 100, done: true, error: null } : u
        ));
        
        setTimeout(() => {
          setUploads(prev => prev.filter(u => u.id !== id));
        }, 2000);
        
        resolve(true);
      } else {
        let errorMsg = 'Error al subir';
        try {
          const response = JSON.parse(xhr.responseText);
          errorMsg = response.error || response.message || errorMsg;
        } catch {
          errorMsg = `Error ${xhr.status}: ${xhr.statusText || 'Error desconocido'}`;
        }
        
        setUploads(prev => prev.map(u => 
          u.id === id ? { ...u, done: true, error: errorMsg } : u
        ));
        
        setTimeout(() => {
          setUploads(prev => prev.filter(u => u.id !== id));
        }, 3000);
        
        resolve(false);
      }
    });

    xhr.addEventListener('error', () => {
      console.error(`Error de red al subir ${file.name}`);
      setUploads(prev => prev.map(u => 
        u.id === id ? { ...u, done: true, error: 'Error de conexión al servidor' } : u
      ));
      
      setTimeout(() => {
        setUploads(prev => prev.filter(u => u.id !== id));
      }, 3000);
      
      resolve(false);
    });


    xhr.addEventListener('abort', () => {
      console.log(`Upload abortado: ${file.name}`);
      setUploads(prev => prev.filter(u => u.id !== id));
      resolve(false);
    });

    xhr.open('POST', `${API}/upload/file`);
    xhr.setRequestHeader('Authorization', `Bearer ${roomToken}`);
    xhr.send(formData);
  });
};

const handleFilesSelected = async (e: React.ChangeEvent<HTMLInputElement>) => {
  if (!e.target.files?.length) return;
  const filesArray = Array.from(e.target.files);
  e.target.value = '';
  
  await Promise.all(filesArray.map(file => uploadFile(file)));
};

  const leaveRoom = async () => {
    try {
      await fetch(`${API}/rooms/leave`, {
        method: 'POST', 
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sala_id: roomId, usuario_id: userContext.userId })
      });
    } catch (e) {
      console.error('Error leaving room:', e);
    } finally {
      if (ws) ws.close();
      router.push('/home');
    }
  };

  const formatTime = (ts?: number) => {
    const d = ts ? new Date(ts * 1000) : new Date();
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (!userContext?.nickname) return null;

  return (
    <div className="flex h-screen overflow-hidden relative">
      {/* Toast de uploads */}
      <UploadToast 
        uploads={uploads} 
          onRemove={(id) => setUploads(prev => prev.filter(u => u.id !== id))}
      />

      {/* Sidebar overlay */}
      {isSidebarOpen && (
        <div 
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 md:hidden"
          onClick={() => setIsSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed inset-y-0 left-0 z-50 w-[280px] bg-slate-950 transition-transform duration-300 ease-in-out transform
        md:relative md:translate-x-0 md:z-auto
        ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
      `}>
        <ChatSidebar 
          roomName={roomName} 
          roomType={roomType} 
          wsStatus={wsStatus} 
          users={users} 
          myNickname={userContext.nickname} 
          leaveRoom={leaveRoom} 
        />
      </aside>

      {/* Main chat area */}
      <main className="flex-1 flex flex-col min-w-0 bg-slate-900/30 overflow-hidden">
        
        {/* Header */}
        <div className="flex justify-between items-center px-6 py-4 bg-slate-950/65 backdrop-blur-md border-b border-white/15 shrink-0">
          <div className="flex items-center gap-3">
            <button 
              onClick={() => setIsSidebarOpen(true)}
              className="md:hidden p-2 -ml-2 text-slate-400 hover:text-white transition-colors"
            >
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <line x1="3" y1="12" x2="21" y2="12"/>
                <line x1="3" y1="6" x2="21" y2="6"/>
                <line x1="3" y1="18" x2="21" y2="18"/>
              </svg>
            </button>

            <div className="flex items-center gap-2.5 text-[0.9rem]">
              <div className={`w-2 h-2 rounded-full shadow-[0_0_6px_currentColor] ${roomType === 'multimedia' ? 'bg-pink-400 text-pink-400' : 'bg-indigo-300 text-indigo-300'}`}></div>
              <span className="truncate max-w-[150px] md:max-w-none"><strong>{roomName}</strong></span>
            </div>
          </div>

          <div className="text-[0.75rem] text-slate-500 hidden sm:block">
            {messages.length} mensajes
          </div>
        </div>

        {/* Messages area */}
        <div className="flex-1 overflow-y-auto p-4 md:p-6 flex flex-col gap-3" ref={chatBodyRef}>
          {messages.length === 0 && (
            <div className="flex-1 flex flex-col items-center justify-center gap-3 text-slate-500 opacity-70">
              <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <p className="text-[0.85rem]">Sé el primero en enviar un mensaje</p>
            </div>
          )}
          {messages.map((msg, idx) => (
            <ChatMessage 
              key={idx} 
              msg={msg} 
              isOwn={msg.nickname === userContext.nickname} 
              formatTime={formatTime} 
              api={API} 
            />
          ))}
        </div>

        {/* Input area */}
        <div className="px-4 py-3 md:px-5 md:py-3.5 bg-slate-950/75 backdrop-blur-[20px] border-t border-white/15 flex items-center gap-2.5 shrink-0">
          {roomType === 'multimedia' && (
            <label className="w-10 h-10 shrink-0 rounded-xl bg-white/5 border border-white/15 flex items-center justify-center cursor-pointer text-slate-500 hover:bg-white/10 transition-all hover:text-indigo-400">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/>
              </svg>
              <input type="file" ref={fileInputRef} onChange={handleFilesSelected} multiple hidden />
            </label>
          )}
          
          <div className="flex-1 flex items-center gap-2 bg-black/45 border border-white/20 rounded-2xl p-1.5 pl-4 focus-within:border-indigo-300/50 transition-all">
            <textarea
              className="flex-1 bg-transparent border-none resize-none text-[0.9rem] py-1 max-h-[120px] outline-none text-slate-100 placeholder:text-slate-500"
              value={currentMessage}
              onChange={e => setCurrentMessage(e.target.value)}
              onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); } }}
              placeholder="Escribe un mensaje…"
              rows={1}
            />
            <button
              className={`w-9 h-9 rounded-xl flex items-center justify-center transition-all ${currentMessage.trim() ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/30' : 'text-slate-500 opacity-50'}`}
              onClick={sendMessage}
              disabled={!currentMessage.trim() || wsStatus !== 'open'}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <line x1="22" y1="2" x2="11" y2="13"/>
                <polygon points="22 2 15 22 11 13 2 9 22 2"/>
              </svg>
            </button>
          </div>
        </div>
      </main>
    </div>
  );
}