interface Props {
  show: boolean;
  onClose: () => void;
  selectedRoomId: string;
  selectedRoomNombre: string;
  editNombre: string;
  setEditNombre: (v: string) => void;
  editPin: string;
  setEditPin: (v: string) => void;
  editType: 'texto' | 'multimedia';
  setEditType: (v: 'texto' | 'multimedia') => void;
  updateRoom: () => void;
  updating: boolean;
  hasChanges: boolean;
}

export function EditRoomModal({ 
  show, 
  onClose, 
  selectedRoomId,
  selectedRoomNombre,
  editNombre,
  setEditNombre,
  editPin, 
  setEditPin, 
  editType, 
  setEditType, 
  updateRoom, 
  updating,
  hasChanges
}: Props) {
  
  const isPinValid = editPin === '' || (editPin.length >= 4 && editPin.length <= 6);

  if (!show) return null;

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/75 backdrop-blur-md flex items-center justify-center animate-fade-in" onClick={(e) => { if(e.target === e.currentTarget) onClose(); }}>
      <div className="w-full max-w-[420px] bg-slate-900/90 backdrop-blur-[28px] border border-white/20 rounded-2xl shadow-[0_24px_80px_rgba(0,0,0,0.7),inset_0_1px_0_rgba(255,255,255,0.14),0_0_0_1px_rgba(99,102,241,0.12)] p-8 flex flex-col gap-5 animate-fade-up">
        
        <div className="flex justify-between items-center pb-3 border-b border-white/10">
          <div>
            <h3 className="text-lg font-bold">Editar sala</h3>
            <code className="text-xs bg-black/50 px-2 py-1 rounded text-slate-400 font-mono mt-1 block">
              ID: {selectedRoomId}
            </code>
          </div>
          <button className="w-8 h-8 flex items-center justify-center rounded-lg bg-white/5 hover:bg-white/10 text-slate-400 hover:text-white transition-all" onClick={onClose}>✕</button>
        </div>

        {/* Campo para editar nombre */}
        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-semibold text-slate-400">Nombre de la sala</label>
          <input 
            type="text" 
            value={editNombre} 
            onChange={e => setEditNombre(e.target.value)} 
            placeholder={selectedRoomNombre}
            maxLength={30} 
            className="w-full p-3 bg-black/40 border border-white/20 rounded-xl text-slate-100 placeholder:text-slate-500 focus:border-indigo-400 focus:bg-indigo-500/10 focus:ring-4 focus:ring-indigo-500/20 transition-all outline-none box-border" 
          />
          <p className="text-xs text-slate-500">Dejar vacío para mantener el nombre actual</p>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-semibold text-slate-400">Tipo de sala</label>
          <div className="flex bg-black/40 border border-white/15 rounded-xl p-1 gap-1">
            <button 
              className={`flex-1 py-2 rounded-lg text-sm transition-all ${editType === 'texto' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-400/30' : 'text-slate-400 hover:bg-white/5'}`} 
              onClick={() => setEditType('texto')}
            >
              Texto
            </button>
            <button 
              className={`flex-1 py-2 rounded-lg text-sm transition-all ${editType === 'multimedia' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-400/30' : 'text-slate-400 hover:bg-white/5'}`} 
              onClick={() => setEditType('multimedia')}
            >
              Multimedia
            </button>
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-xs font-semibold text-slate-400">
            Nuevo PIN (opcional - solo números)
            <span className="text-slate-500 ml-1">4-6 dígitos</span>
          </label>
          <input 
            type="text" 
            value={editPin} 
            onChange={e => {
              const onlyNumbers = e.target.value.replace(/\D/g, '');
              setEditPin(onlyNumbers);
            }} 
            placeholder="Dejar vacío para no cambiar" 
            maxLength={6} 
            inputMode="numeric"  
            pattern="[0-9]*"     
            className={`w-full p-3 bg-black/40 border rounded-xl text-slate-100 placeholder:text-slate-500 focus:ring-4 transition-all outline-none box-border ${
              editPin && !isPinValid 
                ? 'border-red-500 focus:border-red-500 focus:ring-red-500/20' 
                : 'border-white/20 focus:border-indigo-400 focus:bg-indigo-500/10 focus:ring-indigo-500/20'
            }`}
          />
          {editPin && !isPinValid && (
            <p className="text-xs text-red-400">El PIN debe tener entre 4 y 6 dígitos</p>
          )}
        </div>

        <div className="flex justify-end gap-3 mt-2">
          <button 
            className="px-5 py-2.5 rounded-xl font-semibold bg-white/5 hover:bg-white/10 text-slate-300 border border-white/10 transition-all" 
            onClick={onClose}
          >
            Cancelar
          </button>
          <button 
            className="px-5 py-2.5 rounded-xl font-semibold bg-gradient-to-br from-indigo-500 to-purple-600 text-white shadow-[0_4px_20px_rgba(99,102,241,0.4)] hover:shadow-[0_6px_28px_rgba(99,102,241,0.6)] disabled:opacity-50 disabled:shadow-none transition-all active:scale-95" 
            onClick={updateRoom} 
            disabled={!hasChanges || updating || (editPin !== '' && !isPinValid)}
          >
            {updating ? 'Guardando...' : 'Guardar cambios'}
          </button>
        </div>

      </div>
    </div>
  );
}