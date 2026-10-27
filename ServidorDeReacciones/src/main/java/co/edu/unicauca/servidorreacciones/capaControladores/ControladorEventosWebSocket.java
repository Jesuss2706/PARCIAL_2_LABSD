package co.edu.unicauca.servidorreacciones.capaControladores;

import org.springframework.context.event.EventListener;
import org.springframework.messaging.simp.stomp.StompHeaderAccessor;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.messaging.SessionConnectedEvent;
import org.springframework.web.socket.messaging.SessionDisconnectEvent;
import org.springframework.web.socket.messaging.SessionSubscribeEvent;
import org.springframework.web.socket.messaging.SessionUnsubscribeEvent;

import co.edu.unicauca.servidorreacciones.capaFachada.FachadaReacciones;

/**
 * Escucha los eventos de la conexión WebSocket para registrar y des-registrar
 * a los clientes de los canales de reacciones.
 *
 * <ul>
 *   <li>SUBSCRIBE a /canalesReacciones/audio/{idCanal}: el cliente empezó a reproducir (play).</li>
 *   <li>UNSUBSCRIBE: el cliente pausó el audio.</li>
 *   <li>DISCONNECT: el cliente cerró la conexión; se eliminan todas sus suscripciones.</li>
 * </ul>
 */
@Component
public class ControladorEventosWebSocket {

	private final FachadaReacciones fachada;

	public ControladorEventosWebSocket(FachadaReacciones fachada) {
		this.fachada = fachada;
	}

	@EventListener
	public void alConectar(SessionConnectedEvent evento) {
		StompHeaderAccessor cabeceras = StompHeaderAccessor.wrap(evento.getMessage());
		System.out.printf("%n[ECO WS] Cliente CONECTADO por WebSocket (sesión %s)%n", cabeceras.getSessionId());
	}

	@EventListener
	public void alSuscribirse(SessionSubscribeEvent evento) {
		StompHeaderAccessor cabeceras = StompHeaderAccessor.wrap(evento.getMessage());
		System.out.printf("%n[ECO WS] Invocado: SUBSCRIBE %s (sesión %s)%n",
				cabeceras.getDestination(), cabeceras.getSessionId());
		fachada.registrarOyente(cabeceras.getSessionId(), cabeceras.getSubscriptionId(),
				cabeceras.getDestination(),
				cabeceras.getFirstNativeHeader("usuario"),
				cabeceras.getFirstNativeHeader("tituloAudio"));
	}

	@EventListener
	public void alDesuscribirse(SessionUnsubscribeEvent evento) {
		StompHeaderAccessor cabeceras = StompHeaderAccessor.wrap(evento.getMessage());
		System.out.printf("%n[ECO WS] Invocado: UNSUBSCRIBE %s (sesión %s)%n",
				cabeceras.getSubscriptionId(), cabeceras.getSessionId());
		fachada.eliminarOyente(cabeceras.getSessionId(), cabeceras.getSubscriptionId());
	}

	@EventListener
	public void alDesconectar(SessionDisconnectEvent evento) {
		System.out.printf("%n[ECO WS] Cliente DESCONECTADO (sesión %s)%n", evento.getSessionId());
		fachada.eliminarSesion(evento.getSessionId());
	}
}
