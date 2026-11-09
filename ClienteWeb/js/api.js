// Acceso a los servicios REST a través del API Gateway (basado en la guía 10).

/** El servidor respondió con un error de negocio (4xx): no tiene sentido reintentar. */
class ErrorServidor extends Error {
  constructor(mensaje, estado) {
    super(mensaje);
    this.estado = estado;
  }
}

/** Se agotaron los reintentos sin obtener respuesta del servidor. */
class ErrorSinRespuesta extends Error {}

const Api = (() => {
  const dormir = (ms) => new Promise((resolver) => setTimeout(resolver, ms));

  /** Hace una petición y la aborta si el servidor no responde dentro de `tiempoEsperaMs`. */
  async function fetchConTiempoEspera(url, opciones = {}, tiempoEsperaMs = CONFIG.TIEMPO_ESPERA_MS) {
    const controlador = new AbortController();
    const temporizador = setTimeout(() => controlador.abort(), tiempoEsperaMs);
    try {
      const respuesta = await fetch(url, { ...opciones, signal: controlador.signal });
      const datos = await respuesta.json().catch(() => ({}));
      return { respuesta, datos };
    } catch (error) {
      if (error.name === "AbortError") {
        const e = new Error(`el servidor no respondió en ${tiempoEsperaMs / 1000} s`);
        e.tiempoAgotado = true;
        throw e;
      }
      throw new Error(`fallo de red (${error.message})`);
    } finally {
      clearTimeout(temporizador);
    }
  }

  async function peticionSimple(ruta, opciones, tiempoEsperaMs = 8000) {
    let resultado;
    try {
      resultado = await fetchConTiempoEspera(CONFIG.URL_API + ruta, opciones, tiempoEsperaMs);
    } catch (error) {
      throw new ErrorSinRespuesta(`No fue posible acceder al API Gateway: ${error.message}`);
    }
    if (!resultado.respuesta.ok) {
      throw new ErrorServidor(resultado.datos.error || `HTTP ${resultado.respuesta.status}`, resultado.respuesta.status);
    }
    return resultado.datos;
  }

  function postJSON(ruta, cuerpo) {
    return peticionSimple(ruta, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(cuerpo),
    });
  }

  return {
    // ---------- Servidor de audios y metadatos
    listarTipos: () => peticionSimple("/tipos"),
    listarAudiosDeTipo: (idTipo) => peticionSimple(`/tipos/${idTipo}/audios`),
    detalleAudio: (idAudio) => peticionSimple(`/audios/${idAudio}`),

    // ---------- Servidor de estadísticas
    historialCliente: (cliente) => peticionSimple(`/estadisticas/clientes/${encodeURIComponent(cliente)}`),

    // ---------- Servidor de suscripciones
    estadoSuscripciones: () => peticionSimple("/suscripciones/estado"),
    listarPlanes: () => peticionSimple("/suscripciones/planes"),
    generarToken: (datosPago) => postJSON("/suscripciones/tokens", datosPago),
    consultarTransacciones: (cedula) => peticionSimple(`/suscripciones/transacciones/${encodeURIComponent(cedula)}`),

    /**
     * Pago con mecanismo de reintentos (temporizadores). Envía siempre el MISMO token de
     * idempotencia: si el servidor ya procesó un intento anterior, no vuelve a cobrar.
     */
    async realizarPagoConReintentos(token, monto, notificar) {
      const esperas = CONFIG.ESPERAS_REINTENTO_MS;
      const cuerpo = JSON.stringify({ token_idempotencia: token, monto });

      for (let intento = 0; intento <= esperas.length; intento++) {
        if (intento > 0) {
          const espera = esperas[intento - 1];
          notificar({ tipo: "espera", mensaje: `Temporizador: esperando ${espera / 1000} s antes del reintento ${intento} de ${esperas.length}…` });
          await dormir(espera);
        }

        const etiqueta = intento === 0 ? "Intento inicial" : `Reintento ${intento} de ${esperas.length}`;
        notificar({ tipo: "info", mensaje: `${etiqueta}: POST /suscripciones/pagos usando el token ${token}` });

        let resultado;
        try {
          resultado = await fetchConTiempoEspera(CONFIG.URL_API + "/suscripciones/pagos", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: cuerpo,
          });
        } catch (error) {
          notificar({ tipo: "aviso", mensaje: `${etiqueta} fallido: ${error.message}` });
          continue;
        }

        const { respuesta, datos } = resultado;
        if (respuesta.status >= 400 && respuesta.status < 500) {
          notificar({ tipo: "error", mensaje: `${etiqueta}: el servidor rechazó el pago (HTTP ${respuesta.status})` });
          throw new ErrorServidor(datos.error || `HTTP ${respuesta.status}`, respuesta.status);
        }
        if (!respuesta.ok) {
          notificar({ tipo: "aviso", mensaje: `${etiqueta} fallido: HTTP ${respuesta.status} (${datos.error || "sin respuesta del servicio"})` });
          continue;
        }

        notificar({ tipo: "ok", mensaje: `${etiqueta}: respuesta recibida (HTTP ${respuesta.status})` });
        return { datos, intento };
      }

      throw new ErrorSinRespuesta(
        `No fue posible acceder al servidor para realizar el pago (sin respuesta tras ${esperas.length} reintentos).`
      );
    },
  };
})();
