import { useState } from 'react'; 
import { Room } from '@/types';

interface Props {
  room: Room;
  index: number;
  openJoin: (room: Room) => void;
  openEdit: (room: Room) => void;
  deleteRoom: (id: string) => void;
  isAdmin: boolean;
}

export function RoomCard({ room, index, openJoin, openEdit, deleteRoom, isAdmin }: Props) {
  const [showTooltip, setShowTooltip] = useState(false);

  return (
    <div 
      className="relative group animate-fade-up hover:z-50" 
      style={{ animationDelay: `${index * 60}ms` }}
      // En PC se abre con hover, en móvil/clic alternamos con el estado
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <button
        className={`w-full aspect-[4/3] rounded-3xl bg-white/5 border border-white/15 backdrop-blur-[20px] shadow-[inset_0_1px_0_rgba(255,255,255,0.1),0_4px_28px_rgba(0,0,0,0.55)] flex items-center justify-center transition-all duration-300 ease-out hover:-translate-y-1 hover:shadow-[0_20px_60px_rgba(0,0,0,0.65),0_0_24px_rgba(99,102,241,0.2),0_0_0_1px_rgba(99,102,241,0.25)] ${
          room.tipo === 'multimedia' ? 'text-pink-400 hover:border-pink-400/40' : 'text-indigo-400 hover:border-indigo-400/40'
        }`}
        onClick={() => setShowTooltip(!showTooltip)} // Al hacer clic/tocar, mostramos el tooltip
      >
        {room.tipo === 'multimedia' ? (
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="2" y1="7" x2="7" y2="7"/><line x1="2" y1="17" x2="7" y2="17"/><line x1="17" y1="17" x2="22" y2="17"/><line x1="17" y1="7" x2="22" y2="7"/></svg>
        ) : (
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        )}
      </button>

      <div className="mt-3 text-center text-sm font-semibold text-slate-400 truncate px-2 group-hover:text-indigo-300 transition-colors">
        {room.nombre || room.sala_id}
      </div>

      <div 
        className={`absolute top-1/2 left-1/2 w-[220px] bg-slate-900/95 backdrop-blur-[28px] border border-white/20 rounded-2xl p-4 shadow-2xl transition-all duration-300 ease-out z-20 translate-x-[-50%] scale-95 
        ${showTooltip 
          ? 'opacity-100 visible -translate-y-[60%] scale-100 pointer-events-auto' 
          : 'opacity-0 invisible -translate-y-[40%] pointer-events-none'
        }`}
      >
        <div className="flex justify-between items-start mb-3">
          <h4 className="text-[0.95rem] font-bold tracking-tight text-white m-0 max-w-[65%] truncate">{room.nombre || room.sala_id}</h4>
          <span className={`inline-flex items-center px-[0.65rem] py-[0.2rem] rounded-full text-[0.65rem] font-bold uppercase tracking-[0.05em] ${room.tipo === 'texto' ? 'bg-indigo-500/15 text-indigo-200 border border-indigo-400/40' : 'bg-pink-500/15 text-pink-200 border border-pink-400/40'}`}>
            {room.tipo}
          </span>
        </div>

        {isAdmin && (
          <div className="mb-2">
            <p className="text-[0.7rem] text-slate-400">
              <span className="font-semibold text-indigo-300">PIN:</span> {room.pin}
            </p>
          </div>
        )}

        <div className="mb-4 pt-1">
          <div className="flex items-center gap-[0.4rem] text-[0.75rem] font-medium text-slate-300 bg-white/5 py-1.5 px-3 rounded-lg border border-white/10 w-fit">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/></svg>
            {room.usuarios ?? 0} conectados
          </div>
        </div>

        <div className="flex items-center justify-between mt-auto">
          <button 
            className="flex items-center gap-1.5 px-4 py-1.5 bg-gradient-to-br from-indigo-500 to-purple-600 text-white text-xs font-semibold rounded-lg shadow-lg active:scale-95" 
            onClick={(e) => {
              e.stopPropagation();
              openJoin(room);
            }}
          >
            Unirse
          </button>
          
          {isAdmin && (
            <div className="flex items-center gap-1.5">
              <button 
                className="w-7 h-7 rounded-md bg-white/5 border border-white/15 text-slate-400 flex items-center justify-center hover:text-white" 
                onClick={(e) => { e.stopPropagation(); openEdit(room); }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button 
                className="w-7 h-7 rounded-md bg-rose-500/10 border border-rose-500/30 text-rose-400 flex items-center justify-center hover:bg-rose-500/20" 
                onClick={(e) => { e.stopPropagation(); deleteRoom(room.sala_id); }}
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}