// Lógica de la interfaz del cliente web.
//   - Catálogo: tipos -> audios -> detalle -> reproducir por streaming (gRPC-Web)
//   - Reacciones: WebSocket + STOMP; al reproducir se registra en el canal del audio y al pausar se des-registra
//   - Suscripciones: token de idempotencia, temporizadores para reintentar y ecos del pago
//   - Historial: audios reproducidos y cuántas veces (servidor de estadísticas)

const $ = (id) => document.getElementById(id);
const CLAVE_USUARIO = "streaming-usuario";
const CLAVE_ULTIMO_PAGO = "streaming-ultimo-pago";

const estado = {
  usuario: null,
  // streaming
  audioActual: null,      // detalle del audio que se está reproduciendo
  controlStreaming: null, // permite cancelar el stream gRPC
  // reacciones
  clienteStomp: null,
  suscripcionCanal: null, // suscripción STOMP activa al canal del audio
  // suscripciones / pagos
  planes: [],
  ocupado: false,
  inicioPago: 0,
  intervaloCronometro: null,
  ultimoPago: null,
};

const formatoMoneda = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });
const formatearMonto = (valor) => formatoMoneda.format(valor);
const formatearFecha = (iso) => (iso ? new Date(iso).toLocaleString("es-CO") : "—");
const formatearBytes = (b) => (b > 1048576 ? `${(b / 1048576).toFixed(1)} MB` : `${Math.round(b / 1024)} KB`);

/** Crea un elemento DOM. Los textos se asignan con textContent (sin inyección de HTML). */
function el(etiqueta, atributos = {}, hijos = []) {
  const nodo = document.createElement(etiqueta);
  for (const [clave, valor] of Object.entries(atributos)) {
    if (clave === "texto") nodo.textContent = valor;
    else if (clave === "clase") nodo.className = valor;
    else if (clave.startsWith("on")) nodo.addEventListener(clave.slice(2), valor);
    else nodo.setAttribute(clave, valor);
  }
  for (const hijo of [].concat(hijos)) {
    if (hijo != null) nodo.append(hijo);
  }
  return nodo;
}

// ====================================================================== ECOS

/** Eco del cliente: se muestra en la consola del navegador y en el panel "Ecos del cliente". */
function eco(origen, mensaje, tipo = "info") {
  const hora = new Date().toLocaleTimeString("es-CO");
  console.log(`[ECO ${origen}] ${mensaje}`);
  const item = el("li", { clase: tipo }, [el("span", { clase: "hora", texto: hora }), el("span", { clase: "origen", texto: origen }), mensaje]);
  const lista = $("ecos");
  lista.prepend(item);
  while (lista.children.length > 200) lista.lastChild.remove();
}

// ====================================================================== USUARIO

function pedirUsuario() {
  $("input-usuario").value = estado.usuario || "";
  $("modal-usuario").classList.remove("oculto");
  $("input-usuario").focus();
}

function establecerUsuario(evento) {
  evento.preventDefault();
  const nombre = $("input-usuario").value.trim();
  if (!nombre) return;
  estado.usuario = nombre;
  try {
    localStorage.setItem(CLAVE_USUARIO, nombre);
  } catch {
    /* sin almacenamiento local: se pedirá de nuevo al recargar */
  }
  $("nombre-usuario").textContent = nombre;
  $("modal-usuario").classList.add("oculto");
  eco("CLIENTE", `Usuario identificado como "${nombre}"`);
  if (!estado.clienteStomp) conectarReacciones();
}

// ====================================================================== PESTAÑAS

function mostrarSeccion(id) {
  document.querySelectorAll(".pestana").forEach((p) => p.classList.toggle("activa", p.dataset.seccion === id));
  document.querySelectorAll(".seccion").forEach((s) => s.classList.toggle("oculto", s.id !== id));
  if (id === "seccion-historial") cargarHistorial();
  if (id === "seccion-suscripciones" && estado.planes.length === 0) {
    cargarEstadoPasarela();
    cargarPlanes();
  }
}

