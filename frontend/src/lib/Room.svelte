<script>
  import { createEventDispatcher, onMount, onDestroy, tick } from 'svelte';
  const dispatch = createEventDispatcher();

  export let userContext;
  export let room;

  let roomDisplayName = room.nombre || room.id;
  const API = 'http://localhost:8080';

  let ws;
  let messages = [];
  let currentMessage = '';
  let users = [];
  let fileInput;
  let chatBody;
  let wsStatus = 'connecting';

  // Subida paralela: lista de uploads activos
  let uploads = []; // [{ name, progress, done, error }]

  onMount(() => {
    connectWebSocket();
    // Desconexión automática al cerrar la pestaña/navegador
    window.addEventListener('beforeunload', handleBeforeUnload);
  });

  onDestroy(() => {
    window.removeEventListener('beforeunload', handleBeforeUnload);
    if (ws) ws.close();
  });

  function handleBeforeUnload() {
    // sendBeacon garantiza que la petición se complete aunque la página se cierre
    const payload = JSON.stringify({
      sala_id: room.id,
      nickname: userContext.nickname,
      device_id: userContext.deviceId
    });
    navigator.sendBeacon(`${API}/rooms/leave`, new Blob([payload], { type: 'application/json' }));
    if (ws) ws.close();
  }

  async function scrollToBottom() {
    await tick();
    if (chatBody) chatBody.scrollTop = chatBody.scrollHeight;
  }

  function connectWebSocket() {
    wsStatus = 'connecting';
    const wsProtocol = API.startsWith('https') ? 'wss' : 'ws';
    const wsHost = API.replace(/^https?:\/\//, '');
    const url = `${wsProtocol}://${wsHost}/ws/${room.id}?nickname=${encodeURIComponent(userContext.nickname)}&sala_id=${room.id}`;
    ws = new WebSocket(url);

    ws.onopen  = () => { wsStatus = 'open'; };
    ws.onclose = () => { wsStatus = 'closed'; };

    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        if (data.type === 'error' || data.tipo === 'error') {
          alert('Error: ' + data.texto);
          wsStatus = 'closed';
          dispatch('leave'); // Opcional, forzar la salida a la lista de salas
          return;
        } else if (data.type === 'user_list' || data.tipo === 'user_list') {
          try {
            const parsed = typeof data.texto === 'string' ? JSON.parse(data.texto) : data;
            users = parsed.users || [];
          } catch { users = data.users || []; }
        } else {
          messages = [...messages, data];
          scrollToBottom();
        }
      } catch (err) { console.error(err); }
    };
  }

  function sendMessage() {
    if (!currentMessage.trim() || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ tipo: 'chat', texto: currentMessage.trim() }));
    currentMessage = '';
  }

  function handleKeydown(e) {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendMessage(); }
  }

  // Subida paralela de múltiples archivos
  async function handleFilesSelected() {
    if (!fileInput?.files?.length) return;
    const files = Array.from(fileInput.files);
    fileInput.value = '';

    // Lanzar todas las subidas en paralelo
    const promises = files.map((file) => uploadSingleFile(file));
    await Promise.allSettled(promises);
  }

  async function uploadSingleFile(file) {
    const id = Math.random().toString(36).slice(2);
    const entry = { id, name: file.name, progress: 0, done: false, error: null };
    uploads = [...uploads, entry];

    return new Promise((resolve) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      formData.append('sala_id', room.id);
      formData.append('file', file);

      xhr.upload.onprogress = (ev) => {
        if (ev.lengthComputable) {
          const pct = Math.round((ev.loaded / ev.total) * 100);
          uploads = uploads.map(u => u.id === id ? { ...u, progress: pct } : u);
        }
      };

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          uploads = uploads.map(u => u.id === id ? { ...u, progress: 100, done: true } : u);
          setTimeout(() => {
            uploads = uploads.filter(u => u.id !== id);
          }, 2500);
          resolve({ ok: true });
        } else {
          let errorMsg = 'Error al subir';
          try {
            const resp = JSON.parse(xhr.responseText);
            errorMsg = resp.error || errorMsg;
          } catch (e) {}
          uploads = uploads.map(u => u.id === id ? { ...u, error: errorMsg } : u);
          setTimeout(() => { uploads = uploads.filter(u => u.id !== id); }, 5000);
          resolve({ ok: false });
        }
      };

      xhr.onerror = () => {
        uploads = uploads.map(u => u.id === id ? { ...u, error: 'Sin conexión' } : u);
        setTimeout(() => { uploads = uploads.filter(u => u.id !== id); }, 4000);
        resolve({ ok: false });
      };

      xhr.open('POST', `${API}/upload/file`);
      xhr.setRequestHeader('Authorization', `Bearer ${room.token}`);
      xhr.send(formData);
    });
  }

  async function leaveRoom() {
    try {
      await fetch(`${API}/rooms/leave`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          sala_id: room.id,
          nickname: userContext.nickname,
          device_id: userContext.deviceId
        })
      });
    } catch {}
    if (ws) ws.close();
    dispatch('leave');
  }

  function formatTime(ts) {
    const d = ts ? new Date(ts * 1000) : new Date();
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  function isOwn(msg) { return msg.nickname === userContext.nickname; }
</script>

<div class="room-page">
  <!-- Sidebar -->
  <aside class="sidebar">
    <!-- Room info -->
    <div class="sidebar-section room-info">
      <div class="room-title-row">
        <div class="room-icon">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        </div>
        <div>
          <div class="room-name">{roomDisplayName}</div>
          <span class="badge badge--{room.type}">{room.type}</span>
        </div>
      </div>
      <div class="ws-status ws-{wsStatus}">
        <div class="status-dot"></div>
        {wsStatus === 'open' ? 'Conectado' : wsStatus === 'connecting' ? 'Conectando…' : 'Desconectado'}
      </div>
    </div>

    <div class="sidebar-divider"></div>

    <!-- Users list -->
    <div class="sidebar-section users-section">
      <div class="section-label">Participantes · {users.length}</div>
      <div class="users-list">
        {#each users as user}
          <div class="user-item" class:is-me={user === userContext.nickname}>
            <div class="user-avatar" class:avatar-me={user === userContext.nickname}>
              {user.charAt(0).toUpperCase()}
            </div>
            <span class="user-name">{user}{user === userContext.nickname ? ' (tú)' : ''}</span>
            {#if user === userContext.nickname}
              <div class="online-dot"></div>
            {/if}
          </div>
        {/each}
      </div>
    </div>

    <!-- Upload progress panel -->
    {#if uploads.length > 0}
      <div class="sidebar-divider"></div>
      <div class="sidebar-section uploads-panel">
        <div class="section-label">Subiendo archivos</div>
        {#each uploads as up (up.id)}
          <div class="upload-item">
            <div class="upload-name">{up.name}</div>
            {#if up.error}
              <div class="upload-error">{up.error}</div>
            {:else}
              <div class="progress-track">
                <div class="progress-fill" class:done={up.done} style="width:{up.progress}%"></div>
              </div>
              <span class="upload-pct">{up.done ? '✓' : up.progress + '%'}</span>
            {/if}
          </div>
        {/each}
      </div>
    {/if}

    <div class="sidebar-footer">
      <button class="btn-danger leave-btn" on:click={leaveRoom}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
        Salir de la sala
      </button>
    </div>
  </aside>

  <!-- Chat area -->
  <main class="chat-area">
    <!-- Chat header -->
    <div class="chat-header">
      <div class="chat-title">
        <div class="room-dot dot-{room.type}"></div>
        <span> <strong>{roomDisplayName}</strong></span>
      </div>
      <div class="chat-meta">{messages.length} mensajes</div>
    </div>

    <!-- Messages -->
    <div class="chat-body" bind:this={chatBody}>
      {#if messages.length === 0}
        <div class="empty-chat">
          <div class="empty-icon">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
          </div>
          <p>Sé el primero en enviar un mensaje</p>
        </div>
      {/if}

      {#each messages as msg}
        {#if msg.tipo === 'join' || msg.tipo === 'leave'}
          <div class="system-msg">
            <div class="system-line"></div>
            <span>{msg.texto}</span>
            <div class="system-line"></div>
          </div>

        {:else if msg.tipo === 'multimedia'}
          <div class="msg-row {isOwn(msg) ? 'row-end' : 'row-start'}">
            {#if !isOwn(msg)}
              <div class="msg-avatar">{msg.nickname?.charAt(0)?.toUpperCase()}</div>
            {/if}
            <div class="bubble bubble-media">
              {#if !isOwn(msg)}<span class="bubble-sender">{msg.nickname}</span>{/if}
              <p>{msg.texto}</p>
              {#if msg.file_url}
                <a href={API + msg.file_url} target="_blank" class="file-chip">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  Descargar archivo
                </a>
              {/if}
              <span class="bubble-time">{formatTime(msg.timestamp)}</span>
            </div>
          </div>

        {:else}
          <div class="msg-row {isOwn(msg) ? 'row-end' : 'row-start'}">
            {#if !isOwn(msg)}
              <div class="msg-avatar">{msg.nickname?.charAt(0)?.toUpperCase()}</div>
            {/if}
            <div class="bubble {isOwn(msg) ? 'bubble-own' : 'bubble-other'}">
              {#if !isOwn(msg)}<span class="bubble-sender">{msg.nickname}</span>{/if}
              <p>{msg.texto}</p>
              <span class="bubble-time">{formatTime(msg.timestamp)}</span>
            </div>
          </div>
        {/if}
      {/each}
    </div>

    <!-- Input bar -->
    <div class="chat-input-bar">
      {#if room.type === 'multimedia'}
        <label class="attach-btn" title="Adjuntar archivos (múltiples)">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48"/></svg>
          <input type="file" bind:this={fileInput} on:change={handleFilesSelected} multiple hidden />
        </label>
      {/if}

      <div class="input-pill">
        <textarea
          bind:value={currentMessage}
          on:keydown={handleKeydown}
          placeholder="Escribe un mensaje… (Enter para enviar)"
          rows="1"
        ></textarea>
        <button
          class="send-btn"
          class:active={currentMessage.trim()}
          on:click={sendMessage}
          disabled={!currentMessage.trim() || wsStatus !== 'open'}
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
        </button>
      </div>
    </div>
  </main>
</div>

<style>
  .room-page {
    display: grid;
    grid-template-columns: 270px 1fr;
    height: 100vh;
    overflow: hidden;
  }

  /* ── Sidebar ─────────────────────────────── */
  .sidebar {
    background: rgba(3, 5, 15, 0.75);
    backdrop-filter: blur(24px);
    -webkit-backdrop-filter: blur(24px);
    border-right: 1px solid rgba(255,255,255,0.18);
    display: flex; flex-direction: column;
    overflow: hidden;
  }

  .sidebar-section { padding: 1.25rem; }
  .sidebar-divider { height: 1px; background: rgba(255,255,255,0.14); margin: 0; flex-shrink: 0; }

  .room-title-row { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.75rem; }
  .room-icon {
    width: 38px; height: 38px; border-radius: 10px;
    background: var(--indigo-dim);
    border: 1px solid rgba(99,102,241,0.25);
    display: flex; align-items: center; justify-content: center;
    color: var(--indigo-light); flex-shrink: 0;
  }
  .room-name { font-weight: 700; font-size: 0.95rem; letter-spacing: 0.02em; margin-bottom: 0.25rem; }

  .ws-status {
    display: inline-flex; align-items: center; gap: 0.4rem;
    font-size: 0.75rem; padding: 0.25rem 0.6rem;
    border-radius: 999px;
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.07);
    color: var(--text-3);
    width: fit-content;
  }
  .ws-open { color: var(--green); border-color: rgba(16,185,129,0.2); background: rgba(16,185,129,0.08); }
  .ws-connecting { color: #fbbf24; border-color: rgba(251,191,36,0.2); }
  .ws-closed { color: var(--red); border-color: rgba(244,63,94,0.2); }

  .status-dot {
    width: 6px; height: 6px; border-radius: 50%; background: currentColor;
    box-shadow: 0 0 4px currentColor;
  }
  .ws-open .status-dot { animation: pulse-ring 2s infinite; }

  /* Users */
  .users-section { flex: 1; overflow: hidden; display: flex; flex-direction: column; min-height: 0; }
  .section-label { font-size: 0.7rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-3); margin-bottom: 0.75rem; }

  .users-list { display: flex; flex-direction: column; gap: 0.25rem; overflow-y: auto; flex: 1; }
  .user-item {
    display: flex; align-items: center; gap: 0.6rem;
    padding: 0.5rem 0.6rem; border-radius: 8px;
    transition: background 0.15s;
  }
  .user-item:hover { background: rgba(255,255,255,0.04); }
  .user-item.is-me { background: rgba(99,102,241,0.07); }

  .user-avatar {
    width: 28px; height: 28px; border-radius: 50%;
    background: rgba(255,255,255,0.08);
    display: flex; align-items: center; justify-content: center;
    font-size: 0.72rem; font-weight: 700; flex-shrink: 0;
  }
  .avatar-me { background: linear-gradient(135deg, var(--indigo), var(--pink)); }
  .user-name { font-size: 0.83rem; color: var(--text-2); flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .is-me .user-name { color: var(--text-1); }
  .online-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--green); box-shadow: 0 0 5px var(--green); flex-shrink: 0; }

  /* Upload progress */
  .uploads-panel { flex-shrink: 0; }
  .upload-item { margin-bottom: 0.6rem; }
  .upload-name { font-size: 0.72rem; color: var(--text-2); margin-bottom: 0.25rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .progress-track { height: 4px; background: rgba(255,255,255,0.08); border-radius: 99px; overflow: hidden; }
  .progress-fill { height: 100%; border-radius: 99px; background: linear-gradient(90deg, var(--indigo), #7c3aed); transition: width 0.2s ease; }
  .progress-fill.done { background: var(--green); }
  .upload-pct { font-size: 0.65rem; color: var(--text-3); margin-top: 0.15rem; display: block; text-align: right; }
  .upload-error { font-size: 0.7rem; color: var(--red); margin-top: 0.15rem; }

  /* Footer */
  .sidebar-footer {
    padding: 1rem 1.25rem;
    border-top: 1px solid rgba(255,255,255,0.10);
    flex-shrink: 0;
  }
  .leave-btn { width: 100%; font-size: 0.82rem; }

  /* ── Chat ────────────────────────────────── */
  .chat-area {
    display: flex; flex-direction: column; overflow: hidden;
    background: rgba(10, 15, 30, 0.3);
  }

  .chat-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 1rem 1.5rem;
    background: rgba(3, 5, 15, 0.65);
    backdrop-filter: blur(16px);
    border-bottom: 1px solid rgba(255,255,255,0.16);
    flex-shrink: 0;
  }
  .chat-title { display: flex; align-items: center; gap: 0.6rem; font-size: 0.9rem; }
  .chat-meta { font-size: 0.75rem; color: var(--text-3); }
  .room-dot { width: 8px; height: 8px; border-radius: 50%; box-shadow: 0 0 6px currentColor; }
  .dot-texto { background: var(--indigo-light); color: var(--indigo-light); }
  .dot-multimedia { background: #f472b6; color: #f472b6; }

  .chat-body { flex: 1; overflow-y: auto; padding: 1.5rem; display: flex; flex-direction: column; gap: 0.75rem; }

  .empty-chat {
    flex: 1; display: flex; flex-direction: column;
    align-items: center; justify-content: center; gap: 0.75rem;
    color: var(--text-3); opacity: 0.7;
  }
  .empty-icon {
    width: 56px; height: 56px; border-radius: 14px;
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(255,255,255,0.07);
    display: flex; align-items: center; justify-content: center;
  }
  .empty-chat p { font-size: 0.85rem; }

  .system-msg {
    display: flex; align-items: center; gap: 0.75rem;
    margin: 0.25rem 0;
  }
  .system-line { flex: 1; height: 1px; background: rgba(255,255,255,0.06); }
  .system-msg span { font-size: 0.72rem; color: var(--text-3); white-space: nowrap; }

  .msg-row { display: flex; align-items: flex-end; gap: 0.5rem; max-width: 75%; }
  .row-start { align-self: flex-start; }
  .row-end   { align-self: flex-end; flex-direction: row-reverse; }

  .msg-avatar {
    width: 26px; height: 26px; border-radius: 50%;
    background: rgba(255,255,255,0.08);
    display: flex; align-items: center; justify-content: center;
    font-size: 0.68rem; font-weight: 700; flex-shrink: 0; margin-bottom: 2px;
  }

  .bubble {
    padding: 0.65rem 0.9rem;
    border-radius: 14px;
    display: flex; flex-direction: column; gap: 0.3rem;
    max-width: 100%;
    word-break: break-word;
  }
  .bubble-other {
    background: rgba(255,255,255,0.07);
    border: 1px solid rgba(255,255,255,0.18);
    border-bottom-left-radius: 3px;
  }
  .bubble-own {
    background: linear-gradient(135deg, var(--indigo), #7c3aed);
    border-bottom-right-radius: 3px;
    box-shadow: 0 4px 16px rgba(99,102,241,0.3);
  }
  .bubble-media {
    background: rgba(236,72,153,0.08);
    border: 1px solid rgba(236,72,153,0.2);
    border-radius: 14px;
  }

  .bubble-sender { font-size: 0.7rem; font-weight: 700; color: var(--indigo-light); margin-bottom: 0.1rem; }
  .bubble-own .bubble-sender { display: none; }
  .bubble p { font-size: 0.88rem; line-height: 1.45; color: var(--text-1); margin: 0; }
  .bubble-time { font-size: 0.63rem; color: rgba(255,255,255,0.35); align-self: flex-end; }

  .file-chip {
    display: inline-flex; align-items: center; gap: 0.4rem;
    padding: 0.4rem 0.7rem;
    background: rgba(236,72,153,0.15);
    border: 1px solid rgba(236,72,153,0.25);
    border-radius: 8px;
    color: #f472b6;
    font-size: 0.78rem;
    text-decoration: none;
    transition: background 0.2s;
    margin-top: 0.2rem;
  }
  .file-chip:hover { background: rgba(236,72,153,0.25); }

  /* Input bar */
  .chat-input-bar {
    padding: 0.85rem 1.25rem;
    background: rgba(3, 5, 15, 0.75);
    backdrop-filter: blur(20px);
    border-top: 1px solid rgba(255,255,255,0.16);
    display: flex; align-items: center; gap: 0.6rem;
    flex-shrink: 0;
  }

  .attach-btn {
    width: 40px; height: 40px; flex-shrink: 0;
    border-radius: 10px;
    background: rgba(255,255,255,0.05);
    border: 1px solid rgba(255,255,255,0.15);
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; color: var(--text-3);
    transition: all 0.2s;
  }
  .attach-btn:hover { background: rgba(255,255,255,0.12); color: var(--text-2); border-color: rgba(255,255,255,0.28); }

  .input-pill {
    flex: 1; display: flex; align-items: center; gap: 0.5rem;
    background: rgba(0, 0, 0, 0.45);
    border: 1px solid rgba(255,255,255,0.18);
    border-radius: 14px;
    padding: 0.4rem 0.4rem 0.4rem 1rem;
    transition: border-color 0.2s;
    min-width: 0;
  }
  .input-pill:focus-within {
    border-color: rgba(165,180,252,0.50);
    box-shadow: 0 0 12px rgba(99,102,241,0.18);
  }

  textarea {
    flex: 1; background: transparent; border: none; resize: none;
    font-size: 0.9rem; padding: 0.25rem 0; max-height: 120px;
    line-height: 1.45; outline: none; box-shadow: none;
    color: var(--text-1); min-width: 0;
  }

  .send-btn {
    width: 36px; height: 36px; border-radius: 10px; padding: 0;
    background: rgba(255,255,255,0.05);
    border: 1px solid rgba(255,255,255,0.10);
    color: var(--text-3);
    flex-shrink: 0;
    transition: all 0.2s var(--ease-out);
  }
  .send-btn.active {
    background: linear-gradient(135deg, var(--indigo), #7c3aed);
    border-color: transparent;
    color: white;
    box-shadow: 0 4px 14px rgba(99,102,241,0.4);
  }
  .send-btn.active:hover { transform: scale(1.08); }
  .send-btn::after { display: none; }
  .send-btn:disabled { opacity: 0.4; cursor: not-allowed; transform: none !important; }
</style>
