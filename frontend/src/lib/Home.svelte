<script>
  import { createEventDispatcher, onMount } from 'svelte';
  const dispatch = createEventDispatcher();

  export let userContext;

  let rooms = [];
  let loading = true;
  let error = '';
  let showCreateModal = false;
  let showJoinModal = false;
  let newRoomPin = '';
  let newRoomType = 'texto';
  let selectedRoomId = '';
  let joinPin = '';
  let joinError = '';
  let joining = false;
  let creating = false;

  const isAdmin = !!userContext.token;

  async function fetchRooms() {
    loading = true; error = '';
    try {
      const url = isAdmin ? 'http://localhost:8080/admin/rooms' : 'http://localhost:8080/rooms/list';
      const headers = isAdmin ? { Authorization: `Bearer ${userContext.token}` } : {};
      const res = await fetch(url, { headers });
      const data = await res.json();
      rooms = res.ok ? (data.salas || data || []) : [];
      if (!res.ok) error = data.error || 'Error al cargar salas';
    } catch {
      error = 'Error de conexión con el servidor';
    }
    loading = false;
  }

  onMount(fetchRooms);

  async function createRoom() {
    if (newRoomPin.length < 4) return;
    creating = true;
    try {
      const res = await fetch('http://localhost:8080/admin/rooms', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${userContext.token}` },
        body: JSON.stringify({ pin: newRoomPin, tipo: newRoomType })
      });
      if (res.ok) { showCreateModal = false; newRoomPin = ''; await fetchRooms(); }
      else { const d = await res.json(); alert(d.error || 'Error creando sala'); }
    } catch { alert('Error de conexión'); }
    creating = false;
  }

  function openJoin(id) { selectedRoomId = id; joinPin = ''; joinError = ''; showJoinModal = true; }

  async function joinRoom() {
    joining = true; joinError = '';
    try {
      const res = await fetch('http://localhost:8080/rooms/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sala_id: selectedRoomId, pin: joinPin, nickname: userContext.nickname, device_id: userContext.deviceId })
      });
      const data = await res.json();
      if (res.ok) {
        showJoinModal = false;
        dispatch('join', { id: selectedRoomId, token: data.token, type: data.tipo || 'texto' });
      } else {
        joinError = data.error || 'PIN incorrecto';
      }
    } catch { joinError = 'Error de conexión'; }
    joining = false;
  }
</script>

<div class="page">
  <!-- Header -->
  <header class="header">
    <div class="header-brand">
      <div class="header-icon">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
      </div>
      <span>Chat Distribuido</span>
    </div>
    <div class="header-user">
      <div class="avatar">{userContext.nickname.charAt(0).toUpperCase()}</div>
      <span class="username">{userContext.nickname}</span>
      {#if isAdmin}<span class="badge badge--admin">Admin</span>{/if}
    </div>
  </header>

  <div class="content anim-fade-up">
    <!-- Title section -->
    <div class="section-header">
      <div>
        <h2 class="page-title">Salas activas</h2>
        <p class="page-desc">{rooms.length} sala{rooms.length !== 1 ? 's' : ''} disponible{rooms.length !== 1 ? 's' : ''}</p>
      </div>
      <div class="actions">
        <button class="btn-ghost icon-btn" on:click={fetchRooms} title="Refrescar">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/></svg>
        </button>
        {#if isAdmin}
          <button class="btn-primary" on:click={() => showCreateModal = true}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            Nueva sala
          </button>
        {/if}
      </div>
    </div>

    <!-- Rooms -->
    {#if loading}
      <div class="state-container">
        <div class="loading-grid">
          {#each [1,2,3] as _}
            <div class="skeleton-card"></div>
          {/each}
        </div>
      </div>
    {:else if error}
      <div class="state-container">
        <div class="error-banner">{error}</div>
      </div>
    {:else if rooms.length === 0}
      <div class="empty-state">
        <div class="empty-icon">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        </div>
        <h3>Sin salas disponibles</h3>
        <p>{isAdmin ? 'Crea una sala para comenzar.' : 'No hay salas activas en este momento.'}</p>
      </div>
    {:else}
      <div class="rooms-grid">
        {#each rooms as room, i}
          <div class="room-card glass--card" style="animation-delay:{i*60}ms" class:anim-fade-up={true}>
            <div class="room-top">
              <div class="room-id">
                <div class="room-dot dot-{room.tipo}"></div>
                <code>{room.sala_id}</code>
              </div>
              <span class="badge badge--{room.tipo}">{room.tipo}</span>
            </div>

            {#if isAdmin}
              <div class="room-meta">
                <div class="meta-item">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
                  PIN: <strong>{room.pin}</strong>
                </div>
              </div>
            {/if}

            <div class="room-footer">
              <div class="users-count">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                {room.usuarios ?? 0} conectado{room.usuarios !== 1 ? 's' : ''}
              </div>
              <button class="btn-primary join-btn" on:click={() => openJoin(room.sala_id)}>
                Unirse
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Create Modal -->
{#if showCreateModal}
  <div class="modal-backdrop" on:click|self={() => showCreateModal = false}>
    <div class="modal">
      <div class="modal-head">
        <h3>Crear nueva sala</h3>
        <button class="btn-ghost icon-btn" on:click={() => showCreateModal = false}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
      <div class="form-field">
        <label>Tipo de sala</label>
        <div class="type-selector">
          <button class="type-btn {newRoomType === 'texto' ? 'selected' : ''}" on:click={() => newRoomType = 'texto'}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 6h16M4 12h16M4 18h7"/></svg>
            Texto
          </button>
          <button class="type-btn {newRoomType === 'multimedia' ? 'selected' : ''}" on:click={() => newRoomType = 'multimedia'}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
            Multimedia
          </button>
        </div>
      </div>
      <div class="form-field">
        <label>PIN de acceso</label>
        <input type="text" bind:value={newRoomPin} placeholder="Mínimo 4 dígitos" maxlength="6" />
      </div>
      <div class="modal-actions">
        <button class="btn-ghost" on:click={() => showCreateModal = false}>Cancelar</button>
        <button class="btn-primary" on:click={createRoom} disabled={newRoomPin.length < 4 || creating}>
          {creating ? 'Creando...' : 'Crear sala'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Join Modal -->
{#if showJoinModal}
  <div class="modal-backdrop" on:click|self={() => showJoinModal = false}>
    <div class="modal">
      <div class="modal-head">
        <div>
          <h3>Unirse a la sala</h3>
          <code class="room-id-modal">{selectedRoomId}</code>
        </div>
        <button class="btn-ghost icon-btn" on:click={() => showJoinModal = false}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>
      <div class="form-field">
        <label>PIN de acceso</label>
        <input
          type="password" bind:value={joinPin} placeholder="••••••"
          maxlength="6" on:keydown={(e) => e.key === 'Enter' && joinRoom()} autofocus />
      </div>
      {#if joinError}
        <div class="error-banner">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          {joinError}
        </div>
      {/if}
      <div class="modal-actions">
        <button class="btn-ghost" on:click={() => showJoinModal = false}>Cancelar</button>
        <button class="btn-primary" on:click={joinRoom} disabled={joining || joinPin.length < 4}>
          {#if joining}<span class="spinner"></span>{/if}
          {joining ? 'Conectando...' : 'Entrar'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page { min-height: 100vh; display: flex; flex-direction: column; }

  /* Header */
  .header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 1rem 2rem;
    background: rgba(5, 8, 19, 0.6);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border-bottom: 1px solid rgba(255,255,255,0.07);
    position: sticky; top: 0; z-index: 100;
  }
  .header-brand {
    display: flex; align-items: center; gap: 0.6rem;
    font-weight: 700; font-size: 0.95rem; color: var(--text-1);
  }
  .header-icon {
    width: 32px; height: 32px; border-radius: 8px;
    background: var(--indigo-dim);
    border: 1px solid rgba(99,102,241,0.25);
    display: flex; align-items: center; justify-content: center;
    color: var(--indigo-light);
  }
  .header-user { display: flex; align-items: center; gap: 0.6rem; }
  .avatar {
    width: 30px; height: 30px; border-radius: 50%;
    background: linear-gradient(135deg, var(--indigo), var(--pink));
    display: flex; align-items: center; justify-content: center;
    font-size: 0.8rem; font-weight: 700;
  }
  .username { font-size: 0.85rem; color: var(--text-2); }
  .badge--admin {
    background: rgba(236, 72, 153, 0.15);
    color: #f472b6;
    border: 1px solid rgba(236,72,153,0.25);
    padding: 0.15rem 0.55rem;
    border-radius: 999px;
    font-size: 0.68rem; font-weight: 700;
    text-transform: uppercase; letter-spacing: 0.06em;
  }

  /* Content */
  .content { max-width: 1100px; margin: 0 auto; width: 100%; padding: 2.5rem 2rem; flex: 1; }

  .section-header {
    display: flex; justify-content: space-between; align-items: flex-end;
    margin-bottom: 2rem;
  }
  .page-title { font-size: 1.8rem; }
  .page-desc { color: var(--text-3); font-size: 0.85rem; margin-top: 0.25rem; }
  .actions { display: flex; align-items: center; gap: 0.75rem; }

  .icon-btn { padding: 0.6rem; border-radius: 8px; }

  /* Grid */
  .rooms-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(290px, 1fr)); gap: 1.25rem; }

  .room-card {
    padding: 1.5rem;
    display: flex; flex-direction: column; gap: 1.25rem;
    cursor: default;
  }

  .room-top { display: flex; justify-content: space-between; align-items: center; }
  .room-id { display: flex; align-items: center; gap: 0.5rem; }
  .room-id code {
    font-size: 0.9rem; font-weight: 700;
    background: rgba(255,255,255,0.06);
    padding: 0.2rem 0.5rem; border-radius: 5px;
    letter-spacing: 0.05em;
  }
  .room-dot {
    width: 7px; height: 7px; border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 6px currentColor;
  }
  .dot-texto     { background: var(--indigo-light); color: var(--indigo-light); }
  .dot-multimedia{ background: #f472b6; color: #f472b6; }

  .room-meta { font-size: 0.82rem; color: var(--text-3); }
  .meta-item { display: flex; align-items: center; gap: 0.35rem; }
  .meta-item strong { color: var(--text-2); }

  .room-footer { display: flex; justify-content: space-between; align-items: center; }
  .users-count {
    display: flex; align-items: center; gap: 0.35rem;
    font-size: 0.8rem; color: var(--text-3);
  }
  .join-btn { padding: 0.5rem 1rem; font-size: 0.82rem; }

  /* States */
  .state-container { padding: 4rem 0; }
  .loading-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(290px, 1fr)); gap: 1.25rem; }
  .skeleton-card {
    height: 160px; border-radius: 16px;
    background: linear-gradient(90deg, rgba(255,255,255,0.04) 25%, rgba(255,255,255,0.07) 50%, rgba(255,255,255,0.04) 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }

  .empty-state {
    text-align: center; padding: 5rem 2rem;
    display: flex; flex-direction: column; align-items: center; gap: 1rem;
  }
  .empty-icon {
    width: 64px; height: 64px; border-radius: 16px;
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.08);
    display: flex; align-items: center; justify-content: center;
    color: var(--text-3);
  }
  .empty-state h3 { font-size: 1.1rem; }
  .empty-state p  { color: var(--text-3); font-size: 0.9rem; }

  /* Modal extras */
  .modal-head {
    display: flex; justify-content: space-between; align-items: flex-start;
  }
  .modal-head h3 { font-size: 1.2rem; }
  .room-id-modal { font-size: 0.75rem; color: var(--text-3); margin-top: 0.2rem; display: block; }
  .modal-actions { display: flex; justify-content: flex-end; gap: 0.75rem; padding-top: 0.5rem; }

  /* Type selector */
  .type-selector { display: flex; gap: 0.5rem; }
  .type-btn {
    flex: 1; padding: 0.65rem;
    background: rgba(0,0,0,0.2);
    border: 1px solid rgba(255,255,255,0.06);
    border-radius: 8px;
    color: var(--text-3);
    font-size: 0.85rem;
    flex-direction: column; gap: 0.35rem;
  }
  .type-btn:hover { background: rgba(255,255,255,0.06); color: var(--text-2); }
  .type-btn.selected {
    background: var(--indigo-dim);
    border-color: rgba(99,102,241,0.35);
    color: var(--indigo-light);
  }
  .type-btn::after { display: none; }

  .spinner {
    width: 13px; height: 13px;
    border: 2px solid rgba(255,255,255,0.25);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
