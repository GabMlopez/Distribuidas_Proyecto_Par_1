interface Props {
  show: boolean;
  onClose: () => void;
  selectedRoomName: string;
  joinPin: string;
  setJoinPin: (v: string) => void;
  joinRoom: () => void;
  joining: boolean;
  joinError: string;
}

export function JoinRoomModal({ show, onClose, selectedRoomName, joinPin, setJoinPin, joinRoom, joining, joinError }: Props) {
  if (!show) return null;

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/75 backdrop-blur-md flex items-center justify-center animate-fade-in" onClick={(e) => { if(e.target===e.currentTarget) onClose(); }}>
      <div className="w-full max-w-[420px] bg-slate-900/90 backdrop-blur-[28px] border border-white/20 rounded-2xl shadow-[0_24px_80px_rgba(0,0,0,0.7),inset_0_1px_0_rgba(255,255,255,0.14),0_0_0_1px_rgba(99,102,241,0.12)] p-8 flex flex-col gap-5 animate-fade-up">
        
        <div className="flex justify-between items-center pb-3 border-b border-white/10">
          <div>
            <h3 className="text-lg font-bold">Unirse a la sala</h3>
            <code className="text-xs bg-black/50 px-2 py-1 rounded text-slate-400 font-mono mt-1 block">{selectedRoomName}</code>
          </div>
          <button className="w-8 h-8 flex items-center justify-center rounded-lg bg-white/5 hover:bg-white/10 text-slate-400 hover:text-white transition-all" onClick={onClose}>✕</button>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-semibold text-slate-400">PIN de acceso</label>
          <input
            type="password"
            value={joinPin}
            onChange={e => setJoinPin(e.target.value)}
            placeholder="••••••"
            maxLength={6}
            onKeyDown={(e) => e.key === 'Enter' && joinRoom()}
            autoFocus
            className="w-full p-3 bg-black/40 border border-white/20 rounded-xl text-slate-100 placeholder:text-slate-500 focus:border-indigo-400 focus:bg-indigo-500/10 focus:ring-4 focus:ring-indigo-500/20 transition-all outline-none box-border"
          />
        </div>
        
        {joinError && <div className="flex items-center gap-2 p-3 bg-rose-500/10 border border-rose-500/20 rounded-xl text-sm text-rose-300">{joinError}</div>}
        
        <div className="flex justify-end gap-3 mt-2">
          <button className="px-5 py-2.5 rounded-xl font-semibold bg-white/5 hover:bg-white/10 text-slate-300 border border-white/10 transition-all" onClick={onClose}>Cancelar</button>
          <button className="px-5 py-2.5 rounded-xl font-semibold bg-gradient-to-br from-indigo-500 to-purple-600 text-white shadow-[0_4px_20px_rgba(99,102,241,0.4)] hover:shadow-[0_6px_28px_rgba(99,102,241,0.6)] disabled:opacity-50 disabled:shadow-none transition-all active:scale-95" onClick={joinRoom} disabled={joining || joinPin.length < 4}>
            {joining ? 'Conectando...' : 'Entrar'}
          </button>
        </div>

      </div>
    </div>
  );
}
