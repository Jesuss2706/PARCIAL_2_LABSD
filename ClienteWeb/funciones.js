// funciones.js
// 1. Wrapper para pedir el audio por streaming (implementado en bundle.js).
// 2. Cliente del servidor de reacciones (WebSocket STOMP sobre SockJS):
//    - al reproducir (play) el cliente se REGISTRA en el canal del audio,
//    - al pausar (pause) o terminar el audio se DES-REGISTRA del canal,
//    - las reacciones de otros oyentes del mismo audio llegan por callback.

const TIPOS_REACCION = {
    like:  { icono: 'fa-thumbs-up',      clase: 'icon-like',  texto: 'Me gusta' },
    heart: { icono: 'fa-heart',          clase: 'icon-heart', texto: 'Me encanta' },
    sad:   { icono: 'fa-face-sad-tear',  clase: 'icon-sad',   texto: 'Me entristece' },
    fun:   { icono: 'fa-face-laugh',     clase: 'icon-fun',   texto: 'Me divierte' },
};

const estadoReacciones = {
    clienteStomp: null,
    conectado: false,
    suscripcion: null,   // suscripción STOMP al canal del audio que se está reproduciendo
    canalRegistrado: null,
    audioActual: null,   // { titulo, formato, canal }
    temporizadorReconexion: null,
};

// ------------------------------------------------------------------ Bitácora (ecos del cliente)

function writeLog(message, level) {
    const d = document.getElementById('log');
    if (!d) return;
    const p = document.createElement('div');
    p.className = level || '';
    p.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
    d.appendChild(p);
    d.scrollTop = d.scrollHeight;
}

function eco(mensaje, nivel) {
    console.log(`[ECO CLIENTE] ${mensaje}`);
    writeLog(mensaje, nivel);
}

// ------------------------------------------------------------------ Streaming

/** El nombre del canal se deriva del título del audio (sin espacios ni símbolos). */
function nombreCanal(titulo) {
    return titulo.trim().toLowerCase()
        .normalize('NFD').replace(/[̀-ͯ]/g, '')
        .replace(/[^a-z0-9_-]+/g, '_') || 'audio';
}

function pedirCancion(titulo, formato) {
    titulo = (titulo || '').trim();
    if (!titulo) {
        eco('Escriba el título del audio.', 'error');
        return;
    }

    // Si estaba registrado en el canal de otro audio, se des-registra antes de cambiar.
    desregistrarCanal('se pidió un nuevo audio');
    const audio = document.getElementById('audio-player');
    if (audio && !audio.paused) {
        audio.pause();
    }

    estadoReacciones.audioActual = { titulo, formato, canal: nombreCanal(titulo) };
    document.getElementById('status').textContent = `Descargando "${titulo}.${formato}"…`;
    document.getElementById('audio-actual').textContent = `${titulo}.${formato}`;
    actualizarPanelReacciones();
    eco(`Pidiendo el audio "${titulo}.${formato}" al servidor de streaming (gRPC-Web vía API Gateway)`, 'info');

    if (typeof window.iniciar_streaming_cancion === 'function') {
        return window.iniciar_streaming_cancion(titulo, formato);
    }
    if (typeof window.iniciarStreamGRPCImpl === 'function') {
        return window.iniciarStreamGRPCImpl(titulo, formato);
    }
    eco('No se encontró la implementación de iniciar_streaming_cancion (¿falta bundle.js?).', 'error');
}

// ------------------------------------------------------------------ Conexión con el servidor de reacciones

function conectarReacciones() {
    clearTimeout(estadoReacciones.temporizadorReconexion);
    if (estadoReacciones.clienteStomp) {
        return;
    }
    if (typeof SockJS === 'undefined' || typeof Stomp === 'undefined') {
        eco('No se cargaron las librerías SockJS/STOMP (carpeta lib).', 'error');
        return;
    }

    eco(`Conectando con el servidor de reacciones (${CONFIG.URL_REACCIONES})…`, 'info');
    mostrarEstadoConexion('conectando');

    const socket = new SockJS(CONFIG.URL_REACCIONES);
    const cliente = Stomp.over(socket);
    cliente.debug = null; // sin trazas de frames STOMP; se usan los ecos propios
    estadoReacciones.clienteStomp = cliente;

    cliente.connect({}, () => alConectarReacciones(cliente), (error) => alPerderConexion(cliente, error));
}

