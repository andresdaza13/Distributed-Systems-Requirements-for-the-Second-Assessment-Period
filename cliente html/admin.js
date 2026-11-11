// admin.js - registro de audios con metadatos polimórficos (opción f)

const URL_GATEWAY_ADMIN = 'http://localhost:8000';

// Al cambiar el tipo, el formulario se redibuja con los campos de ESE tipo.
async function cargarTiposAdmin() {
    const respuesta = await fetch(`${URL_GATEWAY_ADMIN}/audios/tipos`);
    const tipos = await respuesta.json();
    const select = document.getElementById('admin-tipo');
    select.innerHTML = '';
    for (const tipo of tipos) {
        const opcion = document.createElement('option');
        opcion.value = tipo;
        opcion.textContent = tipo;
        select.appendChild(opcion);
    }
    select.onchange = () => dibujarCampos(select.value);
    dibujarCampos(tipos[0]);
}

async function dibujarCampos(tipo) {
    const respuesta = await fetch(`${URL_GATEWAY_ADMIN}/audios/tipos/${tipo}/campos`);
    const campos = await respuesta.json();
    const contenedor = document.getElementById('campos-dinamicos');
    contenedor.innerHTML = '';
    for (const campo of campos) {
        const etiqueta = document.createElement('label');
        etiqueta.textContent = campo.replaceAll('_', ' ');
        etiqueta.htmlFor = `campo-${campo}`;
        const entrada = document.createElement('input');
        entrada.type = 'text';
        entrada.id = `campo-${campo}`;
        entrada.dataset.campo = campo;
        contenedor.appendChild(etiqueta);
        contenedor.appendChild(entrada);
    }
}

async function registrarAudio() {
    const resultado = document.getElementById('resultado');
    const tipo = document.getElementById('admin-tipo').value;
    const titulo = document.getElementById('admin-titulo').value.trim();
    const entradaArchivo = document.getElementById('admin-archivo');

    if (!entradaArchivo.files[0]) {
        resultado.textContent = 'Selecciona un archivo de audio.';
        resultado.className = 'error';
        return;
    }

    const metadatos = {};
    for (const entrada of document.querySelectorAll('#campos-dinamicos input')) {
        metadatos[entrada.dataset.campo] = entrada.value.trim();
    }

    // multipart/form-data: texto + binario en UNA sola petición
    const datos = new FormData();
    datos.append('titulo', titulo);
    datos.append('tipo', tipo);
    datos.append('metadatos', JSON.stringify(metadatos));
    datos.append('archivo', entradaArchivo.files[0]);

    const respuesta = await fetch(`${URL_GATEWAY_ADMIN}/audios/audios`, {
        method: 'POST',
        body: datos
    });
    const cuerpo = await respuesta.json();
    if (respuesta.ok) {
        resultado.className = 'success';
        resultado.textContent = 'Audio registrado:\n' + JSON.stringify(cuerpo, null, 2);
    } else {
        resultado.className = 'error';
        resultado.textContent = `Error ${respuesta.status}: ${cuerpo.error}`;
    }
}

cargarTiposAdmin();