// ====================================================================== CATÁLOGO

async function cargarTipos() {
  const contenedor = $("lista-tipos");
  eco("REST", "GET /api/tipos -> listar tipos de audio");
  let tipos;
  try {
    tipos = await Api.listarTipos();
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: `No fue posible listar los tipos: ${error.message}` }));
    eco("REST", error.message, "error");
    return;
  }
  contenedor.replaceChildren(
    ...tipos.map((t) =>
      el("button", { clase: "tarjeta tipo", type: "button", "data-id": t.id, onclick: () => seleccionarTipo(t) }, [
        el("h3", { texto: t.nombre }),
        el("p", { texto: `${t.cantidadAudios} audio${t.cantidadAudios === 1 ? "" : "s"}` }),
      ])
    )
  );
}

async function seleccionarTipo(tipo) {
  document.querySelectorAll(".tarjeta.tipo").forEach((t) => t.classList.toggle("seleccionada", Number(t.dataset.id) === tipo.id));
  $("titulo-lista-audios").textContent = `2. Audios de ${tipo.nombre}`;
  const lista = $("lista-audios");
  lista.replaceChildren(el("li", { clase: "texto-tenue", texto: "Cargando…" }));
  eco("REST", `GET /api/tipos/${tipo.id}/audios -> audios de ${tipo.nombre}`);
  try {
    const audios = await Api.listarAudiosDeTipo(tipo.id);
    if (audios.length === 0) {
      lista.replaceChildren(el("li", { clase: "texto-tenue", texto: "No hay audios de este tipo." }));
      return;
    }
    lista.replaceChildren(
      ...audios.map((a) =>
        el("li", {}, el("button", { type: "button", "data-id": a.id, onclick: () => verDetalle(a.id) }, [
          el("span", { clase: "titulo-audio", texto: a.titulo }),
          el("span", { clase: "texto-tenue", texto: a.duracion }),
        ]))
      )
    );
  } catch (error) {
    lista.replaceChildren(el("li", { clase: "texto-tenue", texto: error.message }));
    eco("REST", error.message, "error");
  }
}

async function verDetalle(idAudio) {
  document.querySelectorAll("#lista-audios button").forEach((b) => b.classList.toggle("seleccionado", Number(b.dataset.id) === idAudio));
  const contenedor = $("detalle-audio");
  eco("REST", `GET /api/audios/${idAudio} -> detalle del audio`);
  let audio;
  try {
    audio = await Api.detalleAudio(idAudio);
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: error.message }));
    eco("REST", error.message, "error");
    return;
  }
  const filas = [
    ["Tipo", audio.tipo],
    ...audio.metadatos.map((m) => [m.nombre, m.valor]),
    ["Duración", audio.duracion],
    ["Archivo", audio.nombreArchivo],
    ["Registrado", audio.fechaRegistro],
  ];
  contenedor.replaceChildren(
    el("h3", { clase: "titulo-detalle", texto: audio.titulo }),
    el("dl", { clase: "metadatos" }, filas.flatMap(([k, v]) => [el("dt", { texto: k }), el("dd", { texto: v || "—" })])),
    el("button", { clase: "boton primario", type: "button", texto: "▶ Reproducir", onclick: () => reproducir(audio) })
  );
}

// ====================================================================== STREAMING