function alConectarReacciones(cliente) {
    if (estadoReacciones.clienteStomp !== cliente) return;
    estadoReacciones.conectado = true;
    mostrarEstadoConexion('conectado');
    eco('Conectado al servidor de reacciones por WebSocket.', 'success');

    // Si ya está sonando un audio (p. ej. tras una reconexión), se registra en su canal.
    const audio = document.getElementById('audio-player');
    if (audio && !audio.paused) {
        registrarCanal();
    }
}

function alPerderConexion(cliente, error) {
    if (estadoReacciones.clienteStomp !== cliente) return;
    estadoReacciones.clienteStomp = null;
    estadoReacciones.conectado = false;
    estadoReacciones.suscripcion = null;
    estadoReacciones.canalRegistrado = null;
    mostrarEstadoConexion('desconectado');
    actualizarPanelReacciones();

    const segundos = CONFIG.REINTENTO_REACCIONES_MS / 1000;
    const detalle = typeof error === 'string' ? error : (error && error.headers && error.headers.message) || 'conexión cerrada';
    eco(`Sin conexión con el servidor de reacciones (${detalle}). Reintento en ${segundos} s.`, 'error');
    estadoReacciones.temporizadorReconexion = setTimeout(conectarReacciones, CONFIG.REINTENTO_REACCIONES_MS);
}

// ------------------------------------------------------------------ Registro / des-registro en canales

/** Registra al cliente en el canal del audio que está reproduciendo (SUBSCRIBE). */
function registrarCanal() {
    const audioActual = estadoReacciones.audioActual;
    if (!audioActual) return;
    if (!estadoReacciones.conectado) {
        eco('Reproduciendo sin conexión al servidor de reacciones: no se registró en el canal.', 'error');
        return;
    }
    if (estadoReacciones.canalRegistrado === audioActual.canal) return;
    desregistrarCanal('cambio de canal');

    const destino = `/canalesReacciones/audio/${audioActual.canal}`;
    estadoReacciones.suscripcion = estadoReacciones.clienteStomp.subscribe(destino, recibirReaccion, {
        usuario: valorCabecera(obtenerUsuario()),
        tituloAudio: valorCabecera(audioActual.titulo),
    });
    estadoReacciones.canalRegistrado = audioActual.canal;
    eco(`REGISTRADO en el canal de reacciones "${audioActual.canal}" (SUBSCRIBE ${destino})`, 'success');
    actualizarPanelReacciones();
}

/** Des-registra al cliente del canal (UNSUBSCRIBE) usando la conexión ya creada. */
function desregistrarCanal(motivo) {
    if (!estadoReacciones.suscripcion) return;
    const canal = estadoReacciones.canalRegistrado;
    try {
        estadoReacciones.suscripcion.unsubscribe();
    } catch (e) {
        console.warn('No se pudo enviar UNSUBSCRIBE:', e);
    }
    estadoReacciones.suscripcion = null;
    estadoReacciones.canalRegistrado = null;
    eco(`DES-REGISTRADO del canal "${canal}" (${motivo}).`, 'error');
    actualizarPanelReacciones();
}

// ------------------------------------------------------------------ Envío y recepción de reacciones

function enviarReaccionServidor(tipo, boton) {
    if (!estadoReacciones.suscripcion) {
        eco('Para reaccionar debe estar reproduciendo un audio.', 'error');
        return;
    }
    if (boton) {
        boton.classList.add('gold-glow');
        setTimeout(() => boton.classList.remove('gold-glow'), 300);
    }

    const audioActual = estadoReacciones.audioActual;
    const reaccion = {
        idCanal: audioActual.canal,
        tituloAudio: audioActual.titulo,
        usuario: obtenerUsuario(),
        tipo,
    };
    estadoReacciones.clienteStomp.send(`/app/enviarReaccion/${audioActual.canal}`,
        { 'content-type': 'application/json' }, JSON.stringify(reaccion));
    eco(`Reacción "${tipo}" enviada al canal "${audioActual.canal}" (SEND /app/enviarReaccion/${audioActual.canal})`, 'info');
}

/** Callback: el servidor reenvía aquí las reacciones de los oyentes del mismo audio. */
function recibirReaccion(mensaje) {
    let reaccion;
    try {
        reaccion = JSON.parse(mensaje.body);
    } catch {
        reaccion = { tipo: mensaje.body, usuario: 'Anónimo' };
    }
    const tipo = TIPOS_REACCION[reaccion.tipo];
    if (!tipo) return;

    console.log('[ECO CLIENTE] Reacción reenviada por el servidor:', reaccion);
    mostrarBurbuja(tipo);
    agregarReaccionALista(reaccion, tipo);
}

