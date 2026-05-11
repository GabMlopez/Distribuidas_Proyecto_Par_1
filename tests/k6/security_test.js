import http from 'k6/http';
import { check } from 'k6';

export const options = {
    // 50 Usuarios Virtuales (VUs) atacando a la vez
    vus: 50,
    duration: '10s',
};

const API_IP = __ENV.API_IP || '192.168.100.72';
const API_PORT = __ENV.API_PORT || '8085';

export default function () {
    // Asignar una IP falsa distinta a cada VU para simular tráfico desde un rango de red entero
    const spoofedIp = `192.168.200.${__VU}`;
    
    const params = {
        headers: {
            'X-Forwarded-For': spoofedIp,
        },
    };

    const url = `http://${API_IP}:${API_PORT}/rooms/list`;

    const res = http.get(url, params);

    // Como cada VU está simulando ser una IP distinta (ej. .1, .2, .3, ..., .50)
    // todas van a lograr hacer algunas peticiones (Status 200) y luego serán bloqueadas (Status 429) de forma independiente.
    check(res, {
        'status is 200 (OK) o 429 (Rate Limited)': (r) => r.status === 200 || r.status === 429,
        'bloqueado exitosamente (HTTP 429)': (r) => r.status === 429,
    });
}
