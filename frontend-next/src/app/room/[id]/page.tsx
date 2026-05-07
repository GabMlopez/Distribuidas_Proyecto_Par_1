'use client';

import { useState, useEffect, useRef, use } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/context/UserContext';
import { ChatMessage as MsgType, UploadTask } from '@/types';
import { ChatMessage } from '@/components/room/ChatMessage';
import { ChatSidebar } from '@/components/room/ChatSidebar';

const API = 'http://localhost:8080';

export default function RoomPage({ params }: { params: Promise<{ id: string }> }) {
  const unwrappedParams = use(params);
  const roomId = unwrappedParams.id;
  const router = useRouter();
  const { userContext } = useUser();

  const [roomName, setRoomName] = useState(roomId);
  const [roomType, setRoomType] = useState('texto');
  const [roomToken, setRoomToken] = useState('');

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
      router.push('/'); return;
    }
    setRoomName(sessionStorage.getItem('room_name') || roomId);
    setRoomType(sessionStorage.getItem('room_type') || 'texto');
    setRoomToken(sessionStorage.getItem('room_token') || '');

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
          const parsed = typeof data.texto === 'string' ? JSON.parse(data.texto) : data;
          setUsers(parsed.users || data.users || []);
        } else {
          setMessages(prev => [...prev, data]);
          setTimeout(() => { if (chatBodyRef.current) chatBodyRef.current.scrollTop = chatBodyRef.current.scrollHeight; }, 50);
        }
      } catch (err) {}
    };

    setWs(websocket);

    const handleBeforeUnload = () => {
      const payload = JSON.stringify({ sala_id: roomId, nickname: userContext.nickname, device_id: userContext.deviceId });
      navigator.sendBeacon(`${API}/rooms/leave`, new Blob([payload], { type: 'application/json' }));
      websocket.close();
    };
    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => { window.removeEventListener('beforeunload', handleBeforeUnload); websocket.close(); };
  }, [roomId, userContext, router]);

  const sendMessage = () => {
    if (!currentMessage.trim() || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ tipo: 'chat', texto: currentMessage.trim() }));
    setCurrentMessage('');
  };

  const uploadSingleFile = (file: File) => {
    const id = Math.random().toString(36).slice(2);
    setUploads(prev => [...prev, { id, name: file.name, progress: 0, done: false, error: null }]);

    return new Promise((resolve) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('sala_id', roomId); formData.append('file', file);

      xhr.upload.onprogress = (ev) => {
        if (ev.lengthComputable) {
          const pct = Math.round((ev.loaded / ev.total) * 100);
          setUploads(prev => prev.map(u => u.id === id ? { ...u, progress: pct } : u));
        }
      };

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          setUploads(prev => prev.map(u => u.id === id ? { ...u, progress: 100, done: true } : u));
          setTimeout(() => setUploads(prev => prev.filter(u => u.id !== id)), 2500);
          resolve({ ok: true });
        } else {
          let errorMsg = 'Error al subir';
          try { errorMsg = JSON.parse(xhr.responseText).error || errorMsg; } catch {}
          setUploads(prev => prev.map(u => u.id === id ? { ...u, error: errorMsg } : u));
          setTimeout(() => setUploads(prev => prev.filter(u => u.id !== id)), 5000);
          resolve({ ok: false });
        }
      };
      xhr.onerror = () => {
        setUploads(prev => prev.map(u => u.id === id ? { ...u, error: 'Sin conexión' } : u));
        setTimeout(() => setUploads(prev => prev.filter(u => u.id !== id)), 4000);
        resolve({ ok: false });
      };
      xhr.open('POST', `${API}/upload/file`);
      xhr.setRequestHeader('Authorization', `Bearer ${roomToken}`);
      xhr.send(formData);
    });
  };

  const handleFilesSelected = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files?.length) return;
    const filesArray = Array.from(e.target.files);
    e.target.value = '';
    await Promise.allSettled(filesArray.map(uploadSingleFile));
  };

  const leaveRoom = async () => {
    try {
      await fetch(`${API}/rooms/leave`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sala_id: roomId, usuario_id: userContext.userId})
      });
    } catch {}
    if (ws) ws.close();
    router.push('/home');
  };

  const formatTime = (ts?: number) => {
    const d = ts ? new Date(ts * 1000) : new Date();
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (!userContext?.nickname) return null;

  return (
  <div className="flex h-screen overflow-hidden relative">
    
    {isSidebarOpen && (
      <div 
        className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 md:hidden"
        onClick={() => setIsSidebarOpen(false)}
      />
    )}

    <aside className={`
      fixed inset-y-0 left-0 z-50 w-[280px] bg-slate-950 transition-transform duration-300 ease-in-out transform
      md:relative md:translate-x-0 md:z-auto
      ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'}
    `}>
      <ChatSidebar 
        roomName={roomName} roomType={roomType} wsStatus={wsStatus} 
        users={users} uploads={uploads} myNickname={userContext.nickname} leaveRoom={leaveRoom} 
      />
    </aside>

    <main className="flex-1 flex flex-col min-w-0 bg-slate-900/30 overflow-hidden">
      
      <div className="flex justify-between items-center px-6 py-4 bg-slate-950/65 backdrop-blur-md border-b border-white/15 shrink-0">
        <div className="flex items-center gap-3">
        
          <button 
            onClick={() => setIsSidebarOpen(true)}
            className="md:hidden p-2 -ml-2 text-slate-400 hover:text-white transition-colors"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/>
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

      
      <div className="flex-1 overflow-y-auto p-4 md:p-6 flex flex-col gap-3" ref={chatBodyRef}>
        {messages.length === 0 && (
          <div className="flex-1 flex flex-col items-center justify-center gap-3 text-slate-500 opacity-70">
            
             <p className="text-[0.85rem]">Sé el primero en enviar un mensaje</p>
          </div>
        )}
        {messages.map((msg, idx) => (
          <ChatMessage key={idx} msg={msg} isOwn={msg.nickname === userContext.nickname} formatTime={formatTime} api={API} />
        ))}
      </div>

     
      <div className="px-4 py-3 md:px-5 md:py-3.5 bg-slate-950/75 backdrop-blur-[20px] border-t border-white/15 flex items-center gap-2.5 shrink-0">
        {roomType === 'multimedia' && (
          <label className="w-10 h-10 shrink-0 rounded-xl bg-white/5 border border-white/15 flex items-center justify-center cursor-pointer text-slate-500 hover:bg-white/10 transition-all">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/></svg>
            <input type="file" ref={fileInputRef} onChange={handleFilesSelected} multiple hidden />
          </label>
        )}
        
        <div className="flex-1 flex items-center gap-2 bg-black/45 border border-white/20 rounded-2xl p-1.5 pl-4 focus-within:border-indigo-300/50 transition-all">
          <textarea
            className="flex-1 bg-transparent border-none resize-none text-[0.9rem] py-1 max-h-[120px] outline-none text-slate-100"
            value={currentMessage}
            onChange={e => setCurrentMessage(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); } }}
            placeholder="Escribe un mensaje…"
            rows={1}
          />
          <button
            className={`w-9 h-9 rounded-xl flex items-center justify-center transition-all ${currentMessage.trim() ? 'bg-indigo-600 text-white' : 'text-slate-500 opacity-50'}`}
            onClick={sendMessage}
            disabled={!currentMessage.trim() || wsStatus !== 'open'}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
          </button>
        </div>
      </div>
    </main>
  </div>
);
}
