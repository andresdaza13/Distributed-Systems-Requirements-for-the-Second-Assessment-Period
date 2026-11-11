// reacciones.js - WebSocket + pub/sub por canal (opción a)
// CORREGIDO: se recuerda el canal suscrito (canalSuscrito) por separado del
// audio que se va a reproducir (audioActual), para desuscribirse del canal
// CORRECTO al cambiar de canción.

const URL_WS = 'ws://localhost:8000/reacciones/ws';
let ws = null;
let canalSuscrito = null;   // canal en el que estamos AHORA en el servidor

function escribirReaccion(texto, clase) {
    const d = document.getElementById('log-reacciones');
    if (!d) return;
    const p = document.createElement('div');
    p.className = clase || '';
    p.textContent = `[${new Date().toLocaleTimeString()}] ${texto}`;
    d.appendChild(p);
    d.scrollTop = d.scrollHeight;
}

function conectarReacciones() {
    ws = new WebSocket(URL_WS);

    ws.onopen = () => {
        document.getElementById('estado-ws').textContent = 'conectado';
        escribirReaccion('Conectado al servidor de reacciones.', 'success');
        canalSuscrito = null; // conexión nueva: el servidor no nos tiene en ningún canal
    };

    ws.onmessage = (evento) => {
        const m = JSON.parse(evento.data);
        // Defensa en el cliente: solo mostramos reacciones del canal suscrito.
        if (m.audioId !== canalSuscrito) return;
        escribirReaccion(`${m.usuario} reaccionó ${m.emoji} en "${m.audioId}"`);
    };

    ws.onclose = () => {
        document.getElementById('estado-ws').textContent = 'desconectado (reintentando...)';
        canalSuscrito = null;
        setTimeout(conectarReacciones, 3000);
    };
    ws.onerror = () => ws.close();
}

function enviarWS(mensaje) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(mensaje));
    }
}

// Entrar a un canal saliendo SIEMPRE antes del anterior.
function suscribirA(audioId) {
    if (canalSuscrito === audioId) return;        // ya estamos ahí
    if (canalSuscrito) desuscribirDe(canalSuscrito); // salir del canal VIEJO
    enviarWS({ accion: 'suscribir', audioId: audioId, usuario: window.usuarioActual });
    canalSuscrito = audioId;
    document.getElementById('canal-actual').textContent = audioId;
    escribirReaccion(`Suscrito al canal "${audioId}".`, 'success');
}

function desuscribirDe(audioId) {
    enviarWS({ accion: 'desuscribir', audioId: audioId, usuario: window.usuarioActual });
    if (canalSuscrito === audioId) canalSuscrito = null;
    document.getElementById('canal-actual').textContent = 'ninguno';
    escribirReaccion(`Desuscrito del canal "${audioId}".`, 'error');
}

function enviarReaccion(emoji) {
    if (!canalSuscrito) {
        escribirReaccion('Reproduce un audio para unirte a su canal.', 'error');
        return;
    }
    enviarWS({
        accion: 'reaccion',
        audioId: canalSuscrito,          // se reacciona al canal suscrito, no al "próximo"
        usuario: window.usuarioActual || 'anonimo',
        emoji: emoji
    });
}

(function () {
    function engancharListeners() {
        const audio = document.getElementById('audio-player');
        if (!audio) return;

        // play: suscribirse al audio que REALMENTE está sonando
        audio.addEventListener('play', () => {
            if (!window.audioActual) return;
            suscribirA(window.audioActual);
        });

        // pausa: salir del canal al que estamos suscritos (no del "próximo")
        audio.addEventListener('pause', () => {
            if (!canalSuscrito) return;
            desuscribirDe(canalSuscrito);
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => { engancharListeners(); conectarReacciones(); });
    } else {
        engancharListeners();
        conectarReacciones();
    }
})();