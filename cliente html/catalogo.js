// catalogo.js - consume el catálogo a través del API Gateway

const URL_GATEWAY = 'http://localhost:8000';

let audioSeleccionado = null;
window.audioActual = null;    // canal de reacciones del audio en reproducción
window.usuarioActual = 'andres';

document.getElementById('usuario').addEventListener('input', (e) => {
    window.usuarioActual = e.target.value.trim() || 'anonimo';
});

function usuario() {
    return window.usuarioActual || 'anonimo';
}

// ---------- 1. Tipos ----------
async function cargarTipos() {
    try {
        const respuesta = await fetch(`${URL_GATEWAY}/audios/tipos`);
        const tipos = await respuesta.json();
        const contenedor = document.getElementById('lista-tipos');
        contenedor.innerHTML = '';
        for (const tipo of tipos) {
            const boton = document.createElement('button');
            boton.textContent = tipo;
            boton.onclick = () => cargarAudios(tipo);
            contenedor.appendChild(boton);
        }
    } catch (err) {
        console.error('No se pudo contactar al catálogo (¿gateway y servicio de audios arriba?):', err);
    }
}

// ---------- 2. Lista por tipo ----------
async function cargarAudios(tipo) {
    const respuesta = await fetch(`${URL_GATEWAY}/audios/audios?tipo=${tipo}`);
    if (!respuesta.ok) { console.error('Error del servidor:', respuesta.status); return; }
    const audios = await respuesta.json();

    document.getElementById('titulo-lista').style.display = '';
    const lista = document.getElementById('lista-audios');
    lista.innerHTML = '';
    document.getElementById('panel-detalle').style.display = 'none';

    if (audios.length === 0) {
        lista.innerHTML = '<li style="color:#666">No hay audios de este tipo todavía.</li>';
        return;
    }
    for (const audio of audios) {
        const item = document.createElement('li');
        const enlace = document.createElement('a');
        enlace.href = '#';
        enlace.textContent = audio.titulo;
        enlace.onclick = (e) => { e.preventDefault(); verDetalle(audio.id); };
        item.appendChild(enlace);
        lista.appendChild(item);
    }
}

// ---------- 3. Detalle ----------
async function verDetalle(id) {
    const respuesta = await fetch(`${URL_GATEWAY}/audios/audios/${id}`);
    if (!respuesta.ok) { console.error('Error del servidor:', respuesta.status); return; }
    audioSeleccionado = await respuesta.json();

    document.getElementById('detalle-titulo').textContent =
        `${audioSeleccionado.titulo} (${audioSeleccionado.tipo})`;

    const tabla = document.getElementById('detalle-metadatos');
    tabla.innerHTML = '';
    for (const [clave, valor] of Object.entries(audioSeleccionado.metadatos)) {
        const fila = tabla.insertRow();
        fila.insertCell().textContent = clave;
        fila.insertCell().textContent = valor;
    }
    document.getElementById('panel-detalle').style.display = '';
}

// ---------- 4. Puente con el streaming ----------
/*
document.getElementById('btn-reproducir').onclick = () => {
    if (!audioSeleccionado) return;
    window.audioActual = audioSeleccionado.archivo;   // canal de reacciones
    document.getElementById('song-title').value = audioSeleccionado.archivo;
    document.getElementById('format-select').value = audioSeleccionado.formato;
    pedirCancion(audioSeleccionado.archivo, audioSeleccionado.formato);
};
*/

document.getElementById('btn-reproducir').onclick = () => {
    if (!audioSeleccionado) return;
    window.audioActual = audioSeleccionado.id;   // ← ANTES: audioSeleccionado.archivo
    document.getElementById('song-title').value = audioSeleccionado.archivo;
    document.getElementById('format-select').value = audioSeleccionado.formato;
    pedirCancion(audioSeleccionado.archivo, audioSeleccionado.formato);
};
// ---------- Estadísticas (opción d) ----------
async function cargarHistorial() {
    const respuesta = await fetch(`${URL_GATEWAY}/estadisticas/usuarios/${usuario()}`);
    if (!respuesta.ok) { console.error('Error consultando estadísticas'); return; }
    const historial = await respuesta.json();
    const tabla = document.getElementById('tabla-historial');
    tabla.innerHTML = '<tr><th>Audio</th><th>Veces reproducido</th></tr>';
    if (historial.length === 0) {
        const fila = tabla.insertRow();
        const celda = fila.insertCell();
        celda.colSpan = 2;
        celda.textContent = 'Sin reproducciones registradas todavía.';
        return;
    }
    for (const item of historial) {
        const fila = tabla.insertRow();
        fila.insertCell().textContent = item.audioId;
        fila.insertCell().textContent = item.veces;
    }
}

cargarTipos();