function reproducir(audio) {
  detenerReproduccion();
  estado.audioActual = audio;
  const panel = $("panel-reproductor");
  panel.classList.add("oculto"); // se muestra con el primer fragmento
  $("titulo-reproduccion").textContent = audio.titulo;
  $("progreso-streaming").textContent = "conectando…";
  $("lista-reacciones").replaceChildren(el("li", { clase: "texto-tenue", texto: "Al reproducir quedas registrado en el canal de este audio; al pausar, sales del canal." }));
  $("eco-reproductor").textContent = "";

  eco("gRPC-Web", `Solicitando AudioStream("${audio.nombreArchivo}") al servidor de streaming (vía gateway y Envoy)`);
  estado.controlStreaming = window.ClienteStreaming.reproducirAudio({
    urlGateway: CONFIG.URL_GATEWAY,
    audio,
    usuario: estado.usuario,
    reproductor: $("reproductor"),
    alPrimerFragmento: () => {
      panel.classList.remove("oculto");
      panel.scrollIntoView({ behavior: "smooth", block: "nearest" });
      eco("gRPC-Web", "Primer fragmento recibido: el reproductor ahora es visible", "ok");
    },
    alRecibirFragmento: (n, bytes) => {
      $("progreso-streaming").textContent = `${n} fragmentos · ${formatearBytes(bytes)}`;
    },
    alTerminar: (n, bytes) => {
      $("progreso-streaming").textContent = `${n} fragmentos · ${formatearBytes(bytes)} · completo`;
      eco("gRPC-Web", `Streaming completo: ${n} fragmentos (${formatearBytes(bytes)})`, "ok");
    },
    alFallar: (error) => {
      eco("gRPC-Web", `Error en el streaming: ${error.message}`, "error");
      $("progreso-streaming").textContent = "error";
      panel.classList.remove("oculto");
      $("eco-reproductor").textContent = `No fue posible reproducir: ${error.message}`;
    },
  });
}

function detenerReproduccion() {
  if (estado.controlStreaming) {
    estado.controlStreaming.cancelar();
    estado.controlStreaming = null;
  }
  desregistrarDeCanal("cambio de audio");
}

// ====================================================================== REACCIONES (WebSocket + STOMP)

function conectarReacciones() {
  const cliente = new window.ClienteStomp({
    brokerURL: CONFIG.URL_WS,
    reconnectDelay: 5000,
    heartbeatIncoming: 10000,
    heartbeatOutgoing: 10000,
    debug: () => {},
  });

  cliente.onConnect = () => {
    eco("WS", `Conectado al servidor de reacciones (${CONFIG.URL_WS})`, "ok");
    actualizarIndicadorWs(true);
    // Si se reconectó mientras sonaba un audio, se vuelve a registrar en su canal
    estado.suscripcionCanal = null;
    if (estado.audioActual && !$("reproductor").paused) registrarEnCanal();
  };
  cliente.onWebSocketClose = () => {
    if (estado.clienteStomp) actualizarIndicadorWs(false);
  };
  cliente.onStompError = (frame) => eco("WS", `Error STOMP: ${frame.headers.message || "desconocido"}`, "error");

  estado.clienteStomp = cliente;
  eco("WS", `Conectando con el servidor de reacciones en ${CONFIG.URL_WS}…`);
  cliente.activate();
}

function actualizarIndicadorWs(conectado) {
  const indicador = $("estado-ws");
  indicador.className = `indicador ${conectado ? "conectado" : "desconectado"}`;
  indicador.textContent = `Reacciones: ${conectado ? "conectado" : "desconectado"}`;
  if (!conectado) eco("WS", "Conexión con el servidor de reacciones perdida; reintentando cada 5 s", "aviso");
}

/** Registra al cliente en el canal del audio que está sonando (SUBSCRIBE). */
function registrarEnCanal() {
  const audio = estado.audioActual;
  const cliente = estado.clienteStomp;
  if (!audio || estado.suscripcionCanal) return;
  if (!cliente || !cliente.connected) {
    eco("WS", "No hay conexión con el servidor de reacciones: no se pudo registrar en el canal", "aviso");
    return;
  }
  const canal = `/brokerDeReacciones/audio/${audio.id}`;
  estado.suscripcionCanal = cliente.subscribe(canal, (mensaje) => recibirMensajeCanal(JSON.parse(mensaje.body)), { usuario: estado.usuario });
  estado.suscripcionCanal.idAudio = audio.id;
  $("oyentes").textContent = "registrado en el canal";
  eco("WS", `Registrado en el canal ${canal}`, "ok");
}

