'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useUser } from '@/context/UserContext';
import { Room } from '@/types';
import { RoomCard } from '@/components/home/RoomCard';
import { CreateRoomModal } from '@/components/home/CreateRoomModal';
import { JoinRoomModal } from '@/components/home/JoinRoomModal';
import { EditRoomModal } from '@/components/home/EditRoomModal';
import { DeleteModal } from '@/components/home/DeleteRoomModal'; // Importar el modal

const API = process.env.NEXT_PUBLIC_API_URL;

export default function HomePage() {
  const router = useRouter();
  const { userContext, logout, setUserId, login,updateToken } = useUser();
  const isAdmin = userContext.isAdmin;

  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);

  // Modales
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false); 

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
  const [editNombre, setEditNombre] = useState('');
  const [editType, setEditType] = useState<'texto' | 'multimedia'>('texto');
  const [editPin, setEditPin] = useState('');
  const [updating, setUpdating] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  // Estados Eliminar
  const [deleteRoomRef, setDeleteRoomRef] = useState<Room | null>(null);
  const [deleting, setDeleting] = useState(false);

  const url = isAdmin ? `${API}/admin/rooms` : `${API}/rooms/list`;

  const fetchRooms = useCallback(async () => {
    try {
      const res = isAdmin ? await fetch(url, {
        headers: { 'Authorization': `Bearer ${userContext.token}` }
      }) : await fetch(url);
      if (res.ok) {
        const data = await res.json();
        if (isAdmin) {
          setRooms(data.salas || []);
        } else {
          setRooms(Array.isArray(data) ? data : (data.salas || []));
        }
      }
    } catch (e) {
      console.error('Error fetching rooms:', e);
    } finally {
      setLoading(false);
    }
  }, [isAdmin, url, userContext.token]);

  useEffect(() => {
    if (!userContext?.nickname) {
      router.push('/');
      return;
    }
    fetchRooms();
    const interval = setInterval(fetchRooms, 5000);
    return () => clearInterval(interval);
  }, [userContext, fetchRooms, router]);

  // Validar cambios en edición
  useEffect(() => {
    if (editRoomRef) {
      const pinChanged = editPin !== '' && editPin !== editRoomRef.pin;
      const typeChanged = editType !== editRoomRef.tipo;
      const nombreChanged = editNombre !== '' && editNombre !== editRoomRef.nombre;
      setHasChanges(pinChanged || typeChanged || nombreChanged);
    } else {
      setHasChanges(false);
    }
  }, [editPin, editType, editNombre, editRoomRef]);

  const handleLogout = () => {
    logout();
    router.push('/');
  };

  const createRoom = async () => {
    setCreating(true);
    try {
      const res = await fetch(`${API}/admin/rooms`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${userContext.token}` },
        body: JSON.stringify({ nombre: salanombre, tipo: newRoomType, pin: newRoomPin })
      });
      if (res.ok) {
        setShowCreateModal(false);
        setSalanombre('');
        setNewRoomPin('');
        fetchRooms();
      }
    } catch (error) {
      console.error('Error creating room:', error);
    } finally {
      setCreating(false);
    }
  };

  const openDeleteModal = (room: Room) => {
    setDeleteRoomRef(room);
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    if (!deleteRoomRef) return;
    
    setDeleting(true);
    try {
      const res = await fetch(`${API}/admin/rooms/${deleteRoomRef.sala_id}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${userContext.token}` }
      });
      
      if (res.ok) {
        setShowDeleteModal(false);
        setDeleteRoomRef(null);
        fetchRooms();
      } else {
        const error = await res.json();
        alert(error.error || 'Error al eliminar la sala');
      }
    } catch (error) {
      console.error('Error deleting room:', error);
      alert('Error de conexión al eliminar la sala');
    } finally {
      setDeleting(false);
    }
  };

  const updateRoom = async () => {
    if (!editRoomRef) return;
    
    if (!hasChanges) {
      alert('No hay cambios para guardar');
      return;
    }
    
    if (editPin && (editPin.length < 4 || editPin.length > 6)) {
      alert('El PIN debe tener entre 4 y 6 dígitos');
      return;
    }
    
    setUpdating(true);
    try {
      const updateData: { pin?: string; tipo?: string; nombre?: string } = {};
      
      if (editPin && editPin !== editRoomRef.pin) {
        updateData.pin = editPin;
      }
      if (editType !== editRoomRef.tipo) {
        updateData.tipo = editType;
      }
      if (editNombre && editNombre !== editRoomRef.nombre) {
        updateData.nombre = editNombre;
      }
      
      const res = await fetch(`${API}/admin/rooms/${editRoomRef.sala_id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${userContext.token}` },
        body: JSON.stringify(updateData)
      });
      
      if (res.ok) {
        setShowEditModal(false);
        setEditRoomRef(null);
        setEditPin('');
        setEditNombre('');
        setEditType('texto');
        setHasChanges(false);
        fetchRooms();
      } else {
        const error = await res.json();
        alert(error.error || 'Error al actualizar la sala');
      }
    } catch (error) {
      console.error('Error updating room:', error);
      alert('Error de conexión al actualizar la sala');
    } finally {
      setUpdating(false);
    }
  };

const joinRoom = async () => {
  if (!selectedRoom) return;
  setJoinError('');
  setJoining(true);
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
      const backendUserId = data.usuario_id;
      console.log(data)
      // Guardar token de sala SEPARADAMENTE
      sessionStorage.setItem('room_token', data.token);
      sessionStorage.setItem('room_name', selectedRoom.nombre || selectedRoom.sala_id);
      sessionStorage.setItem('room_type', selectedRoom.tipo);
      sessionStorage.setItem('room_user_id', backendUserId);
      
      if (!userContext.isAdmin) {
        login(userContext.token, userContext.nickname, backendUserId, userContext.isAdmin);
        updateToken(data.token);
        setUserId(backendUserId);
      }
      
      router.push(`/room/${selectedRoom.sala_id}`);
    } else {
      setJoinError(data.error || 'PIN incorrecto');
    }
  } catch {
    setJoinError('Error de conexión');
  }
  setJoining(false);
};

  const openEditModal = (room: Room) => {
    setEditRoomRef(room);
    setEditNombre(room.nombre || '');
    setEditType(room.tipo);
    setEditPin('');
    setHasChanges(false);
    setShowEditModal(true);
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

        {/* Lista de salas */}
        {loading ? (
          <div className="flex-1 flex flex-col items-center justify-center gap-4 opacity-70">
            <div className="w-8 h-8 border-4 border-white/20 border-t-indigo-400 rounded-full animate-spin"></div>
            <p className="text-slate-400 font-medium">Cargando salas...</p>
          </div>
        ) : rooms.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center gap-5 opacity-70">
            <div className="w-16 h-16 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center text-slate-500">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
                <line x1="9" y1="3" x2="9" y2="21"/>
              </svg>
            </div>
            <p className="text-slate-400">No hay salas disponibles.</p>
            {isAdmin && (
              <button className="px-5 py-2.5 bg-white/5 border border-white/10 text-slate-300 font-semibold rounded-xl hover:bg-white/10 hover:border-white/20 transition-all" onClick={() => setShowCreateModal(true)}>
                Crear la primera sala
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] md:grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-9 auto-rows-max p-2 pb-10">
            {rooms.map((room, index) => (
              <RoomCard
                key={room.sala_id}
                room={room}
                index={index}
                isAdmin={isAdmin}
                openJoin={(r) => { setSelectedRoom(r); setJoinPin(''); setJoinError(''); setShowJoinModal(true); }}
                openEdit={openEditModal}
                deleteRoom={openDeleteModal} 
              />
            ))}
          </div>
        )}
      </main>

      {/* Modales */}
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
        onClose={() => {
          setShowEditModal(false);
          setEditRoomRef(null);
          setEditPin('');
          setEditNombre('');
          setHasChanges(false);
        }}
        selectedRoomId={editRoomRef?.sala_id || ''}
        selectedRoomNombre={editRoomRef?.nombre || ''}
        editNombre={editNombre}
        setEditNombre={setEditNombre}
        editPin={editPin}
        setEditPin={setEditPin}
        editType={editType}
        setEditType={setEditType}
        updateRoom={updateRoom}
        updating={updating}
        hasChanges={hasChanges}
      />

      
      <DeleteModal
        show={showDeleteModal}
        onClose={() => {
          setShowDeleteModal(false);
          setDeleteRoomRef(null);
        }}
        onConfirm={confirmDelete}
        roomName={deleteRoomRef?.nombre || deleteRoomRef?.sala_id || ''}
        roomId={deleteRoomRef?.sala_id || ''}
        deleting={deleting}
      />
    </div>
  );
}