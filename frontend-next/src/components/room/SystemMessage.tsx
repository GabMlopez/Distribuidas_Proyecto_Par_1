'use client';

interface SystemMessageProps {
  text: string;
  timestamp?: number;
  formatTime: (ts?: number) => string;
}

export function SystemMessage({ text, timestamp, formatTime }: SystemMessageProps) {
  
  const isJoin = text.includes('se ha unido');
  const isLeave = text.includes('ha salido');
  

  
  const getColor = () => {
    if (isJoin) return 'from-emerald-500/20 to-emerald-600/10 border-emerald-500/30 text-emerald-400';
    if (isLeave) return 'from-amber-500/20 to-amber-600/10 border-amber-500/30 text-amber-400';
    return 'from-indigo-500/20 to-indigo-600/10 border-indigo-500/30 text-indigo-400';
  };

  return (
    <div className="flex justify-center my-2">
      <div className="flex items-center gap-2 max-w-[80%]">
        <div className="flex-1 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
        
        <div className={`flex items-center gap-2 px-3 py-1.5 rounded-full bg-gradient-to-r ${getColor()} border backdrop-blur-sm text-[0.7rem] font-medium`}>
          <span>{text}</span>
        </div>
        
        <div className="flex-1 h-px bg-gradient-to-l from-transparent via-white/10 to-transparent" />
      </div>
    </div>
  );
}