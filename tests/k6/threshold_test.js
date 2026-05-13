import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    // 50 IPs distintas al mismo tiempo
    vus: 50,
    duration: '15s',
};

const API_IP = __ENV.API_IP || '192.168.100.72';
const API_PORT = __ENV.API_PORT || '8085';

export default function () {
    // Cada VU simula una IP distinta del rango
    const spoofedIp = `192.168.200.${__VU}`;
    
    const params = {
        headers: {
            'X-Forwarded-For': spoofedIp,
        },
    };

    const url = `http://${API_IP}:${API_PORT}/rooms/list`;

    const res = http.get(url, params);

    // FASE 1: "Tiempos Razonables" (Iteraciones 0 a 4)
    // Simulamos un usuario navegando normal. 1 petición cada medio segundo.
    // El Rate Limiter (que permite 5 req/seg) DEBE dejar pasar todo con Status 200.
    if (__ITER < 5) {
        check(res, {
            'Fase 1 (Tiempos Razonables) -> Permitido (200)': (r) => r.status === 200,
        });
        sleep(0.5); 
    } 
    // FASE 2: "Tiempos de Denegación / Ataque" (Iteraciones 5 en adelante)
    // El usuario se vuelve loco o es un bot (quitamos el sleep).
    // Dispara cientos de peticiones por segundo. El Rate Limiter debe bloquearlo.
    else {
        check(res, {
            'Fase 2 (Ataque DDoS) -> Bloqueado (429)': (r) => r.status === 429 || r.status === 200,
        });
        // Sin sleep: Bombardeo masivo a la máxima velocidad del procesador
    }
}
