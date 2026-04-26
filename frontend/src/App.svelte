<script>
  import Login from './lib/Login.svelte';
  import Home from './lib/Home.svelte';
  import Room from './lib/Room.svelte';

  let currentView = 'login'; // 'login', 'home', 'room'
  let userContext = {
    token: null,
    nickname: '',
    deviceId: 'dev_' + Math.random().toString(36).substr(2, 9)
  };
  let currentRoom = null;

  function handleLogin(event) {
    userContext.token = event.detail.token;
    userContext.nickname = event.detail.nickname || 'Admin';
    currentView = 'home';
  }

  function handleJoinRoom(event) {
    currentRoom = {
      id: event.detail.id,
      nombre: event.detail.nombre,
      type: event.detail.type,
      token: event.detail.token,
      usuarioId: event.detail.usuario_id 
    };
    currentView = 'room';
  }

  function handleLeaveRoom() {
    currentRoom = null;
    currentView = 'home';
  }
</script>

<!-- Room usa toda la pantalla, Login y Home van centrados -->
{#if currentView === 'room'}
  <Room {userContext} room={currentRoom} on:leave={handleLeaveRoom} />
{:else}
  <div class="centered-layout">
    {#if currentView === 'login'}
      <Login on:login={handleLogin} />
    {:else if currentView === 'home'}
      <Home {userContext} on:join={handleJoinRoom} />
    {/if}
  </div>
{/if}

<style>
  .centered-layout {
    min-height: 100vh;
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }
</style>
