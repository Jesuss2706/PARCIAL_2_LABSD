package co.edu.unicauca.servidorreacciones.capaFachada;

import java.time.LocalTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

import co.edu.unicauca.servidorreacciones.capaConfiguracion.WebSocketConfig;
import co.edu.unicauca.servidorreacciones.dtos.CanalDTO;
import co.edu.unicauca.servidorreacciones.dtos.ReaccionDTO;

/**
 * Lógica de los canales de reacciones. Cada audio tiene su propio canal; un cliente
 * se registra en el canal cuando reproduce el audio (play) y se des-registra al pausar.
 * Las reacciones se reenvían (callback) solo a los clientes registrados en ese canal,
 * usando la conexión WebSocket que ya tienen abierta.
 */
@Service
public class FachadaReacciones {

	private static final Set<String> TIPOS_VALIDOS = Set.of("like", "heart", "sad", "fun");
	private static final DateTimeFormatter FORMATO_HORA = DateTimeFormatter.ofPattern("HH:mm:ss");

	private final SimpMessagingTemplate plantillaMensajes;

	/** sesión -> (id de suscripción -> canal), para saber qué canal abandona un UNSUBSCRIBE. */
	private final Map<String, Map<String, String>> suscripcionesPorSesion = new HashMap<>();
	/** canal -> (sesión -> usuario): oyentes registrados en cada canal. */
	private final Map<String, Map<String, String>> oyentesPorCanal = new LinkedHashMap<>();
	/** canal -> título del audio. */
	private final Map<String, String> titulosPorCanal = new HashMap<>();

	public FachadaReacciones(SimpMessagingTemplate plantillaMensajes) {
		this.plantillaMensajes = plantillaMensajes;
	}

	/** Registra la sesión en el canal del audio. Retorna la cantidad de oyentes del canal. */
	public synchronized int registrarOyente(String idSesion, String idSuscripcion, String destino,
			String usuario, String tituloAudio) {
		String idCanal = extraerCanal(destino);
		if (idCanal == null) {
			return -1;
		}
		suscripcionesPorSesion.computeIfAbsent(idSesion, s -> new HashMap<>()).put(idSuscripcion, idCanal);
		oyentesPorCanal.computeIfAbsent(idCanal, c -> new LinkedHashMap<>()).put(idSesion, nombreUsuario(usuario));
		if (tituloAudio != null && !tituloAudio.isBlank()) {
			titulosPorCanal.put(idCanal, tituloAudio);
		}
		int oyentes = oyentesPorCanal.get(idCanal).size();
		System.out.printf("[ECO WS] %s REGISTRADO en el canal \"%s\" (audio: %s) · oyentes: %d%n",
				nombreUsuario(usuario), idCanal, titulosPorCanal.getOrDefault(idCanal, idCanal), oyentes);
		return oyentes;
	}

	/** Des-registra una suscripción concreta (el cliente pausó el audio). */
	public synchronized void eliminarOyente(String idSesion, String idSuscripcion) {
		Map<String, String> suscripciones = suscripcionesPorSesion.get(idSesion);
		if (suscripciones == null) {
			return;
		}
		String idCanal = suscripciones.remove(idSuscripcion);
		if (idCanal != null) {
			quitarDeCanal(idSesion, idCanal, "DES-REGISTRADO");
		}
	}

	/** Elimina todas las suscripciones de una sesión (el cliente cerró la conexión). */
	public synchronized void eliminarSesion(String idSesion) {
		Map<String, String> suscripciones = suscripcionesPorSesion.remove(idSesion);
		if (suscripciones == null) {
			return;
		}
		for (String idCanal : suscripciones.values()) {
			quitarDeCanal(idSesion, idCanal, "DESCONECTADO");
		}
	}

	/**
	 * Reenvía la reacción a todos los clientes registrados en el canal del audio.
	 * Retorna la cantidad de clientes a los que se envió, o -1 si la reacción no es válida.
	 */
	public int reenviarReaccion(String idCanal, ReaccionDTO reaccion, String idSesion) {
		if (reaccion == null || reaccion.getTipo() == null || !TIPOS_VALIDOS.contains(reaccion.getTipo())) {
			System.out.printf("[ECO WS] Reacción inválida descartada en el canal \"%s\"%n", idCanal);
			return -1;
		}
		int oyentes;
		synchronized (this) {
			Map<String, String> oyentesCanal = oyentesPorCanal.getOrDefault(idCanal, Map.of());
			oyentes = oyentesCanal.size();
			if (reaccion.getUsuario() == null || reaccion.getUsuario().isBlank()) {
				reaccion.setUsuario(oyentesCanal.getOrDefault(idSesion, "Anónimo"));
			}
			if (reaccion.getTituloAudio() == null || reaccion.getTituloAudio().isBlank()) {
				reaccion.setTituloAudio(titulosPorCanal.getOrDefault(idCanal, idCanal));
			}
		}
		reaccion.setIdCanal(idCanal);
		reaccion.setFechaHora(LocalTime.now().format(FORMATO_HORA));

		System.out.printf("[ECO WS] Reacción \"%s\" de %s en \"%s\" -> reenviando a %d cliente(s) del canal%n",
				reaccion.getTipo(), reaccion.getUsuario(), reaccion.getTituloAudio(), oyentes);
		// Callback: el broker entrega el mensaje por la conexión WebSocket de cada suscriptor.
		plantillaMensajes.convertAndSend(WebSocketConfig.PREFIJO_CANAL_AUDIO + idCanal, reaccion);
		return oyentes;
	}

	/** Estado actual de los canales (para el servicio REST de consulta). */
	public synchronized List<CanalDTO> listarCanales() {
		List<CanalDTO> canales = new ArrayList<>();
		oyentesPorCanal.forEach((idCanal, oyentes) -> canales.add(new CanalDTO(idCanal,
				titulosPorCanal.getOrDefault(idCanal, idCanal), oyentes.size(), new ArrayList<>(oyentes.values()))));
		return canales;
	}

	private void quitarDeCanal(String idSesion, String idCanal, String motivo) {
		Map<String, String> oyentes = oyentesPorCanal.get(idCanal);
		if (oyentes == null) {
			return;
		}
		String usuario = oyentes.remove(idSesion);
		System.out.printf("[ECO WS] %s %s del canal \"%s\" · oyentes: %d%n",
				usuario == null ? idSesion : usuario, motivo, idCanal, oyentes.size());
		if (oyentes.isEmpty()) {
			oyentesPorCanal.remove(idCanal);
			titulosPorCanal.remove(idCanal);
		}
	}

	private static String extraerCanal(String destino) {
		if (destino == null || !destino.startsWith(WebSocketConfig.PREFIJO_CANAL_AUDIO)) {
			return null;
		}
		String idCanal = destino.substring(WebSocketConfig.PREFIJO_CANAL_AUDIO.length());
		return idCanal.isBlank() ? null : idCanal;
	}

	private static String nombreUsuario(String usuario) {
		return usuario == null || usuario.isBlank() ? "Anónimo" : usuario;
	}
}
