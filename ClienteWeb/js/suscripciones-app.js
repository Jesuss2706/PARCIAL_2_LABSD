// Lógica de la página de planes de suscripción (laboratorio de tokens idempotentes).

const estado = {
  suscripciones: [],
  ocupado: false,
  inicioPago: 0,
  intervaloCronometro: null,
};

const $ = (id) => document.getElementById(id);
const CLAVE_ULTIMO_PAGO = "spotify-ultimo-pago";

const formatoMoneda = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });
const formatearMonto = (valor) => formatoMoneda.format(valor);
const formatearFecha = (iso) => (iso ? new Date(iso).toLocaleString("es-CO") : "—");
const nombreSuscripcion = (id) => estado.suscripciones.find((s) => s.id === id)?.nombre || id;

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

// ---------------------------------------------------------------- Estado del servidor

async function cargarEstadoServidor() {
  const indicador = $("estado-servidor");
  try {
    const datos = await Api.estado();
    indicador.className = "estado-servidor activo";
    indicador.textContent =
      datos.modo_fallo === "NINGUNO"
        ? "Suscripciones: activo · sin simulación de fallos"
        : `Suscripciones: activo · simulando pérdida de ${datos.modo_fallo} (${datos.cantidad_fallos} fallos por token)`;
  } catch {
    indicador.className = "estado-servidor caido";
    indicador.textContent = "Suscripciones: no disponible";
  }
}

// ---------------------------------------------------------------- Suscripciones

async function cargarSuscripciones() {
  const contenedor = $("lista-suscripciones");
  try {
    estado.suscripciones = await Api.listarPlanes();
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: `No fue posible listar las suscripciones: ${error.message}` }));
    return;
  }

  contenedor.replaceChildren(
    ...estado.suscripciones.map((s) =>
      el("article", { clase: "tarjeta", "data-id": s.id }, [
        el("h3", { texto: s.nombre }),
        el("div", { clase: "precio", texto: `${formatearMonto(s.precio)}/mes` }),
        el("p", { texto: `${s.descripcion} (${s.cuentas} cuenta${s.cuentas > 1 ? "s" : ""})` }),
        el("button", { clase: "boton primario", type: "button", texto: "Elegir plan", onclick: () => seleccionarSuscripcion(s.id, true) }),
      ])
    )
  );

  const selector = $("suscripcion");
  const actual = selector.value;
  selector.replaceChildren(...estado.suscripciones.map((s) => el("option", { value: s.id, texto: s.nombre })));
  seleccionarSuscripcion(actual || estado.suscripciones[0]?.id, false);
}

function seleccionarSuscripcion(id, desplazar) {
  const suscripcion = estado.suscripciones.find((s) => s.id === id);
  if (!suscripcion) return;
  $("suscripcion").value = id;
  $("monto").value = formatearMonto(suscripcion.precio);
  document.querySelectorAll(".tarjeta").forEach((t) => t.classList.toggle("seleccionada", t.dataset.id === id));
  if (desplazar) $("form-pago").scrollIntoView({ behavior: "smooth", block: "center" });
}

// ---------------------------------------------------------------- Bitácora

function limpiarBitacora() {
  $("bitacora").replaceChildren();
  $("resultado").className = "resultado oculto";
}

function registrar({ tipo, mensaje }) {
  console.log(`[ECO CLIENTE] ${mensaje}`);
  const segundos = ((Date.now() - estado.inicioPago) / 1000).toFixed(1);
  const hora = new Date().toLocaleTimeString("es-CO");
  const item = el("li", { clase: tipo }, [el("span", { clase: "hora", texto: `${hora} (+${segundos} s)` }), mensaje]);
  $("bitacora").append(item);
  item.scrollIntoView({ block: "nearest" });
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
  const lista = el("dl", {}, filas.flatMap(([etiqueta, valor]) => [el("dt", { texto: etiqueta }), el("dd", { texto: valor })]));
  caja.replaceChildren(el("h3", { texto: titulo }), lista);
}