/** Des-registra al cliente del canal (UNSUBSCRIBE). */
function desregistrarDeCanal(motivo) {
  if (!estado.suscripcionCanal) return;
  const idAudio = estado.suscripcionCanal.idAudio;
  try {
    estado.suscripcionCanal.unsubscribe();
  } catch {
    /* la conexión ya estaba cerrada */
  }
  estado.suscripcionCanal = null;
  $("oyentes").textContent = "canal sin registrar";
  eco("WS", `Des-registrado del canal /brokerDeReacciones/audio/${idAudio} (${motivo})`);
}

/** Callback: el servidor reenvía por la conexión abierta las reacciones del audio. */
function recibirMensajeCanal(datos) {
  const lista = $("lista-reacciones");
  if (lista.firstChild?.classList?.contains("texto-tenue")) lista.replaceChildren();

  if (datos.tipo === "AVISO") {
    $("oyentes").textContent = `${datos.oyentes} oyente${datos.oyentes === 1 ? "" : "s"} en el canal`;
    const accion = datos.evento === "ENTRO" ? "empezó a escuchar" : "dejó de escuchar";
    lista.prepend(el("li", { clase: "aviso-canal", texto: `${datos.usuario} ${accion} este audio` }));
    return;
  }
  const propia = datos.usuario === estado.usuario;
  lista.prepend(
    el("li", { clase: propia ? "propia" : "" }, [
      el("span", { clase: "reaccion", texto: datos.reaccion }),
      el("span", { texto: ` ${propia ? "Tú" : datos.usuario}` }),
      el("span", { clase: "texto-tenue", texto: ` · ${datos.fechaHora?.slice(11) || ""}` }),
    ])
  );
  while (lista.children.length > 50) lista.lastChild.remove();
  eco("WS", `Reacción recibida (callback): ${datos.reaccion} de ${datos.usuario} en "${datos.tituloAudio}"`);
}

function enviarReaccion(reaccion) {
  const audio = estado.audioActual;
  if (!audio || !estado.suscripcionCanal || !estado.clienteStomp?.connected) {
    $("eco-reproductor").textContent = "Reproduce el audio para poder reaccionar (al pausar sales del canal).";
    eco("WS", "Reacción no enviada: no está registrado en el canal del audio", "aviso");
    return;
  }
  estado.clienteStomp.publish({
    destination: `/app/enviarReaccion/${audio.id}`,
    body: JSON.stringify({ idAudio: audio.id, tituloAudio: audio.titulo, usuario: estado.usuario, reaccion }),
  });
  eco("WS", `Reacción enviada: ${reaccion} -> /app/enviarReaccion/${audio.id}`);
}

// Listener del reproductor: se ejecuta cada vez que el usuario reproduce o pausa el audio.
function configurarReproductor() {
  const reproductor = $("reproductor");
  reproductor.addEventListener("play", () => {
    if (!estado.audioActual) return;
    $("eco-reproductor").textContent = `▶ Reproduciendo "${estado.audioActual.titulo}"`;
    eco("REPRODUCTOR", `Se está reproduciendo "${estado.audioActual.titulo}": registro en el canal de notificaciones`);
    registrarEnCanal();
  });
  reproductor.addEventListener("pause", () => {
    if (!estado.audioActual || !reproductor.currentSrc) return;
    if (reproductor.ended) return; // lo maneja "ended"
    $("eco-reproductor").textContent = `⏸ Se ha pausado el audio "${estado.audioActual.titulo}"`;
    eco("REPRODUCTOR", `Se ha pausado el audio "${estado.audioActual.titulo}": se des-registra del canal`);
    desregistrarDeCanal("audio pausado");
  });
  reproductor.addEventListener("ended", () => {
    if (!estado.audioActual) return;
    $("eco-reproductor").textContent = `■ Terminó "${estado.audioActual.titulo}"`;
    eco("REPRODUCTOR", `Terminó el audio "${estado.audioActual.titulo}"`);
    desregistrarDeCanal("audio terminado");
  });

  document.querySelectorAll("#botones-reaccion button[data-reaccion]").forEach((b) =>
    b.addEventListener("click", () => enviarReaccion(b.dataset.reaccion))
  );
  $("form-reaccion").addEventListener("submit", (e) => {
    e.preventDefault();
    const texto = $("texto-reaccion").value.trim();
    if (!texto) return;
    enviarReaccion(texto);
    $("texto-reaccion").value = "";
  });
}

