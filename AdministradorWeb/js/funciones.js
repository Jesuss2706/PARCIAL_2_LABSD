// Administrador web: registra un nuevo audio enviando sus metadatos y el mp3.
// Cada tipo de audio tiene sus propios metadatos; el formulario se arma según el tipo elegido.

const $ = (id) => document.getElementById(id);
const estado = { tipos: [], metadatos: [] };

function el(etiqueta, atributos = {}, hijos = []) {
  const nodo = document.createElement(etiqueta);
  for (const [clave, valor] of Object.entries(atributos)) {
    if (clave === "texto") nodo.textContent = valor;
    else if (clave === "clase") nodo.className = valor;
    else if (clave.startsWith("on")) nodo.addEventListener(clave.slice(2), valor);
    else nodo.setAttribute(clave, valor);
  }
  for (const hijo of [].concat(hijos)) if (hijo != null) nodo.append(hijo);
  return nodo;
}

function eco(origen, mensaje, tipo = "info") {
  console.log(`[ECO ${origen}] ${mensaje}`);
  const item = el("li", { clase: tipo }, [
    el("span", { clase: "hora", texto: new Date().toLocaleTimeString("es-CO") }),
    el("span", { clase: "origen", texto: origen }),
    mensaje,
  ]);
  $("ecos").prepend(item);
}

async function obtenerJSON(ruta) {
  const respuesta = await fetch(CONFIG.URL_API + ruta);
  const datos = await respuesta.json().catch(() => ({}));
  if (!respuesta.ok) throw new Error(datos.error || `HTTP ${respuesta.status}`);
  return datos;
}

// ---------------------------------------------------------------- tipos y metadatos

async function cargarTipos() {
  eco("REST", "GET /api/tipos -> listar tipos de audio");
  try {
    estado.tipos = await obtenerJSON("/tipos");
    $("estado-servidor").className = "indicador conectado";
    $("estado-servidor").textContent = "API Gateway y servidor de audios activos";
  } catch (error) {
    $("estado-servidor").className = "indicador desconectado";
    $("estado-servidor").textContent = "Servidor no disponible";
    $("tipo").replaceChildren(el("option", { value: "", texto: "No disponible" }));
    eco("REST", `No fue posible listar los tipos: ${error.message}`, "error");
    return;
  }
  $("tipo").replaceChildren(
    el("option", { value: "", texto: "Seleccione…" }),
    ...estado.tipos.map((t) => el("option", { value: t.id, texto: t.nombre }))
  );
  $("filtro-tipo").replaceChildren(...estado.tipos.map((t) => el("option", { value: t.id, texto: t.nombre })));
  cargarAudios();
}

async function cargarMetadatos() {
  const idTipo = $("tipo").value;
  const contenedor = $("campos-metadatos");
  const leyenda = el("legend", { texto: "Metadatos propios del tipo" });
  if (!idTipo) {
    estado.metadatos = [];
    contenedor.replaceChildren(leyenda, el("p", { clase: "texto-tenue", texto: "Seleccione un tipo de audio." }));
    return;
  }
  eco("REST", `GET /api/tipos/${idTipo}/metadatos -> metadatos propios del tipo`);
  try {
    estado.metadatos = await obtenerJSON(`/tipos/${idTipo}/metadatos`);
  } catch (error) {
    contenedor.replaceChildren(leyenda, el("p", { clase: "texto-tenue", texto: error.message }));
    return;
  }
  const tipo = estado.tipos.find((t) => String(t.id) === idTipo);
  leyenda.textContent = `Metadatos de ${tipo?.nombre ?? "tipo"}`;
  contenedor.replaceChildren(
    leyenda,
    ...estado.metadatos.map((m) =>
      el("label", {}, [
        m.nombre,
        el("input", { type: "text", "data-metadato": m.id, required: "", maxlength: "120", placeholder: m.descripcion }),
      ])
    )
  );
  $("filtro-tipo").value = idTipo;
  cargarAudios();
}

// ---------------------------------------------------------------- archivo

function formatearDuracion(segundos) {
  if (!isFinite(segundos)) return "";
  const s = Math.round(segundos);
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const r = String(s % 60).padStart(2, "0");
  return h > 0 ? `${h}:${String(m).padStart(2, "0")}:${r}` : `${m}:${r}`;
}

function alElegirArchivo() {
  const archivo = $("archivo").files[0];
  $("info-archivo").textContent = "";
  if (!archivo) return;
  $("info-archivo").textContent = `${archivo.name} · ${(archivo.size / 1048576).toFixed(2)} MB`;
  if (!$("titulo").value.trim()) $("titulo").value = archivo.name.replace(/\.mp3$/i, "").replace(/[_-]+/g, " ");
  // La duración se lee del propio mp3 con un elemento <audio> temporal
  const url = URL.createObjectURL(archivo);
  const audio = new Audio();
  audio.preload = "metadata";
  audio.onloadedmetadata = () => {
    $("duracion").value = formatearDuracion(audio.duration);
    URL.revokeObjectURL(url);
  };
  audio.onerror = () => URL.revokeObjectURL(url);
  audio.src = url;
}

