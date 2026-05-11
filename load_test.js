import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';

// Configuración de múltiples escenarios de prueba de carga
export const options = {
    scenarios: {
        // Escenario 1: Prueba de Carga Progresiva en WebSockets (Ramping VUs)
        websocket_ramp: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '5s', target: 20 },  // Sube a 20 usuarios en 5 seg
                { duration: '15s', target: 50 }, // Pico de 50 usuarios requeridos por el PDF
                { duration: '10s', target: 100 }, // Prueba superior a lo pedido: 100 usuarios en 10 seg
                { duration: '5s', target: 0 },   // Baja a 0 suavemente
            ],
            exec: 'ws_test',
        },
        // Escenario 2: Prueba de Estrés constante en API REST (Peticiones simultáneas)
        api_stress: {
            executor: 'constant-vus',
            vus: 30, // 30 usuarios simultáneos 
            duration: '20s',
            exec: 'api_test',
        }
    }
};

// Configuración dinámica de URL
const BASE_URL = __ENV.BASE_URL || '127.0.0.1:8085';
const WS_URL = BASE_URL.replace('http://', 'ws://').replace('https://', 'wss://');

// Prueba 1: WebSockets
export function ws_test() {
    // Generar device_id único para no chocar con restricción de sesión única
    const deviceId = `k6-ws-${__VU}-${__ITER}`;
    const url = `${WS_URL}/ws/test-room-1?nickname=Usuario_${__VU}&device_id=${deviceId}`;

    const res = ws.connect(url, null, function (socket) {
        socket.on('open', () => {
            // Emite un mensaje cada 2 segundos
            socket.setInterval(function timeout() {
                socket.send(JSON.stringify({
                    tipo: "chat",
                    texto: `Hola desde k6 - Usuario ${__VU}`,
                    sala_id: "sala-demo-k6",
                    nickname: `Usuario_${__VU}`
                }));
            }, 2000); 
        });

        // Simular tiempo de vida
        socket.setTimeout(function () {
            socket.close();
        }, 15000);
    });

    check(res, { 'Conexión WS lograda (101)': (r) => r && r.status === 101 });
}

// Prueba 2: Endpoints REST HTTP
export function api_test() {
    let res = http.post(`http://${BASE_URL}/upload`, {});
    
    check(res, {
        'API HTTP responde rapido': (r) => r.timings.duration < 500, // Menos de 500ms
        'API no cae (No 500)': (r) => r.status !== 500,
    });
    sleep(1);
}