// ====================================================================== SUSCRIPCIONES

async function cargarEstadoPasarela() {
  const indicador = $("estado-pasarela");
  try {
    const datos = await Api.estadoSuscripciones();
    indicador.className = "indicador conectado";
    indicador.textContent =
      datos.modo_fallo === "NINGUNO"
        ? "Servidor activo · sin simulación de fallos"
        : `Servidor activo · simulando pérdida de ${datos.modo_fallo} (${datos.cantidad_fallos} fallos por token)`;
  } catch {
    indicador.className = "indicador desconectado";
    indicador.textContent = "Servidor de suscripciones no disponible";
  }
}

async function cargarPlanes() {
  const contenedor = $("lista-planes");
  eco("REST", "GET /api/suscripciones/planes -> listar planes");
  try {
    estado.planes = await Api.listarPlanes();
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: `No fue posible listar los planes: ${error.message}` }));
    return;
  }
  contenedor.replaceChildren(
    ...estado.planes.map((s) =>
      el("article", { clase: "tarjeta plan", "data-id": s.id }, [
        el("h3", { texto: s.nombre }),
        el("div", { clase: "precio", texto: `${formatearMonto(s.precio)}/mes` }),
        el("p", { texto: `${s.descripcion} (${s.cuentas} cuenta${s.cuentas > 1 ? "s" : ""})` }),
        el("button", { clase: "boton primario mini", type: "button", texto: "Elegir plan", onclick: () => seleccionarPlan(s.id, true) }),
      ])
    )
  );
  $("plan").replaceChildren(...estado.planes.map((s) => el("option", { value: s.id, texto: s.nombre })));
  seleccionarPlan(estado.planes[0]?.id, false);
}

function seleccionarPlan(id, desplazar) {
  const plan = estado.planes.find((s) => s.id === id);
  if (!plan) return;
  $("plan").value = id;
  $("monto").value = formatearMonto(plan.precio);
  document.querySelectorAll(".tarjeta.plan").forEach((t) => t.classList.toggle("seleccionada", t.dataset.id === id));
  if (desplazar) $("form-pago").scrollIntoView({ behavior: "smooth", block: "center" });
}

function limpiarBitacora() {
  $("bitacora").replaceChildren();
  $("resultado").className = "resultado oculto";
}

function registrar({ tipo, mensaje }) {
  const segundos = ((Date.now() - estado.inicioPago) / 1000).toFixed(1);
  const hora = new Date().toLocaleTimeString("es-CO");
  const item = el("li", { clase: tipo }, [el("span", { clase: "hora", texto: `${hora} (+${segundos} s)` }), mensaje]);
  $("bitacora").append(item);
  item.scrollIntoView({ block: "nearest" });
  eco("PAGO", mensaje, tipo === "ok" ? "ok" : tipo === "error" ? "error" : tipo === "aviso" ? "aviso" : "info");
}

function iniciarCronometro() {
  estado.inicioPago = Date.now();
  clearInterval(estado.intervaloCronometro);
  estado.intervaloCronometro = setInterval(() => {
    $("cronometro").textContent = `${((Date.now() - estado.inicioPago) / 1000).toFixed(1)} s`;
  }, 100);
}

function detenerCronometro() {
  clearInterval(estado.intervaloCronometro);
}

