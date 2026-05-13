/**
 * setup_local.js
 * Detecta automáticamente todas las IPs locales del equipo y configura:
 * 1. frontend-next/.env  → NEXT_PUBLIC_API_URL con la IP Wi-Fi/Ethernet
 * 2. frontend-next/next.config.ts → allowedDevOrigins con TODAS las IPs
 * 
 * Ejecutar: node setup_local.js
 */

const os = require('os');
const fs = require('fs');
const path = require('path');

// ── 1. Detectar todas las IPs IPv4 no-internas ──
const nets = os.networkInterfaces();
const allIPs = ['10.9.8.25']; // Aseguramos incluir siempre la que pide Next.js
let primaryIP = 'localhost';

for (const name of Object.keys(nets)) {
  for (const net of nets[name]) {
    if (net.family === 'IPv4' && !net.internal) {
      // Solo agregar IPs de la red local real, ignorar adaptadores virtuales (WSL, Docker, etc)
      if (net.address.startsWith('192.168.100.')) {
        allIPs.push(net.address);
      }
      
      // Ignorar IPs de VirtualBox (192.168.56.x) y WSL/Hyper-V usuales
      const isVirtual = name.toLowerCase().includes('virtual') || name.toLowerCase().includes('wsl') || name.toLowerCase().includes('hyper') || net.address.startsWith('192.168.56.');
      
      // Preferir Wi-Fi o Ethernet real (192.168.x.x o 10.x.x.x)
      if (
        !isVirtual &&
        (net.address.startsWith('192.168.') || net.address.startsWith('10.')) &&
        primaryIP === 'localhost'
      ) {
        primaryIP = net.address;
      }
    }
  }
}

// Si no encontramos una IP 192.168/10.x, usar la primera disponible
if (primaryIP === 'localhost' && allIPs.length > 0) {
  primaryIP = allIPs[0];
}

console.log('');
console.log('╔══════════════════════════════════════════════════╗');
console.log('║       CONFIGURACIÓN DE RED LOCAL DETECTADA       ║');
console.log('╠══════════════════════════════════════════════════╣');
console.log(`║  IP Principal (Wi-Fi/LAN): ${primaryIP.padEnd(22)}║`);
console.log('║                                                  ║');
console.log('║  Todas las IPs detectadas:                       ║');
allIPs.forEach(ip => {
  console.log(`║    • ${ip.padEnd(44)}║`);
});
console.log('║                                                  ║');
console.log(`║  Frontend: http://${primaryIP}:3000`.padEnd(52) + '║');
console.log(`║  Backend:  http://${primaryIP}:8085`.padEnd(52) + '║');
console.log('╚══════════════════════════════════════════════════╝');
console.log('');

// ── 2. Actualizar frontend-next/.env ──
const envPath = path.join(__dirname, 'frontend-next', '.env');
const envContent = `NEXT_PUBLIC_API_URL=http://${primaryIP}:8085\n`;
fs.writeFileSync(envPath, envContent, 'utf8');
console.log(`✅ .env actualizado → NEXT_PUBLIC_API_URL=http://${primaryIP}:8085`);

// ── 3. Actualizar frontend-next/next.config.mjs ──
const nextConfigPath = path.join(__dirname, 'frontend-next', 'next.config.mjs');
const originsArray = allIPs.map(ip => `'${ip}'`).join(', ');
const nextConfigContent = `/** @type {import('next').NextConfig} */
const nextConfig = {
  allowedDevOrigins: [${originsArray}],
};

export default nextConfig;
`;
fs.writeFileSync(nextConfigPath, nextConfigContent, 'utf8');
console.log(`✅ next.config.mjs actualizado → allowedDevOrigins: [${originsArray}]`);
console.log('');
console.log('🚀 Configuración lista. Iniciando servicios...');
console.log('');
