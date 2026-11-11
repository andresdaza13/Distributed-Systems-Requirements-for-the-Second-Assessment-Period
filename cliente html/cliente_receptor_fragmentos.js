// cliente_receptor_fragmentos.js
// MODIFICADO: adjunta el metadato x-usuario a la llamada gRPC-Web para que el
// servidor sepa quién reproduce (opción d) SIN cambiar el contrato .proto.
// Tras editar este archivo hay que ejecutar: npx webpack

const { peticionDTO } = require('./stubs_generados/servicios_pb.js');
const { AudioServicePromiseClient } = require('./stubs_generados/servicios_grpc_web_pb.js');

const ENVOY_HOST = 'http://localhost:8080';
const client = new AudioServicePromiseClient(ENVOY_HOST);

let audioChunks = [];
const audioPlayer = document.getElementById('audio-player');
let _firstFragmentReceived = false;

function iniciar_streaming_cancion(titulo, formato) {
    console.log("Iniciando llamada gRPC-Web a través de Envoy...");

    const request = new peticionDTO();
    request.setTitulo(titulo);
    request.setFormato(formato);

    // Metadatos de la llamada (viajan como cabeceras HTTP; Envoy debe permitirlos en CORS)
    const metadata = { 'x-usuario': (window.usuarioActual || 'anonimo') };

    console.log("Solicitud preparada:", request.toObject(), "usuario:", metadata['x-usuario']);
    const stream = client.enviarCancionMedianteStream(request, metadata);

    stream.on('data', (response) => {
        const data = response.getData_asU8();
        console.log(`Fragmento recibido desde el servidor. Tamaño: ${data.length} bytes`);
        audioChunks.push(data);

        if (!_firstFragmentReceived) {
            _firstFragmentReceived = true;
            if (audioPlayer) {
                audioPlayer.style.display = '';
            }
            if (typeof window !== 'undefined' && window.__writeAudioLog) {
                window.__writeAudioLog('Primer fragmento recibido. Reproductor mostrado.', 'success');
            }
        }
    });

    stream.on('end', () => {
        console.log('✅ Transmisión finalizada. Preparando reproducción...');
        const audioBlob = new Blob(audioChunks, { type: 'audio/mp3' });
        const audioUrl = URL.createObjectURL(audioBlob);
        if (audioPlayer) {
            audioPlayer.src = audioUrl;
            audioPlayer.play()
                .then(() => console.log('▶️ Reproducción iniciada.'))
                .catch(error => console.error('❌ Dé clic en el reproductor para iniciar.', error));
        }
        audioChunks = [];
        _firstFragmentReceived = false;
    });

    stream.on('error', (err) => {
        console.error('ERROR en la comunicación gRPC-Web:', err);
    });
}

if (typeof window !== 'undefined') {
    window.iniciar_streaming_cancion = iniciar_streaming_cancion;
    window.iniciarStreamGRPCImpl = iniciar_streaming_cancion;
    console.log('Exported iniciar_streaming_cancion to window');
}