function mostrarResultado(tipo, titulo, filas = []) {
  const caja = $("resultado");
  caja.className = `resultado ${tipo}`;
  caja.replaceChildren(el("h3", { texto: titulo }), el("dl", {}, filas.flatMap(([k, v]) => [el("dt", { texto: k }), el("dd", { texto: v })])));
}

function setOcupado(ocupado) {
  estado.ocupado = ocupado;
  $("btn-pagar").disabled = ocupado;
  $("btn-simular-token").disabled = ocupado;
  $("btn-reenviar").disabled = ocupado || !leerUltimoPago();
  document.querySelectorAll("[data-reenviar]").forEach((b) => (b.disabled = ocupado));
}

function guardarUltimoPago(pago) {
  try {
    localStorage.setItem(CLAVE_ULTIMO_PAGO, JSON.stringify(pago));
  } catch {
    /* sin almacenamiento local */
  }
  estado.ultimoPago = pago;
  mostrarUltimoPago();
}

function leerUltimoPago() {
  if (estado.ultimoPago) return estado.ultimoPago;
  try {
    estado.ultimoPago = JSON.parse(localStorage.getItem(CLAVE_ULTIMO_PAGO)) || null;
  } catch {
    estado.ultimoPago = null;
  }
  return estado.ultimoPago;
}

function mostrarUltimoPago() {
  const pago = leerUltimoPago();
  $("ultimo-token").textContent = pago ? `${pago.token} · ${formatearMonto(pago.monto)}` : "—";
  $("btn-reenviar").disabled = estado.ocupado || !pago;
}

async function procesarPago(evento) {
  evento.preventDefault();
  if (estado.ocupado) return;
  const plan = estado.planes.find((s) => s.id === $("plan").value);
  if (!plan) return;

  const datosPago = {
    cedula: $("cedula").value.trim(),
    suscripcion_id: plan.id,
    monto: plan.precio,
    nombre_titular: $("nombre-titular").value.trim(),
    numero_tarjeta: $("tarjeta").value.replace(/\s+/g, ""),
  };

  limpiarBitacora();
  iniciarCronometro();
  setOcupado(true);
  try {
    registrar({ tipo: "info", mensaje: `Solicitando token de idempotencia (POST /suscripciones/tokens) para el plan ${plan.nombre}…` });
    const token = await Api.generarToken(datosPago);
    registrar({ tipo: "ok", mensaje: `Token de idempotencia obtenido: ${token.token_idempotencia}` });
    guardarUltimoPago({ token: token.token_idempotencia, monto: token.monto, cedula: token.cedula });
    await ejecutarPago(token.token_idempotencia, token.monto);
  } catch (error) {
    registrar({ tipo: "error", mensaje: error.message });
    mostrarResultado("error", "No fue posible obtener el token de idempotencia", [["Detalle", error.message]]);
  } finally {
    detenerCronometro();
    setOcupado(false);
  }
}

async function reenviarConToken(token, monto, motivo = "Reenviando pago con un token ya existente") {
  if (estado.ocupado) return;
  limpiarBitacora();
  iniciarCronometro();
  setOcupado(true);
  registrar({ tipo: "info", mensaje: `${motivo}: ${token}` });
  try {
    await ejecutarPago(token, monto);
  } finally {
    detenerCronometro();
    setOcupado(false);
  }
}

/** Opción e) del requerimiento: reenvía un token con el que YA se pagó. */
async function simularTokenUtilizado() {
  if (estado.ocupado) return;
  const cedula = $("cedula").value.trim();
  eco("PAGO", `Buscando un token ya utilizado para la cédula ${cedula}…`);
  let datos;
  try {
    datos = await Api.consultarTransacciones(cedula);
  } catch (error) {
    eco("PAGO", `No fue posible consultar las transacciones: ${error.message}`, "error");
    return;
  }
  const usado = datos.tokens_utilizados[datos.tokens_utilizados.length - 1];
  if (!usado) {
    limpiarBitacora();
    registrar({ tipo: "aviso", mensaje: `La cédula ${cedula} aún no tiene pagos: realice primero un pago para tener un token utilizado.` });
    return;
  }
  await reenviarConToken(usado.token_idempotencia, usado.monto, "SIMULACIÓN: se envía un token previamente utilizado");
}