function setOcupado(ocupado) {
  estado.ocupado = ocupado;
  $("btn-pagar").disabled = ocupado;
  $("btn-reenviar").disabled = ocupado || !leerUltimoPago();
  document.querySelectorAll("[data-reenviar]").forEach((b) => (b.disabled = ocupado));
}

// ---------------------------------------------------------------- Último token (para reenviarlo)

function guardarUltimoPago(pago) {
  try {
    localStorage.setItem(CLAVE_ULTIMO_PAGO, JSON.stringify(pago));
  } catch {
    /* almacenamiento no disponible: solo se pierde la comodidad de recordar el token */
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

// ---------------------------------------------------------------- Pago

async function procesarPago(evento) {
  evento.preventDefault();
  if (estado.ocupado) return;

  const suscripcion = estado.suscripciones.find((s) => s.id === $("suscripcion").value);
  if (!suscripcion) return;

  const datosPago = {
    cedula: $("cedula").value.trim(),
    id_plan: suscripcion.id,
    monto: suscripcion.precio,
    nombre_titular: $("nombre").value.trim(),
    numero_tarjeta: $("tarjeta").value.replace(/\s+/g, ""),
  };

  limpiarBitacora();
  iniciarCronometro();
  setOcupado(true);
  try {
    registrar({ tipo: "info", mensaje: `Solicitando token de idempotencia (POST /api/suscripciones/tokens) para ${suscripcion.nombre}…` });
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

async function reenviarConToken(token, monto) {
  if (estado.ocupado) return;
  limpiarBitacora();
  iniciarCronometro();
  setOcupado(true);
  registrar({ tipo: "aviso", mensaje: `Simulando el uso de un token previamente utilizado: ${token}` });
  try {
    await ejecutarPago(token, monto);
  } finally {
    detenerCronometro();
    setOcupado(false);
  }
}

async function ejecutarPago(token, monto) {
  registrar({ tipo: "info", mensaje: `Token de idempotencia utilizado en el pago: ${token}` });
  try {
    const { datos, intento } = await Api.realizarPagoConReintentos(token, monto, registrar);
    const enQueIntento = intento === 0 ? "Primer intento" : `Reintento ${intento}`;

    if (datos.transaccion_realizada) {
      registrar({ tipo: "ok", mensaje: `Pago realizado exitosamente (transacción ${datos.id_transaccion})` });
      registrar({
        tipo: datos.pago_almacenado ? "ok" : "error",
        mensaje: datos.pago_almacenado ? "El servidor confirmó que el pago quedó almacenado" : "El servidor NO confirmó el almacenamiento del pago",
      });
      mostrarResultado("ok", "✔ Pago realizado exitosamente", [
        ["Obtenido en", enQueIntento],
        ["Monto de la transacción", formatearMonto(datos.monto_transaccion)],
        ["Transacción realizada", "Sí"],
        ["Pago almacenado", datos.pago_almacenado ? "Sí" : "No"],
        ["Realizada previamente", "No"],
        ["ID transacción", datos.id_transaccion],
        ["Token", datos.token_idempotencia],
        ["Fecha", formatearFecha(datos.fecha_transaccion)],
      ]);
    } else if (datos.previamente_realizada) {
      registrar({ tipo: "aviso", mensaje: `Token ya utilizado: el servidor NO volvió a cobrar (transacción previa ${datos.id_transaccion})` });
      // En un reintento, "ya utilizado" significa que un intento anterior SÍ pagó y solo se perdió su respuesta.
      const titulo = intento > 0
        ? "✔ Pago confirmado: un intento anterior ya lo había realizado (su respuesta se perdió) y no se cobró dos veces"
        : "⚠ Este token ya había sido utilizado: el pago no se ejecutó de nuevo";
      mostrarResultado(intento > 0 ? "ok" : "aviso", titulo, [
        ["Obtenido en", enQueIntento],
        ["Monto enviado", formatearMonto(datos.monto_transaccion)],
        ["Transacción realizada", "No"],
        ["Realizada previamente", "Sí"],
        ["Monto de la transacción previa", formatearMonto(datos.monto_transaccion_previa)],
        ["ID transacción previa", datos.id_transaccion],
        ["Fecha transacción previa", formatearFecha(datos.fecha_transaccion)],
        ["Token", datos.token_idempotencia],
      ]);
    }
    registrar({ tipo: "ok", mensaje: datos.mensaje });
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

// ---------------------------------------------------------------- Consulta de transacciones

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

  const filasTransacciones = datos.transacciones.map((t, i) =>
    el("tr", {}, [
      el("td", { texto: String(i + 1) }),
      el("td", { texto: formatearFecha(t.fecha) }),
      el("td", { texto: nombreSuscripcion(t.id_plan) }),
      el("td", { texto: formatearMonto(t.monto) }),
      el("td", {}, el("code", { texto: t.token_idempotencia })),
      el("td", {},
        el("button", {
          clase: "boton secundario mini", type: "button", "data-reenviar": "", texto: "Simular reutilizar token",
          onclick: () => reenviarConToken(t.token_idempotencia, t.monto),
        })
      ),
    ])
  );

  const tablaTokens = (tokens) =>
    tokens.length === 0
      ? el("p", { clase: "texto-tenue", texto: "Ninguno." })
      : el("div", { clase: "tabla-contenedor" },
          el("table", {}, [
            el("thead", {}, el("tr", {}, ["Token", "Suscripción", "Monto", "Generado"].map((t) => el("th", { texto: t })))),
            el("tbody", {}, tokens.map((t) =>
              el("tr", {}, [
                el("td", {}, el("code", { texto: t.token_idempotencia })),
                el("td", { texto: nombreSuscripcion(t.id_plan) }),
                el("td", { texto: formatearMonto(t.monto) }),
                el("td", { texto: formatearFecha(t.fecha_creacion) }),
              ])
            )),
          ])
        );

  contenedor.replaceChildren(
    el("p", { clase: "resumen" }, [`Cédula ${datos.cedula} · transacciones realizadas: `, el("span", { texto: String(datos.total_transacciones) })]),
    el("p", { clase: "resumen" }, ["Plan actual: ", el("span", { texto: datos.plan_actual ? datos.plan_actual.nombre : "Sin plan (gratuito)" })]),
    datos.transacciones.length === 0
      ? el("p", { clase: "texto-tenue", texto: "No hay transacciones registradas." })
      : el("div", { clase: "tabla-contenedor" },
          el("table", {}, [
            el("thead", {}, el("tr", {}, ["#", "Fecha", "Suscripción", "Monto", "Token de idempotencia", ""].map((t) => el("th", { texto: t })))),
            el("tbody", {}, filasTransacciones),
          ])
        ),
    el("h3", { texto: `Tokens de idempotencia utilizados (${datos.tokens_utilizados.length})`, style: "margin-top:1.25rem" }),
    tablaTokens(datos.tokens_utilizados),
    el("h3", { texto: `Tokens generados sin transacción (${datos.tokens_sin_utilizar.length})`, style: "margin-top:1.25rem" }),
    tablaTokens(datos.tokens_sin_utilizar)
  );
}

// ---------------------------------------------------------------- Inicio

document.addEventListener("DOMContentLoaded", () => {
  $("form-pago").addEventListener("submit", procesarPago);
  $("form-consulta").addEventListener("submit", consultarTransacciones);
  $("suscripcion").addEventListener("change", (e) => seleccionarSuscripcion(e.target.value, false));
  $("btn-recargar-suscripciones").addEventListener("click", () => {
    cargarEstadoServidor();
    cargarSuscripciones();
  });
  $("btn-reenviar").addEventListener("click", () => {
    const pago = leerUltimoPago();
    if (pago) reenviarConToken(pago.token, pago.monto);
  });

  mostrarUltimoPago();
  cargarEstadoServidor();
  cargarSuscripciones();
});
