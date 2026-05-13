import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '5s', target: 50 },  // Rampa de subida a 50 usuarios
        { duration: '20s', target: 50 }, // Mantener 50 usuarios
        { duration: '5s', target: 0 },   // Rampa de bajada
    ],
};

const API_IP = __ENV.API_IP || '192.168.100.72';
const API_PORT = __ENV.API_PORT || '8085';

export default function () {
    const nickname = `K6_VU_${__VU}_${__ITER}`;
    const deviceId = `device_k6_${__VU}_${__ITER}`;
    const roomId = 'TEST_ROOM_K6';

    const url = `ws://${API_IP}:${API_PORT}/ws/${roomId}?nickname=${nickname}&device_id=${deviceId}&sala_id=${roomId}`;

    const res = ws.connect(url, null, function (socket) {
        socket.on('open', () => {
            // Enviar 3 mensajes espaciados
            for (let i = 0; i < 3; i++) {
                socket.send(JSON.stringify({
                    tipo: 'chat',
                    texto: `Mensaje de prueba ${i} desde VU ${__VU}`,
                    sala_id: roomId,
                    nickname: nickname,
                    timestamp: Date.now()
                }));
                socket.setInterval(function timeout() {
                    // Esperar 2 segundos entre mensajes
                }, 2000);
            }
        });

        socket.on('message', (msg) => {
            // Verificar que estamos recibiendo mensajes (del broadcast)
            check(msg, { 'mensaje recibido ok': (m) => m && m.length > 0 });
        });

        socket.on('close', () => {
            // Se cierra la conexión
        });

        // Cerrar socket después de 10 segundos
        socket.setTimeout(function () {
            socket.close();
        }, 10000);
    });

    if (res && res.status !== 101) {
        console.log(`Error de conexión (Status ${res.status}): ${res.body}`);
    }

    check(res, { 'status is 101': (r) => r && r.status === 101 });
    sleep(1);
}
