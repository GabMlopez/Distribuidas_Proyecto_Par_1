'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/context/UserContext';
import { Room } from '@/types';
import { RoomCard } from '@/components/home/RoomCard';
import { CreateRoomModal } from '@/components/home/CreateRoomModal';
import { JoinRoomModal } from '@/components/home/JoinRoomModal';
import { EditRoomModal } from '@/components/home/EditRoomModal';

const API = 'http://localhost:8080';

export default function HomePage() {
  const router = useRouter();
  const { userContext, logout } = useUser();
  const isAdmin = userContext?.nickname === 'Admin';

  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);

  // Modales
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);

  // Estados Crear Sala
  const [salanombre, setSalanombre] = useState('');
  const [newRoomType, setNewRoomType] = useState<'texto' | 'multimedia'>('texto');
  const [newRoomPin, setNewRoomPin] = useState('');
  const [creating, setCreating] = useState(false);

  // Estados Unirse
  const [selectedRoom, setSelectedRoom] = useState<Room | null>(null);
  const [joinPin, setJoinPin] = useState('');
  const [joining, setJoining] = useState(false);
  const [joinError, setJoinError] = useState('');

  // Estados Editar
  const [editRoomRef, setEditRoomRef] = useState<Room | null>(null);
  const [editType, setEditType] = useState<'texto' | 'multimedia'>('texto');
  const [editPin, setEditPin] = useState('');
  const [updating, setUpdating] = useState(false);

  const fetchRooms = useCallback(async () => {
    try {
      const res = await fetch(`${API}/rooms/list`);
      if (res.ok) {
        const data = await res.json();
        setRooms(data || []);
      }
    } catch (e) {
      console.error('Error fetching rooms:', e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!userContext?.nickname) {
      router.push('/');
      return;
    }
    fetchRooms();
    const interval = setInterval(fetchRooms, 3000);
    return () => clearInterval(interval);
  }, [userContext, fetchRooms, router]);

  const handleLogout = () => {
    logout();
    router.push('/');
  };

  const createRoom = async () => {
    setCreating(true);
    try {
      const res = await fetch(`${API}/rooms/create`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${userContext.token}` },
        body: JSON.stringify({ nombre: salanombre, tipo: newRoomType, pin: newRoomPin })
      });
      if (res.ok) {
        setShowCreateModal(false);
        setSalanombre(''); setNewRoomPin('');
        fetchRooms();
      }
    } catch {}
    setCreating(false);
  };

  const deleteRoom = async (id: string) => {
    if (!confirm('¿Seguro que deseas eliminar esta sala?')) return;
    try {
      await fetch(`${API}/rooms/delete`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${userContext.token}` },
        body: JSON.stringify({ sala_id: id })
      });
      fetchRooms();
    } catch {}
  };

  const updateRoom = async () => {
    if (!editRoomRef) return;
    setUpdating(true);
    try {
      await fetch(`${API}/rooms/update`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${userContext.token}` },
        body: JSON.stringify({ sala_id: editRoomRef.sala_id, tipo: editType, pin: editPin })
      });
      setShowEditModal(false);
      fetchRooms();
    } catch {}
    setUpdating(false);
  };

  const joinRoom = async () => {
    if (!selectedRoom) return;
    setJoinError(''); setJoining(true);
    try {
      const res = await fetch(`${API}/rooms/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          sala_id: selectedRoom.sala_id,
          pin: joinPin,
          nickname: userContext.nickname,
          device_id: userContext.deviceId
        })
      });
      const data = await res.json();
      if (res.ok) {
        sessionStorage.setItem('room_token', data.token);
        sessionStorage.setItem('room_name', selectedRoom.nombre || selectedRoom.sala_id);
        sessionStorage.setItem('room_type', selectedRoom.tipo);
        router.push(`/room/${selectedRoom.sala_id}`);
      } else {
        setJoinError(data.error || 'PIN incorrecto');
      }
    } catch {
      setJoinError('Error de conexión');
    }
    setJoining(false);
  };

  if (!userContext?.nickname) return null;

  return (
    <div className="min-h-screen flex flex-col p-5 md:p-8 relative">
      <header className="flex justify-between items-center px-6 py-4 bg-slate-900/40 backdrop-blur-xl border border-white/15 rounded-2xl shadow-[0_8px_32px_rgba(0,0,0,0.5)] mb-8 shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500/20 to-pink-500/10 border border-indigo-400/30 flex items-center justify-center text-indigo-300 shadow-[0_0_15px_rgba(99,102,241,0.2)]">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <h1 className="text-xl font-bold bg-gradient-to-br from-slate-100 to-slate-400 bg-clip-text text-transparent">Salas de Chat</h1>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2.5 px-3.5 py-1.5 bg-black/30 border border-white/10 rounded-full">
            <div className="w-6 h-6 rounded-full bg-gradient-to-br from-indigo-500 to-pink-500 flex items-center justify-center text-[0.65rem] font-bold text-white shadow-inner">
              {userContext.nickname.charAt(0).toUpperCase()}
            </div>
            <span className="text-sm font-semibold text-slate-300">{userContext.nickname}</span>
          </div>
          <button className="px-4 py-2 bg-rose-500/10 text-rose-400 font-semibold text-sm rounded-lg border border-rose-500/20 hover:bg-rose-500/20 hover:border-rose-500/40 transition-all" onClick={handleLogout}>Salir</button>
        </div>
      </header>

      <main className="flex-1 max-w-[1200px] w-full mx-auto relative z-10 flex flex-col">
        <div className="flex justify-between items-center mb-8 shrink-0">
          <h2 className="text-2xl font-bold text-slate-100">Explorar Salas</h2>
          {isAdmin && (
            <button className="px-5 py-2.5 bg-gradient-to-br from-indigo-500 to-purple-600 text-white font-semibold text-sm rounded-xl shadow-[0_4px_20px_rgba(99,102,241,0.4)] hover:shadow-[0_6px_28px_rgba(99,102,241,0.6)] hover:-translate-y-0.5 transition-all active:scale-95" onClick={() => setShowCreateModal(true)}>
              + Nueva Sala
            </button>
          )}
        </div>

        {loading ? (
          <div className="flex-1 flex flex-col items-center justify-center gap-4 opacity-70">
            <div className="w-8 h-8 border-4 border-white/20 border-t-indigo-400 rounded-full animate-spin"></div>
            <p className="text-slate-400 font-medium">Cargando salas...</p>
          </div>
        ) : rooms.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center gap-5 opacity-70">
            <div className="w-16 h-16 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center text-slate-500">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="9" y1="3" x2="9" y2="21"/></svg>
            </div>
            <p className="text-slate-400">No hay salas disponibles.</p>
            {isAdmin && <button className="px-5 py-2.5 bg-white/5 border border-white/10 text-slate-300 font-semibold rounded-xl hover:bg-white/10 hover:border-white/20 transition-all" onClick={() => setShowCreateModal(true)}>Crear la primera sala</button>}
          </div>
        ) : (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-6 auto-rows-max p-2 pb-10">
            {rooms.map((room, index) => (
              <RoomCard
                key={room.sala_id}
                room={room}
                index={index}
                isAdmin={isAdmin}
                openJoin={(r) => { setSelectedRoom(r); setJoinPin(''); setJoinError(''); setShowJoinModal(true); }}
                openEdit={(r) => { setEditRoomRef(r); setEditType(r.tipo); setEditPin(''); setShowEditModal(true); }}
                deleteRoom={deleteRoom}
              />
            ))}
          </div>
        )}
      </main>

      <CreateRoomModal
        show={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        salanombre={salanombre}
        setSalanombre={setSalanombre}
        newRoomPin={newRoomPin}
        setNewRoomPin={setNewRoomPin}
        newRoomType={newRoomType}
        setNewRoomType={setNewRoomType}
        createRoom={createRoom}
        creating={creating}
      />

      <JoinRoomModal
        show={showJoinModal}
        onClose={() => setShowJoinModal(false)}
        selectedRoomName={selectedRoom?.nombre || selectedRoom?.sala_id || ''}
        joinPin={joinPin}
        setJoinPin={setJoinPin}
        joinRoom={joinRoom}
        joining={joining}
        joinError={joinError}
      />

      <EditRoomModal
        show={showEditModal}
        onClose={() => setShowEditModal(false)}
        selectedRoomId={editRoomRef?.sala_id || ''}
        editPin={editPin}
        setEditPin={setEditPin}
        editType={editType}
        setEditType={setEditType}
        updateRoom={updateRoom}
        updating={updating}
      />
    </div>
  );
}
