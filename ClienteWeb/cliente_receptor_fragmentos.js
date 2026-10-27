// cliente_receptor_fragmentos.js
// Recibe por gRPC-Web los fragmentos del audio. La petición sale hacia el API Gateway,
// que la reenvía al Proxy Envoy (HTTP/1.1 -> HTTP/2) y este al Servidor de Streaming.
// Si modifica este archivo ejecute "npx webpack" para regenerar bundle.js.

const { peticionDTO } = require('./stubs_generados/servicios_pb.js');
const { AudioServicePromiseClient } = require('./stubs_generados/servicios_grpc_web_pb.js');

// La URL se toma de js/config.js (cargado antes que bundle.js).
const URL_STREAMING = (typeof window !== 'undefined' && window.CONFIG && window.CONFIG.URL_STREAMING)
    || 'http://localhost:8080';

const client = new AudioServicePromiseClient(URL_STREAMING);

// Fragmentos de audio recibidos
let audioChunks = [];
let _firstFragmentReceived = false;
let streamActual = null;

const audioPlayer = document.getElementById('audio-player');

function escribirLog(mensaje, nivel) {
    if (typeof window !== 'undefined' && window.__writeAudioLog) {
        window.__writeAudioLog(mensaje, nivel);
    } else {
        console.log(mensaje);
    }
}

function iniciar_streaming_cancion(titulo, formato) {

    console.log(`Iniciando llamada gRPC-Web a través del API Gateway (${URL_STREAMING})...`);

    // Si había una descarga en curso, se cancela y se descartan sus fragmentos
    if (streamActual) {
        streamActual.cancel();
    }
    audioChunks = [];
    _firstFragmentReceived = false;

    const request = new peticionDTO();
    request.setTitulo(titulo);
    request.setFormato(formato);

    console.log("Solicitud preparada:", request.toObject());
    const stream = client.enviarCancionMedianteStream(request, {});
    streamActual = stream;

    // Maneja la llegada de cada fragmento
    stream.on('data', (response) => {
        const data = response.getData_asU8();
        console.log(`Fragmento recibido desde el servidor. Tamaño: ${data.length} bytes`);
        audioChunks.push(data);

        // Si es el primer fragmento, se muestra el reproductor
        if (!_firstFragmentReceived) {
            _firstFragmentReceived = true;
            if (audioPlayer) {
                audioPlayer.style.display = '';
            }
            escribirLog('Primer fragmento recibido. Reproductor mostrado.', 'success');
        }
    });

    // Maneja el final del stream
    stream.on('end', () => {
        if (stream !== streamActual) {
            return; // descarga cancelada por una nueva petición
        }
        streamActual = null;
        console.log('✅ Transmisión de la canción finalizada. Preparando reproducción...');

        if (audioChunks.length === 0) {
            escribirLog(`El servidor no envió fragmentos para "${titulo}.${formato}".`, 'error');
            return;
        }

        const tipoMime = formato === 'wav' ? 'audio/wav' : formato === 'm4a' ? 'audio/mp4' : 'audio/mpeg';
        const audioBlob = new Blob(audioChunks, { type: tipoMime });
        const audioUrl = URL.createObjectURL(audioBlob);
        escribirLog(`Audio recibido completo: ${audioChunks.length} fragmentos.`, 'success');

        if (audioPlayer) {
            audioPlayer.src = audioUrl;
            audioPlayer.play()
                .then(() => console.log('▶️ Reproducción de la canción iniciada con éxito.'))
                .catch(error => {
                    console.error('❌ Error al iniciar la reproducción automática. Dé clic en el reproductor.', error);
                    escribirLog('Dé clic en ▶ del reproductor para escuchar el audio.', 'error');
                });
        }

        audioChunks = [];
        _firstFragmentReceived = false;
    });

    stream.on('error', (err) => {
        if (stream !== streamActual) {
            return;
        }
        streamActual = null;
        console.error('ERROR en la comunicación gRPC-Web:', err);
        escribirLog(`Error en el streaming: ${err.message || err.code}`, 'error');
    });
}

// Se expone la implementación para funciones.js
if (typeof window !== 'undefined') {
    window.iniciar_streaming_cancion = iniciar_streaming_cancion;
    window.iniciarStreamGRPCImpl = iniciar_streaming_cancion;
    console.log('Exported iniciar_streaming_cancion to window');
}