async function ejecutarPago(token, monto) {
  try {
    const { datos, intento } = await Api.realizarPagoConReintentos(token, monto, registrar);
    const enQueIntento = intento === 0 ? "Primer intento" : `Reintento ${intento}`;

    if (datos.transaccion_realizada) {
      mostrarResultado("ok", "✔ Pago realizado exitosamente", [
        ["Token utilizado", datos.token_idempotencia],
        ["Obtenido en", enQueIntento],
        ["Monto", formatearMonto(datos.monto_transaccion)],
        ["¿Pago realizado?", "Sí"],
        ["¿Pago almacenado?", datos.almacenado ? "Sí (transacciones.json)" : "No"],
        ["ID transacción", datos.id_transaccion],
        ["Fecha", formatearFecha(datos.fecha_transaccion)],
      ]);
      registrar({ tipo: "ok", mensaje: `Pago realizado exitosamente con el token ${datos.token_idempotencia}` });
      registrar({ tipo: datos.almacenado ? "ok" : "error", mensaje: datos.almacenado ? `El pago quedó almacenado (transacción ${datos.id_transaccion})` : "El pago NO quedó almacenado" });
    } else if (datos.previamente_realizada) {
      mostrarResultado("aviso", "⚠ Este token ya había sido utilizado: el pago no se ejecutó de nuevo", [
        ["Token utilizado", datos.token_idempotencia],
        ["Obtenido en", enQueIntento],
        ["¿Pago realizado ahora?", "No"],
        ["¿Realizado previamente?", "Sí"],
        ["Monto de la transacción previa", formatearMonto(datos.monto_transaccion_previa)],
        ["ID transacción previa", datos.id_transaccion],
        ["Fecha transacción previa", formatearFecha(datos.fecha_transaccion)],
        ...(intento > 0 ? [["Interpretación", "Un intento anterior SÍ fue procesado (se perdió su respuesta); el reintento con el mismo token no cobró dos veces."]] : []),
      ]);
      registrar({ tipo: "aviso", mensaje: datos.mensaje });
    }
  } catch (error) {
    if (error instanceof ErrorSinRespuesta) {
      registrar({ tipo: "error", mensaje: error.message });
      mostrarResultado("error", "✖ No fue posible acceder al servidor para realizar el pago", [
        ["Token utilizado", token],
        ["Intentos", `1 intento inicial + ${CONFIG.ESPERAS_REINTENTO_MS.length} reintentos sin respuesta`],
        ["Sugerencia", "Puede reenviar el pago con el mismo token: si el servidor alcanzó a procesarlo, no se cobrará dos veces."],
      ]);
    } else {
      registrar({ tipo: "error", mensaje: error.message });
      mostrarResultado("error", "✖ El servidor rechazó el pago", [["Detalle", error.message]]);
    }
  }
}

