<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  let mode = 'user';
  let username = '';
  let password = '';
  let nickname = '';
  let error = '';
  let loading = false;

  async function handleLogin() {
    error = '';
    loading = true;

    if (mode === 'admin') {
      try {
        const res = await fetch('http://localhost:8080/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ usuario: username, contrasenia: password })
        });
        const data = await res.json();
        if (res.ok) {
          dispatch('login', { token: data.token, nickname: 'Admin' });
        } else {
          error = data.error || 'Credenciales incorrectas';
        }
      } catch {
        error = 'No se pudo conectar al servidor';
      }
    } else {
      if (!nickname.trim()) {
        error = 'El nickname es requerido';
      } else {
        dispatch('login', { token: null, nickname: nickname.trim() });
      }
    }
    loading = false;
  }
</script>

<div class="page">
  <!-- Floating orbs -->
  <div class="orb orb-1"></div>
  <div class="orb orb-2"></div>
  <div class="orb orb-3"></div>

  <div class="card anim-fade-up">
    <!-- Logo area -->
    <div class="brand">
      <div class="logo-ring">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        </svg>
      </div>
      <div>
        <h1 class="title gradient-text">Chat Distribuido</h1>
        <p class="subtitle">Plataforma de mensajería en tiempo real</p>
      </div>
    </div>

    <div class="divider"></div>

    <!-- Tabs -->
    <div class="tabs">
      <button
        class="tab {mode === 'user' ? 'active' : ''}"
        on:click={() => { mode = 'user'; error = ''; }}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M20 21a8 8 0 1 0-16 0"/></svg>
        Invitado
      </button>
      <button
        class="tab {mode === 'admin' ? 'active' : ''}"
        on:click={() => { mode = 'admin'; error = ''; }}>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
        Admin
      </button>
    </div>

    <!-- Form -->
    <form on:submit|preventDefault={handleLogin} class="form">
      {#if mode === 'admin'}
        <div class="form-field">
          <label for="usr">Usuario</label>
          <input id="usr" type="text" bind:value={username} placeholder="admin" required />
        </div>
        <div class="form-field">
          <label for="pwd">Contraseña</label>
          <input id="pwd" type="password" bind:value={password} placeholder="••••••••" required />
        </div>
      {:else}
        <div class="form-field">
          <label for="nick">Tu nickname</label>
          <input id="nick" type="text" bind:value={nickname} placeholder="Ej: CoolUser99" required maxlength="20" />
        </div>
        <p class="hint">No necesitas contraseña para unirte como invitado</p>
      {/if}

      {#if error}
        <div class="error-banner">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          {error}
        </div>
      {/if}

      <button type="submit" class="btn-primary submit-btn" disabled={loading}>
        {#if loading}
          <span class="spinner"></span> Autenticando...
        {:else}
          Entrar al chat
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
        {/if}
      </button>
    </form>
  </div>
</div>

<style>
  .page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.5rem;
    position: relative;
  }

  /* Orbs */
  .orb {
    position: fixed;
    border-radius: 50%;
    filter: blur(80px);
    pointer-events: none;
    animation: float 8s ease-in-out infinite;
  }
  .orb-1 {
    width: 400px; height: 400px;
    background: rgba(99, 102, 241, 0.25);
    top: -10%; left: -5%;
    animation-delay: 0s;
  }
  .orb-2 {
    width: 350px; height: 350px;
    background: rgba(236, 72, 153, 0.2);
    bottom: -10%; right: -5%;
    animation-delay: -3s;
  }
  .orb-3 {
    width: 250px; height: 250px;
    background: rgba(34, 211, 238, 0.15);
    top: 50%; left: 60%;
    animation-delay: -5s;
  }
  @keyframes float {
    0%, 100% { transform: translate(0, 0)     scale(1); }
    33%       { transform: translate(15px, -20px) scale(1.03); }
    66%       { transform: translate(-10px, 10px) scale(0.97); }
  }

  .card {
    width: 100%; max-width: 400px;
    background: rgba(8, 12, 28, 0.75);
    backdrop-filter: saturate(180%) blur(28px);
    -webkit-backdrop-filter: saturate(180%) blur(28px);
    border: 1px solid rgba(255, 255, 255, 0.22);
    border-radius: 24px;
    box-shadow:
      0 0 0 1px rgba(255,255,255,0.06) inset,
      0 24px 80px rgba(0, 0, 0, 0.7),
      0 0 60px rgba(99, 102, 241, 0.12);
    padding: 2.25rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    position: relative;
  }

  /* Top highlight line */
  .card::before {
    content: '';
    position: absolute;
    top: 0; left: 8%; right: 8%;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(165,180,252,0.6), transparent);
    border-radius: 99px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .logo-ring {
    width: 52px; height: 52px;
    border-radius: 14px;
    background: linear-gradient(135deg, rgba(99,102,241,0.3), rgba(236,72,153,0.2));
    border: 1px solid rgba(255,255,255,0.12);
    display: flex; align-items: center; justify-content: center;
    color: var(--indigo-light);
    flex-shrink: 0;
    box-shadow: 0 0 20px rgba(99,102,241,0.25), inset 0 1px 0 rgba(255,255,255,0.1);
    animation: pulse-ring 3s ease-in-out infinite;
  }

  .title { font-size: 1.5rem; line-height: 1; margin-bottom: 0.25rem; }
  .subtitle { font-size: 0.78rem; color: var(--text-3); }

  /* Tabs */
  .tabs {
    display: flex;
    background: rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255,255,255,0.15);
    border-radius: 10px;
    padding: 3px;
    gap: 3px;
  }
  .tab {
    flex: 1;
    background: transparent;
    color: var(--text-3);
    padding: 0.5rem;
    border-radius: 7px;
    font-size: 0.82rem;
    font-weight: 500;
    transition: all 0.2s ease;
  }
  .tab:hover { color: var(--text-2); background: rgba(255,255,255,0.06); }
  .tab.active {
    background: rgba(99, 102, 241, 0.20);
    color: #c7d2fe;
    border: 1px solid rgba(165,180,252,0.35);
    text-shadow: 0 0 8px rgba(165,180,252,0.4);
  }
  .tab::after { display: none; }

  .form { display: flex; flex-direction: column; gap: 1rem; }

  .hint {
    font-size: 0.78rem;
    color: var(--text-3);
    line-height: 1.4;
  }

  .submit-btn {
    width: 100%;
    padding: 0.85rem;
    margin-top: 0.5rem;
    font-size: 0.95rem;
  }

  /* Spinner */
  .spinner {
    width: 14px; height: 14px;
    border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
    flex-shrink: 0;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
