// catalogo.js - consume el servidor de audios y metadatos (puerto 5001)

const URL_AUDIOS = 'http://localhost:5001';

// Estado: el audio actualmente seleccionado (lo usa el botón Reproducir)
let audioSeleccionado = null;

// ---------- 1. Cargar los tipos al abrir la página ----------
async function cargarTipos() {
    try {
        const respuesta = await fetch(`${URL_AUDIOS}/tipos`);
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
        console.error('No se pudo contactar al servidor de audios:', err);
    }
}

// ---------- 2. Cargar la lista de audios de un tipo ----------
async function cargarAudios(tipo) {
    const respuesta = await fetch(`${URL_AUDIOS}/audios?tipo=${tipo}`);
    if (!respuesta.ok) {
        console.error('Error del servidor:', respuesta.status);
        return;
    }
    const audios = await respuesta.json();

    document.getElementById('titulo-lista').style.display = '';
    const lista = document.getElementById('lista-audios');
    lista.innerHTML = '';
    document.getElementById('panel-detalle').style.display = 'none';

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

// ---------- 3. Ver el detalle de un audio ----------
async function verDetalle(id) {
    const respuesta = await fetch(`${URL_AUDIOS}/audios/${id}`);
    if (!respuesta.ok) {
        console.error('Error del servidor:', respuesta.status);
        return;
    }
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

// ---------- 4. Reproducir: puente con el streaming ----------
document.getElementById('btn-reproducir').onclick = () => {
    if (!audioSeleccionado) return;
    document.getElementById('song-title').value = audioSeleccionado.archivo;
    document.getElementById('format-select').value = audioSeleccionado.formato;
    pedirCancion(audioSeleccionado.archivo, audioSeleccionado.formato);
};

// Arranque
cargarTipos();