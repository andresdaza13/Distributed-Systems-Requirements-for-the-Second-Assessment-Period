// pagos.js - pagos con token idempotente, temporizador y reintentos (opciones c y e)

function escribirEcoPago(texto, clase) {
    const d = document.getElementById('eco-pagos');
    const p = document.createElement('div');
    p.className = clase || '';
    p.textContent = `[${new Date().toLocaleTimeString()}] ${texto}`;
    d.appendChild(p);
    d.scrollTop = d.scrollHeight;
}

async function cargarPlanes() {
    try {
        const respuesta = await fetch(`${URL_GATEWAY}/suscripciones/planes`);
        const planes = await respuesta.json();
        const contenedor = document.getElementById('lista-planes');
        contenedor.innerHTML = '';
        for (const plan of planes) {
            const boton = document.createElement('button');
            boton.textContent = `${plan.nombre} ($${plan.precio})`;
            boton.onclick = () => pagarPlan(plan);
            contenedor.appendChild(boton);
        }
    } catch (err) {
        document.getElementById('lista-planes').textContent = 'Servicio de suscripciones no disponible.';
    }
}

function pagarPlan(plan) {
    // UN token por intención de compra (no por reintento)
    const token = crypto.randomUUID();
    console.log('ECO cliente: usando token', token);
    escribirEcoPago(`ECO: nueva compra del plan "${plan.id}" con token ${token}`);
    intentarPago(token, { usuario: window.usuarioActual || 'anonimo', plan: plan.id, monto: plan.precio }, 1);
}

function intentarPago(token, datos, intento) {
    escribirEcoPago(`ECO: intento #${intento} con token ${token}`);
    const control = new AbortController();
    const temporizador = setTimeout(() => control.abort(), 4000); // timeout de 4 s

    fetch(`${URL_GATEWAY}/suscripciones/pagos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Idempotency-Key': token },
        body: JSON.stringify(datos),
        signal: control.signal
    })
    .then(r => r.json())
    .then(resultado => {
        clearTimeout(temporizador);
        window.ultimoToken = token;
        window.ultimoPago = datos;
        document.getElementById('btn-simular-token').disabled = false;
        console.log('ECO cliente:', resultado);
        escribirEcoPago(`ECO servidor: estado=${resultado.estado}, almacenado=${resultado.almacenado}, repetido=${resultado.repetido}`,
            resultado.repetido ? 'error' : 'success');
        if (resultado.repetido) {
            escribirEcoPago(`→ ${resultado.mensaje}`, 'error');
        }
    })
    .catch(() => {
        clearTimeout(temporizador);
        if (intento < 4) {
            const espera = 2000 * intento; // backoff creciente
            escribirEcoPago(`Sin respuesta. Reintento en ${espera / 1000} s con el MISMO token...`, 'error');
            setTimeout(() => intentarPago(token, datos, intento + 1), espera);
        } else {
            escribirEcoPago('No se pudo completar el pago tras 4 intentos.', 'error');
        }
    });
}

// Opción e: reenviar un pago con un token YA utilizado
function simularTokenUsado() {
    if (!window.ultimoToken) return;
    escribirEcoPago(`ECO: simulando reuso del token ${window.ultimoToken}`, 'error');
    intentarPago(window.ultimoToken, window.ultimoPago, 1);
}

cargarPlanes();