function mostrarBurbuja(tipo) {
    const contenedor = document.getElementById('mensajes');
    const burbuja = document.createElement('div');
    burbuja.className = 'reaccion-burbuja';
    burbuja.style.left = `${15 + Math.random() * 70}%`;
    const icono = document.createElement('i');
    icono.className = `fa-solid ${tipo.icono} ${tipo.clase}`;
    burbuja.appendChild(icono);
    contenedor.appendChild(burbuja);
    setTimeout(() => burbuja.remove(), 3300);
}

function agregarReaccionALista(reaccion, tipo) {
    const lista = document.getElementById('lista-reacciones');
    const vacio = lista.querySelector('.texto-tenue');
    if (vacio) vacio.remove();

    const item = document.createElement('li');
    const icono = document.createElement('i');
    icono.className = `fa-solid ${tipo.icono} ${tipo.clase}`;
    const texto = document.createElement('span');
    const propia = reaccion.usuario === obtenerUsuario() ? ' (tú)' : '';
    texto.textContent = ` ${reaccion.usuario}${propia}: ${tipo.texto} · ${reaccion.fechaHora || new Date().toLocaleTimeString()}`;
    item.append(icono, texto);
    lista.prepend(item);
    while (lista.children.length > 30) lista.lastChild.remove();
}

// ------------------------------------------------------------------ Interfaz

function obtenerUsuario() {
    const valor = document.getElementById('usuario').value.trim();
    return valor || 'Anónimo';
}

/** Los valores de cabeceras STOMP no pueden llevar ':' ni saltos de línea. */
function valorCabecera(texto) {
    return String(texto).replace(/[:\r\n]/g, ' ');
}

function mostrarEstadoConexion(estado) {
    const indicador = document.getElementById('estado-reacciones');
    const textos = {
        conectando: 'Conectando con reacciones…',
        conectado: 'Reacciones: conectado',
        desconectado: 'Reacciones: sin conexión',
    };
    indicador.textContent = textos[estado];
    indicador.className = `estado-servidor ${estado === 'conectado' ? 'activo' : estado === 'desconectado' ? 'caido' : ''}`;
}

function actualizarPanelReacciones() {
    const registrado = Boolean(estadoReacciones.suscripcion);
    document.querySelectorAll('#reacciones .reaction').forEach((b) => { b.disabled = !registrado; });
    document.getElementById('canal-actual').textContent = registrado
        ? `Registrado en el canal "${estadoReacciones.canalRegistrado}"`
        : (estadoReacciones.audioActual ? 'No registrado (el audio no se está reproduciendo)' : 'Pida un audio y reprodúzcalo para reaccionar');
    document.getElementById('canal-actual').className = registrado ? 'canal registrado' : 'canal';
}

// ------------------------------------------------------------------ Listeners de play / pause

function attachAudioListeners() {
    const audio = document.getElementById('audio-player');
    if (!audio) {
        writeLog('No se encontró el elemento audio#audio-player.', 'error');
        return;
    }
    audio.addEventListener('play', function () {
        writeLog('Reproducción iniciada (play).', 'success');
        document.getElementById('status').textContent = `Reproduciendo "${estadoReacciones.audioActual?.titulo ?? ''}"`;
        registrarCanal();
    });
    audio.addEventListener('pause', function () {
        writeLog(audio.ended ? 'Reproducción finalizada (ended).' : 'Reproducción pausada (pause).', 'error');
        document.getElementById('status').textContent = audio.ended ? 'Audio finalizado' : 'Audio en pausa';
        desregistrarCanal(audio.ended ? 'el audio terminó' : 'el audio se pausó');
    });
}

function iniciarInterfaz() {
    attachAudioListeners();
    const usuario = document.getElementById('usuario');
    if (!usuario.value) {
        usuario.value = `Oyente-${Math.floor(100 + Math.random() * 900)}`;
    }
    // Si cambia el nombre estando registrado, se vuelve a registrar con el nombre nuevo.
    usuario.addEventListener('change', () => {
        if (estadoReacciones.suscripcion) {
            desregistrarCanal('cambio de nombre');
            registrarCanal();
        }
    });
    actualizarPanelReacciones();
    conectarReacciones();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', iniciarInterfaz);
} else {
    iniciarInterfaz();
}

// Exponer funciones para uso manual o pruebas
window.__writeAudioLog = writeLog;
window.__attachAudioListeners = attachAudioListeners;
