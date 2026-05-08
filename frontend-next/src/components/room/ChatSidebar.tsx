'use client';

interface Props {
  roomName: string;
  roomType: string;
  wsStatus: string;
  users: string[];
  myNickname: string;
  leaveRoom: () => void;
}

export function ChatSidebar({ roomName, roomType, wsStatus, users, myNickname, leaveRoom }: Props) {
  return (
    <aside className="bg-slate-950/75 backdrop-blur-[24px] border-r border-white/20 flex flex-col overflow-hidden w-[270px] shrink-0">
      {/* Room info */}
      <div className="p-5">
        <div className="flex items-center gap-3 mb-3">
          <div className="w-[38px] h-[38px] rounded-xl bg-indigo-500/20 border border-indigo-500/25 flex items-center justify-center text-indigo-300 shrink-0">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <div>
            <div className="font-bold text-[0.95rem] tracking-tight mb-1">{roomName}</div>
            <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-[0.7rem] font-bold tracking-[0.07em] uppercase ${roomType === 'texto' ? 'bg-indigo-500/15 text-indigo-200 border border-indigo-400/40 shadow-[0_0_8px_rgba(165,180,252,0.5)]' : 'bg-pink-500/15 text-pink-200 border border-pink-400/40 shadow-[0_0_8px_rgba(249,168,212,0.5)]'}`}>
              {roomType === 'texto' ? 'Texto' : 'Multimedia'}
            </span>
          </div>
        </div>
        
        {/* Status */}
        <div className={`inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full bg-white/5 border border-white/10 w-fit ${wsStatus === 'open' ? 'text-emerald-400 border-emerald-500/20 bg-emerald-500/10' : wsStatus === 'connecting' ? 'text-amber-400 border-amber-500/20' : 'text-rose-400 border-rose-500/20'}`}>
          <div className={`w-1.5 h-1.5 rounded-full bg-current shadow-[0_0_4px_currentColor] ${wsStatus === 'open' ? 'animate-pulse' : ''}`}></div>
          {wsStatus === 'open' ? 'Conectado' : wsStatus === 'connecting' ? 'Conectando…' : 'Desconectado'}
        </div>
      </div>

      <div className="h-px bg-white/10 m-0 shrink-0"></div>

      {/* Users list */}
      <div className="flex-1 overflow-hidden flex flex-col p-5 min-h-0">
        <div className="text-[0.7rem] font-bold uppercase tracking-[0.08em] text-slate-500 mb-3">
          Participantes · {users.length}
        </div>
        <div className="flex flex-col gap-1 overflow-y-auto flex-1">
          {users.map(user => (
            <div key={user} className={`flex items-center gap-2.5 p-2 rounded-lg transition-colors ${user === myNickname ? 'bg-indigo-500/10' : 'hover:bg-white/5'}`}>
              <div className={`w-7 h-7 rounded-full bg-white/10 flex items-center justify-center text-[0.72rem] font-bold shrink-0 ${user === myNickname ? 'bg-gradient-to-br from-indigo-500 to-pink-500' : ''}`}>
                {user.charAt(0).toUpperCase()}
              </div>
              <span className={`text-[0.83rem] flex-1 min-w-0 overflow-hidden text-ellipsis whitespace-nowrap ${user === myNickname ? 'text-slate-100' : 'text-slate-400'}`}>
                {user}{user === myNickname ? ' (tú)' : ''}
              </span>
              {user === myNickname && (
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 shadow-[0_0_5px_rgba(52,211,153,1)] shrink-0"></div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Leave button */}
      <div className="p-4 border-t border-white/10 shrink-0">
        <button 
          className="w-full flex items-center justify-center gap-2 px-4 py-3 text-[0.82rem] font-semibold rounded-xl bg-rose-500/10 text-rose-400 border border-rose-500/30 hover:bg-rose-500/20 hover:border-rose-500/50 hover:shadow-[0_0_20px_rgba(251,113,133,0.2)] transition-all active:scale-95" 
          onClick={leaveRoom}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
          Salir de la sala
        </button>
      </div>
    </aside>
  );
}