import { ChatMessage as MsgType } from '@/types';

interface Props {
  msg: MsgType;
  isOwn: boolean;
  formatTime: (ts?: number) => string;
  api: string;
}

export function ChatMessage({ msg, isOwn, formatTime, api }: Props) {
  if (msg.tipo === 'join' || msg.tipo === 'leave') {
    return (
      <div className="flex items-center gap-3 my-1">
        <div className="flex-1 h-px bg-white/5"></div>
        <span className="text-xs text-slate-500 whitespace-nowrap">{msg.texto}</span>
        <div className="flex-1 h-px bg-white/5"></div>
      </div>
    );
  }

  if (msg.tipo === 'multimedia') {
    return (
      <div className={`flex items-end gap-2 max-w-[75%] ${isOwn ? 'self-end flex-row-reverse' : 'self-start'}`}>
        {!isOwn && <div className="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center text-[0.65rem] font-bold shrink-0 mb-0.5">{msg.nickname?.charAt(0)?.toUpperCase()}</div>}
        <div className="p-3 rounded-2xl flex flex-col gap-1 max-w-full break-words bg-pink-500/10 border border-pink-500/20">
          {!isOwn && <span className="text-[0.7rem] font-bold text-indigo-300 mb-[0.1rem]">{msg.nickname}</span>}
          <p className="text-sm leading-snug text-slate-100 m-0">{msg.texto}</p>
          {msg.file_url && (
            <a href={api + msg.file_url} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-pink-500/15 border border-pink-500/25 rounded-lg text-pink-400 text-xs no-underline transition-colors mt-1 hover:bg-pink-500/25">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="mr-0.5">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
              Descargar archivo
            </a>
          )}
          <span className="text-[0.6rem] text-white/35 self-end mt-1">{formatTime(msg.timestamp)}</span>
        </div>
      </div>
    );
  }

  return (
    <div className={`flex items-end gap-2 max-w-[75%] ${isOwn ? 'self-end flex-row-reverse' : 'self-start'}`}>
      {!isOwn && <div className="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center text-[0.65rem] font-bold shrink-0 mb-0.5">{msg.nickname?.charAt(0)?.toUpperCase()}</div>}
      <div className={`p-3 rounded-2xl flex flex-col gap-1 max-w-full break-words ${isOwn ? 'bg-gradient-to-br from-indigo-500 to-purple-600 rounded-br-sm shadow-[0_4px_16px_rgba(99,102,241,0.3)]' : 'bg-white/5 border border-white/20 rounded-bl-sm'}`}>
        {!isOwn && <span className="text-[0.7rem] font-bold text-indigo-300 mb-[0.1rem]">{msg.nickname}</span>}
        <p className="text-sm leading-snug text-slate-100 m-0">{msg.texto}</p>
        <span className="text-[0.6rem] text-white/35 self-end mt-1">{formatTime(msg.timestamp)}</span>
      </div>
    </div>
  );
}
