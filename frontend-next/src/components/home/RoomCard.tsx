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
  return (
    <div className="relative group animate-fade-up" style={{ animationDelay: `${index * 60}ms` }}>
      <button
        className={`w-full aspect-[4/3] rounded-3xl bg-white/5 border border-white/15 backdrop-blur-[20px] shadow-[inset_0_1px_0_rgba(255,255,255,0.1),0_4px_28px_rgba(0,0,0,0.55)] flex items-center justify-center transition-all duration-300 ease-out hover:-translate-y-1 hover:shadow-[0_20px_60px_rgba(0,0,0,0.65),0_0_24px_rgba(99,102,241,0.2),0_0_0_1px_rgba(99,102,241,0.25)] ${room.tipo === 'multimedia' ? 'text-pink-400 hover:border-pink-400/40' : 'text-indigo-400 hover:border-indigo-400/40'}`}
        onClick={() => openJoin(room)}
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

      <div className="absolute top-1/2 left-1/2 w-[220px] bg-slate-900/95 backdrop-blur-[28px] border border-white/20 rounded-2xl p-4 shadow-[0_24px_80px_rgba(0,0,0,0.8),inset_0_1px_0_rgba(255,255,255,0.1),0_0_0_1px_rgba(99,102,241,0.15)] opacity-0 invisible translate-x-[-50%] translate-y-[-40%] scale-95 transition-all duration-300 ease-out z-20 pointer-events-none group-hover:opacity-100 group-hover:visible group-hover:translate-x-[-50%] group-hover:translate-y-[-60%] group-hover:scale-100 group-hover:pointer-events-auto">
        <div className="flex justify-between items-start mb-3">
          <h4 className="text-[0.95rem] font-bold tracking-tight text-white m-0 max-w-[65%] truncate">{room.nombre || room.sala_id}</h4>
          <span className={`inline-flex items-center px-[0.65rem] py-[0.2rem] rounded-full text-[0.65rem] font-bold uppercase tracking-[0.05em] ${room.tipo === 'texto' ? 'bg-indigo-500/15 text-indigo-200 border border-indigo-400/40 shadow-[0_0_8px_rgba(165,180,252,0.5)]' : 'bg-pink-500/15 text-pink-200 border border-pink-400/40 shadow-[0_0_8px_rgba(249,168,212,0.5)]'}`}>
            {room.tipo}
          </span>
        </div>

        <div className="mb-4 pt-1">
          <div className="flex items-center gap-[0.4rem] text-[0.75rem] font-medium text-slate-300 bg-white/5 py-1.5 px-3 rounded-lg border border-white/10 w-fit">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
            {room.usuarios ?? 0} conectado{(room.usuarios ?? 0) !== 1 ? 's' : ''}
          </div>
        </div>

        <div className="flex items-center justify-between mt-auto">
          <button className="flex items-center gap-1.5 px-4 py-1.5 bg-gradient-to-br from-indigo-500 to-purple-600 text-white text-xs font-semibold rounded-lg shadow-[0_4px_20px_rgba(99,102,241,0.4)] hover:shadow-[0_6px_28px_rgba(99,102,241,0.6)] hover:-translate-y-px transition-all active:scale-95" onClick={() => openJoin(room)}>
            Unirse
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </button>
          {isAdmin && (
            <div className="flex items-center gap-1.5">
              <button className="w-7 h-7 rounded-md bg-white/5 border border-white/15 text-slate-400 flex items-center justify-center hover:bg-white/10 hover:text-white transition-all" onClick={() => openEdit(room)} title="Editar">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </button>
              <button className="w-7 h-7 rounded-md bg-rose-500/10 border border-rose-500/30 text-rose-400 flex items-center justify-center hover:bg-rose-500/20 hover:border-rose-500/50 hover:shadow-[0_0_15px_rgba(251,113,133,0.2)] transition-all" onClick={() => deleteRoom(room.sala_id)} title="Eliminar">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