// ---------------------------------------------------------------- registro

function mostrarResultado(tipo, titulo, filas = []) {
  const caja = $("resultado");
  caja.className = `resultado ${tipo}`;
  caja.replaceChildren(el("h3", { texto: titulo }), el("dl", {}, filas.flatMap(([k, v]) => [el("dt", { texto: k }), el("dd", { texto: v })])));
}

function registrarAudio(evento) {
  evento.preventDefault();
  const archivo = $("archivo").files[0];
  if (!archivo) return;
  if (!/\.mp3$/i.test(archivo.name)) {
    mostrarResultado("error", "El archivo debe ser .mp3");
    return;
  }

  const metadatos = {};
  document.querySelectorAll("[data-metadato]").forEach((input) => (metadatos[input.dataset.metadato] = input.value.trim()));

  const formulario = new FormData();
  formulario.append("idTipo", $("tipo").value);
  formulario.append("titulo", $("titulo").value.trim());
  formulario.append("duracion", $("duracion").value.trim());
  formulario.append("metadatos", JSON.stringify(metadatos));
  formulario.append("archivo", archivo);

  const tipo = estado.tipos.find((t) => String(t.id) === $("tipo").value);
  eco("REST", `POST /api/audios -> registrar "${$("titulo").value.trim()}" (${tipo?.nombre}) con ${Object.keys(metadatos).length} metadatos y ${archivo.name}`);

  // XMLHttpRequest permite mostrar el progreso de subida del mp3
  const xhr = new XMLHttpRequest();
  xhr.open("POST", CONFIG.URL_API + "/audios");
  $("progreso").classList.remove("oculto");
  $("btn-registrar").disabled = true;
  $("resultado").className = "resultado oculto";

  xhr.upload.onprogress = (e) => {
    if (!e.lengthComputable) return;
    const porcentaje = Math.round((e.loaded / e.total) * 100);
    $("barra").style.width = `${porcentaje}%`;
    $("texto-progreso").textContent = `Subiendo… ${porcentaje}%`;
  };
  xhr.onload = () => {
    $("btn-registrar").disabled = false;
    let datos = {};
    try {
      datos = JSON.parse(xhr.responseText);
    } catch {
      /* respuesta sin JSON */
    }
    if (xhr.status === 201) {
      $("texto-progreso").textContent = "Subida completa";
      eco("REST", `Audio #${datos.id} registrado: ${datos.nombreArchivo}`, "ok");
      mostrarResultado("ok", `✔ Audio registrado (#${datos.id})`, [
        ["Título", datos.titulo],
        ["Tipo", datos.tipo],
        ...datos.metadatos.map((m) => [m.nombre, m.valor]),
        ["Archivo en el servidor de streaming", datos.nombreArchivo],
        ["Duración", datos.duracion],
      ]);
      const idTipo = $("tipo").value;
      $("form-audio").reset();
      $("tipo").value = idTipo; // se conserva el tipo para registrar otro audio y refrescar la lista
      $("info-archivo").textContent = "";
      cargarMetadatos();
    } else {
      eco("REST", `Registro rechazado (HTTP ${xhr.status}): ${datos.error || xhr.statusText}`, "error");
      mostrarResultado("error", "✖ No fue posible registrar el audio", [["Detalle", datos.error || `HTTP ${xhr.status}`]]);
    }
  };
  xhr.onerror = () => {
    $("btn-registrar").disabled = false;
    eco("REST", "No fue posible conectarse con el API Gateway", "error");
    mostrarResultado("error", "✖ No fue posible conectarse con el API Gateway");
  };
  xhr.send(formulario);
}

// ---------------------------------------------------------------- audios registrados

async function cargarAudios() {
  const idTipo = $("filtro-tipo").value;
  const contenedor = $("lista-audios");
  if (!idTipo) return;
  try {
    const audios = await obtenerJSON(`/tipos/${idTipo}/audios`);
    if (audios.length === 0) {
      contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: "No hay audios de este tipo." }));
      return;
    }
    contenedor.replaceChildren(
      el("div", { clase: "tabla-contenedor" },
        el("table", {}, [
          el("thead", {}, el("tr", {}, ["#", "Título", "Duración"].map((t) => el("th", { texto: t })))),
          el("tbody", {}, audios.map((a) => el("tr", {}, [
            el("td", { texto: String(a.id) }), el("td", { texto: a.titulo }), el("td", { texto: a.duracion }),
          ]))),
        ])
      )
    );
  } catch (error) {
    contenedor.replaceChildren(el("p", { clase: "texto-tenue", texto: error.message }));
  }
}

document.addEventListener("DOMContentLoaded", () => {
  $("tipo").addEventListener("change", cargarMetadatos);
  $("archivo").addEventListener("change", alElegirArchivo);
  $("form-audio").addEventListener("submit", registrarAudio);
  $("filtro-tipo").addEventListener("change", cargarAudios);
  $("btn-limpiar-ecos").addEventListener("click", () => $("ecos").replaceChildren());
  cargarTipos();
});