async function consultarTransacciones(evento) {
  evento?.preventDefault();
  const contenedor = $("resultado-consulta");
  const cedula = $("cedula-consulta").value.trim();
  contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: "Consultando…" }));
  let datos;
  try {
    datos = await Api.consultarTransacciones(cedula);
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: `Error: ${error.message}` }));
    return;
  }
  const nombrePlan = (id) => estado.planes.find((s) => s.id === id)?.nombre || id;
  contenedor.replaceChildren(
    el("p", { clase: "resumen", texto: `Cédula ${datos.cedula} · transacciones realizadas: ${datos.total_transacciones}` }),
    datos.transacciones.length === 0
      ? el("p", { clase: "texto-tenue", texto: "No hay transacciones registradas." })
      : el("div", { clase: "tabla-contenedor" },
          el("table", {}, [
            el("thead", {}, el("tr", {}, ["#", "Fecha", "Plan", "Monto", "Token de idempotencia", ""].map((t) => el("th", { texto: t })))),
            el("tbody", {}, datos.transacciones.map((t, i) =>
              el("tr", {}, [
                el("td", { texto: String(i + 1) }),
                el("td", { texto: formatearFecha(t.fecha) }),
                el("td", { texto: nombrePlan(t.suscripcion_id) }),
                el("td", { texto: formatearMonto(t.monto) }),
                el("td", {}, el("code", { texto: t.token_idempotencia })),
                el("td", {}, el("button", {
                  clase: "boton secundario mini", type: "button", "data-reenviar": "", texto: "Reenviar con este token",
                  onclick: () => reenviarConToken(t.token_idempotencia, t.monto),
                })),
              ])
            )),
          ])
        )
  );
}

// ====================================================================== HISTORIAL

async function cargarHistorial() {
  const contenedor = $("historial");
  if (!estado.usuario) return;
  eco("REST", `GET /api/estadisticas/clientes/${estado.usuario} -> audios reproducidos`);
  let datos;
  try {
    datos = await Api.historialCliente(estado.usuario);
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: `No fue posible consultar el historial: ${error.message}` }));
    return;
  }
  if (datos.audios.length === 0) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: `${datos.cliente} aún no ha reproducido audios.` }));
    return;
  }
  contenedor.replaceChildren(
    el("p", { clase: "resumen", texto: `${datos.cliente} ha hecho ${datos.totalReproducciones} reproducciones de ${datos.audios.length} audios distintos.` }),
    el("div", { clase: "tabla-contenedor" },
      el("table", {}, [
        el("thead", {}, el("tr", {}, ["Audio", "Tipo", "Veces reproducido", "Última reproducción"].map((t) => el("th", { texto: t })))),
        el("tbody", {}, datos.audios.map((a) =>
          el("tr", {}, [
            el("td", { texto: a.titulo }),
            el("td", { texto: a.tipo }),
            el("td", { clase: "numero", texto: String(a.cantidad) }),
            el("td", { texto: a.ultimaReproduccion }),
          ])
        )),
      ])
    )
  );
}

// ====================================================================== INICIO

document.addEventListener("DOMContentLoaded", () => {
  if (!window.ClienteStreaming) {
    alert("Falta dist/main.js. Ejecute \"npx webpack\" en la carpeta ClienteWeb y recargue la página.");
    return;
  }
  document.querySelectorAll(".pestana").forEach((p) => p.addEventListener("click", () => mostrarSeccion(p.dataset.seccion)));
  $("form-usuario").addEventListener("submit", establecerUsuario);
  $("btn-cambiar-usuario").addEventListener("click", pedirUsuario);
  $("btn-recargar-tipos").addEventListener("click", cargarTipos);
  $("btn-recargar-historial").addEventListener("click", cargarHistorial);
  $("btn-limpiar-ecos").addEventListener("click", () => $("ecos").replaceChildren());
  $("form-pago").addEventListener("submit", procesarPago);
  $("form-consulta").addEventListener("submit", consultarTransacciones);
  $("plan").addEventListener("change", (e) => seleccionarPlan(e.target.value, false));
  $("btn-reenviar").addEventListener("click", () => {
    const pago = leerUltimoPago();
    if (pago) reenviarConToken(pago.token, pago.monto);
  });
  $("btn-simular-token").addEventListener("click", simularTokenUtilizado);
  configurarReproductor();
  mostrarUltimoPago();

  try {
    estado.usuario = localStorage.getItem(CLAVE_USUARIO);
  } catch {
    estado.usuario = null;
  }
  if (estado.usuario) {
    $("nombre-usuario").textContent = estado.usuario;
    conectarReacciones();
  } else {
    pedirUsuario();
  }
  cargarTipos();
});
