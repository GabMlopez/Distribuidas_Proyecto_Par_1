interface DeleteModalProps {
  show: boolean;
  onClose: () => void;
  onConfirm: () => void;
  roomName: string;
  roomId: string;
  deleting: boolean;
}

export function DeleteModal({ 
  show, 
  onClose, 
  onConfirm, 
  roomName, 
  roomId, 
  deleting 
}: DeleteModalProps) {
  if (!show) return null;

  return (
    <div 
      className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center animate-fade-in" 
      onClick={(e) => { if(e.target === e.currentTarget) onClose(); }}
    >
      <div className="w-full max-w-[400px] bg-slate-900/95 backdrop-blur-[28px] border border-white/20 rounded-2xl shadow-[0_24px_80px_rgba(0,0,0,0.7),inset_0_1px_0_rgba(255,255,255,0.14),0_0_0_1px_rgba(239,68,68,0.12)] p-6 flex flex-col gap-5 animate-fade-up">
        
        {/* Header con icono de advertencia */}
        <div className="flex flex-col items-center text-center gap-3">
          
          <div>
            <h3 className="text-xl font-bold text-red-400">Eliminar sala</h3>
            <p className="text-slate-400 text-sm mt-1">
              Esta acción no se puede deshacer.
            </p>
          </div>
        </div>

        {/* Información de la sala a eliminar */}
        <div className="bg-black/40 border border-white/10 rounded-xl p-4 text-center">
          <code className="text-xs text-slate-500">ID: {roomId}</code>
          <p className="font-semibold text-slate-200 mt-1">{roomName}</p>
          <p className="text-xs text-red-400/70 mt-2">
            Todos los mensajes y archivos serán eliminados
          </p>
        </div>

        {/* Botones de acción */}
        <div className="flex gap-3 mt-2">
          <button 
            className="flex-1 px-4 py-2.5 rounded-xl font-semibold bg-white/5 hover:bg-white/10 text-slate-300 border border-white/10 transition-all" 
            onClick={onClose}
            disabled={deleting}
          >
            Cancelar
          </button>
          <button 
            className="flex-1 px-4 py-2.5 rounded-xl font-semibold bg-gradient-to-br from-red-500 to-rose-600 text-white shadow-[0_4px_20px_rgba(239,68,68,0.3)] hover:shadow-[0_6px_28px_rgba(239,68,68,0.5)] disabled:opacity-50 disabled:shadow-none transition-all active:scale-95 flex items-center justify-center gap-2" 
            onClick={onConfirm}
            disabled={deleting}
          >
            {deleting ? (
              <>
                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                Eliminando...
              </>
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                </svg>
                Eliminar sala
              </>
            )}
          </button>
        </div>

      </div>
    </div>
  );
}