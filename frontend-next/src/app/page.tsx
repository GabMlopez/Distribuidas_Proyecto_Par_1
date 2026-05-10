'use client';

import { use, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/context/UserContext';

export default function LoginPage() {
  const router = useRouter();
  const { login } = useUser();
  
  const [mode, setMode] = useState<'user' | 'admin'>('user');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [nickname, setNickname] = useState(() => (typeof window !== 'undefined' ? localStorage.getItem('nickname') || '' : ''));
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  
  const API = process.env.NEXT_PUBLIC_API_URL ;

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    if (mode === 'admin') {
      try {
        const res = await fetch(`${API}/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ usuario: username, contrasenia: password })
        });
        const data = await res.json();
        if (res.ok) {
          login(data.token, 'Admin', data.usuario_id,true);
          console.log("admin here")
          router.push('/home');
        } else {
          setError(data.error || 'Credenciales incorrectas');
        }
      } catch {
        setError('No se pudo conectar al servidor');
      }
    } else {
      if (!nickname.trim()) {
        setError('El nickname es requerido');
      } else {
        login(null, nickname.trim(), 'user_' + Math.random().toString(36).substring(2, 11),false);
        console.log("admin here not")
        router.push('/home');
      }
    }
    setLoading(false);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-6 relative overflow-hidden">
      {/* Orbs */}
      <div className="fixed rounded-full blur-[80px] pointer-events-none animate-float w-[400px] h-[400px] bg-indigo-500/25 -top-[10%] -left-[5%] delay-0"></div>
      <div className="fixed rounded-full blur-[80px] pointer-events-none animate-float w-[350px] h-[350px] bg-pink-500/20 -bottom-[10%] -right-[5%] delay-[3s]"></div>
      <div className="fixed rounded-full blur-[80px] pointer-events-none animate-float w-[250px] h-[250px] bg-cyan-500/15 top-[50%] left-[60%] delay-[5s]"></div>

      <div className="w-full max-w-[400px] bg-slate-950/75 backdrop-blur-[28px] border border-white/20 rounded-3xl shadow-[inset_0_0_0_1px_rgba(255,255,255,0.06),0_24px_80px_rgba(0,0,0,0.7),0_0_60px_rgba(99,102,241,0.12)] p-9 flex flex-col gap-6 relative animate-fade-up">
        {/* Top highlight line */}
        <div className="absolute top-0 left-[8%] right-[8%] h-[1px] bg-gradient-to-r from-transparent via-indigo-300/60 to-transparent rounded-full"></div>

        <div className="flex items-center gap-4">
          <div className="w-13 h-13 rounded-2xl bg-gradient-to-br from-indigo-500/30 to-pink-500/20 border border-white/10 flex items-center justify-center text-indigo-300 shrink-0 shadow-[0_0_20px_rgba(99,102,241,0.25),inset_0_1px_0_rgba(255,255,255,0.1)] animate-pulse-ring">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <div>
            <h1 className="text-2xl font-bold leading-none mb-1 bg-gradient-to-br from-slate-100 via-indigo-300 to-pink-500 bg-clip-text text-transparent">Chat Distribuido</h1>
            <p className="text-sm text-slate-400">Plataforma de mensajería</p>
          </div>
        </div>

        <div className="h-[1px] bg-white/10 w-full"></div>

        {/* Tabs */}
        <div className="flex bg-black/35 border border-white/15 rounded-xl p-1 gap-1">
          <button 
            type="button" 
            className={`flex-1 p-2 rounded-lg text-sm font-medium transition-all ${mode === 'user' ? 'bg-indigo-500/20 text-indigo-200 border border-indigo-300/35 shadow-[0_0_8px_rgba(165,180,252,0.4)]' : 'text-slate-400 hover:text-slate-300 hover:bg-white/5'}`} 
            onClick={() => { setMode('user'); setError(''); }}
          >
            Invitado
          </button>
          <button 
            type="button" 
            className={`flex-1 p-2 rounded-lg text-sm font-medium transition-all ${mode === 'admin' ? 'bg-indigo-500/20 text-indigo-200 border border-indigo-300/35 shadow-[0_0_8px_rgba(165,180,252,0.4)]' : 'text-slate-400 hover:text-slate-300 hover:bg-white/5'}`} 
            onClick={() => { setMode('admin'); setError(''); }}
          >
            Admin
          </button>
        </div>

        <form onSubmit={handleLogin} className="flex flex-col gap-4">
          {mode === 'admin' ? (
            <>
              <div className="flex flex-col gap-1.5">
                <label htmlFor="usr" className="text-xs font-semibold text-slate-400 tracking-wide">Usuario</label>
                <input id="usr" type="text" value={username} onChange={e => setUsername(e.target.value)} placeholder="admin" required className="w-full p-3 bg-black/40 border border-white/20 rounded-xl text-slate-100 placeholder:text-slate-500 focus:border-indigo-400 focus:bg-indigo-500/10 focus:ring-4 focus:ring-indigo-500/20 transition-all outline-none" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label htmlFor="pwd" className="text-xs font-semibold text-slate-400 tracking-wide">Contraseña</label>
                <input id="pwd" type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="••••••••" required className="w-full p-3 bg-black/40 border border-white/20 rounded-xl text-slate-100 placeholder:text-slate-500 focus:border-indigo-400 focus:bg-indigo-500/10 focus:ring-4 focus:ring-indigo-500/20 transition-all outline-none" />
              </div>
            </>
          ) : (
            <>
              <div className="flex flex-col gap-1.5">
                <label htmlFor="nick" className="text-xs font-semibold text-slate-400 tracking-wide">Tu nickname</label>
                <input id="nick" type="text" value={nickname} onChange={e => setNickname(e.target.value)} placeholder="Ej: CoolUser99" required maxLength={20} className="w-full p-3 bg-black/40 border border-white/20 rounded-xl text-slate-100 placeholder:text-slate-500 focus:border-indigo-400 focus:bg-indigo-500/10 focus:ring-4 focus:ring-indigo-500/20 transition-all outline-none" />
              </div>
            </>
          )}

          {error && <div className="flex items-center gap-2 p-3 bg-rose-500/10 border border-rose-500/20 rounded-xl text-sm text-rose-300">{error}</div>}

          <button type="submit" disabled={loading} className="w-full mt-2 p-3.5 bg-gradient-to-br from-indigo-500 to-purple-600 text-white rounded-xl font-semibold shadow-[0_4px_20px_rgba(99,102,241,0.4)] hover:shadow-[0_6px_28px_rgba(99,102,241,0.6)] hover:-translate-y-0.5 disabled:opacity-50 disabled:transform-none disabled:shadow-none transition-all active:scale-95">
            {loading ? 'Autenticando...' : 'Entrar al chat'}
          </button>
        </form>
      </div>
    </div>
  );
}
