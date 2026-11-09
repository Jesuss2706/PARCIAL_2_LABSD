// Cliente gRPC-Web que recibe del servidor de streaming los fragmentos de un audio.
//
//   navegador --gRPC-Web (HTTP/1.1)--> API Gateway :8000 --> Proxy Envoy :8080 --gRPC (HTTP/2)--> Servidor de Streaming :50051
//
// Si modifica este archivo ejecute "npx webpack" (o "npm run build") para regenerar dist/main.js.

const { AudioRequest } = require("./serviciosAudio/audio_pb.js");
const { AudioServiceClient } = require("./serviciosAudio/audio_grpc_web_pb.js");
const { Client: ClienteStomp } = require("@stomp/stompjs");

const TIPO_MIME = "audio/mpeg";

/**
 * Reproduce un audio por streaming.
 *
 * @param {object} opciones
 * @param {string} opciones.urlGateway       URL del API Gateway (p. ej. http://localhost:8000)
 * @param {object} opciones.audio            detalle del audio ({id, titulo, tipo, nombreArchivo})
 * @param {string} opciones.usuario          nombre del cliente (para las estadísticas)
 * @param {HTMLAudioElement} opciones.reproductor  elemento <audio> donde se reproduce
 * @param {(n:number, bytes:number)=>void} [opciones.alRecibirFragmento]
 * @param {()=>void} [opciones.alPrimerFragmento]
 * @param {(total:number, bytes:number)=>void} [opciones.alTerminar]
 * @param {(error:Error)=>void} [opciones.alFallar]
 * @returns {{cancelar: () => void}}
 */
function reproducirAudio(opciones) {
  const { urlGateway, audio, usuario, reproductor } = opciones;
  const alRecibirFragmento = opciones.alRecibirFragmento || (() => {});
  const alPrimerFragmento = opciones.alPrimerFragmento || (() => {});
  const alTerminar = opciones.alTerminar || (() => {});
  const alFallar = opciones.alFallar || (() => {});

  const cliente = new AudioServiceClient(urlGateway, null, null);
  const peticion = new AudioRequest();
  peticion.setFilename(audio.nombreArchivo);
  peticion.setIdAudio(audio.id);
  peticion.setTitulo(audio.titulo);
  peticion.setTipo(audio.tipo);
  peticion.setCliente(usuario);

  console.log(`[ECO gRPC-Web] Conectando con el servidor de streaming: AudioStream(${audio.nombreArchivo}) vía ${urlGateway}`);

  let cancelado = false;
  let numeroFragmento = 0;
  let totalBytes = 0;
  const fragmentos = []; // copia de todo lo recibido (para el respaldo sin MediaSource)

  // ---- Reproducción progresiva con MediaSource: cada fragmento se agrega al buffer del <audio>
  const usarMediaSource = typeof MediaSource !== "undefined" && MediaSource.isTypeSupported(TIPO_MIME);
  let mediaSource = null;
  let sourceBuffer = null;
  const pendientes = [];
  let streamTerminado = false;
  let urlObjeto = null;

  function vaciarPendientes() {
    if (!sourceBuffer || sourceBuffer.updating) return;
    if (pendientes.length > 0) {
      try {
        sourceBuffer.appendBuffer(pendientes.shift());
      } catch (e) {
        console.warn("[ECO gRPC-Web] No se pudo agregar el fragmento al buffer:", e);
      }
      return;
    }
    if (streamTerminado && mediaSource.readyState === "open") {
      mediaSource.endOfStream();
    }
  }

  if (usarMediaSource) {
    mediaSource = new MediaSource();
    urlObjeto = URL.createObjectURL(mediaSource);
    reproductor.src = urlObjeto;
    mediaSource.addEventListener("sourceopen", () => {
      sourceBuffer = mediaSource.addSourceBuffer(TIPO_MIME);
      sourceBuffer.addEventListener("updateend", vaciarPendientes);
      vaciarPendientes();
    }, { once: true });
  }

  const stream = cliente.audioStream(peticion, {});

  stream.on("data", (respuesta) => {
    if (cancelado) return;
    const datos = respuesta.getData_asU8();
    numeroFragmento++;
    totalBytes += datos.length;
    fragmentos.push(datos);
    console.log(`[ECO gRPC-Web] Fragmento #${numeroFragmento} recibido (${datos.length} bytes)`);
    alRecibirFragmento(numeroFragmento, totalBytes);

    if (usarMediaSource) {
      pendientes.push(datos);
      vaciarPendientes();
    }
    if (numeroFragmento === 1) {
      alPrimerFragmento();
      if (usarMediaSource) reproductor.play().catch(() => {});
    }
  });

  stream.on("status", (estado) => {
    if (estado && estado.code !== 0) {
      console.warn(`[ECO gRPC-Web] Estado gRPC ${estado.code}: ${estado.details}`);
    }
  });

  stream.on("error", (error) => {
    if (cancelado) return;
    console.error("[ECO gRPC-Web] Error en el streaming:", error);
    alFallar(new Error(error.message || "error en el streaming"));
  });

  stream.on("end", () => {
    if (cancelado) return;
    console.log(`[ECO gRPC-Web] Streaming terminado: ${numeroFragmento} fragmentos, ${totalBytes} bytes`);
    streamTerminado = true;
    if (usarMediaSource) {
      vaciarPendientes();
    } else {
      // Respaldo: sin MediaSource se reproduce cuando llega el audio completo
      urlObjeto = URL.createObjectURL(new Blob(fragmentos, { type: TIPO_MIME }));
      reproductor.src = urlObjeto;
      alPrimerFragmento();
      reproductor.play().catch(() => {});
    }
    alTerminar(numeroFragmento, totalBytes);
  });

  return {
    cancelar() {
      if (cancelado) return;
      cancelado = true;
      try {
        stream.cancel();
      } catch (e) {
        /* el stream ya había terminado */
      }
      reproductor.pause();
      reproductor.removeAttribute("src");
      reproductor.load();
      if (urlObjeto) URL.revokeObjectURL(urlObjeto);
      console.log(`[ECO gRPC-Web] Streaming de "${audio.titulo}" cancelado por el cliente`);
    },
  };
}

// Se publican como globales para que funciones.js (script normal) los use.
window.ClienteStreaming = { reproducirAudio };
window.ClienteStomp = ClienteStomp;
