package co.edu.unicauca.reacciones.capaFachada;

import co.edu.unicauca.reacciones.dtos.AvisoCanalDTO;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.springframework.context.event.EventListener;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.messaging.simp.stomp.StompHeaderAccessor;
import org.springframework.stereotype.Service;
import org.springframework.web.socket.messaging.SessionConnectedEvent;
import org.springframework.web.socket.messaging.SessionDisconnectEvent;
import org.springframework.web.socket.messaging.SessionSubscribeEvent;
import org.springframework.web.socket.messaging.SessionUnsubscribeEvent;

/**
 * Lleva el registro de qué cliente está suscrito al canal de qué audio.
 * El cliente se registra (SUBSCRIBE) cuando reproduce y se des-registra (UNSUBSCRIBE)
 * cuando pausa; si cierra la pestaña, el DISCONNECT lo saca de todos los canales.
 */
@Service
public class GestorCanales {

  public static final String PREFIJO_CANAL = "/brokerDeReacciones/audio/";
  private static final Pattern CANAL = Pattern.compile("^/brokerDeReacciones/audio/(\\d+)$");

  /** Una suscripción viva: sesión + id de suscripción STOMP. */
  private record Suscripcion(int idAudio, String usuario) {}

  // clave: idSesion + "|" + idSuscripcion
  private final Map<String, Suscripcion> suscripciones = new ConcurrentHashMap<>();
  private final SimpMessagingTemplate mensajeria;

  public GestorCanales(SimpMessagingTemplate mensajeria) {
    this.mensajeria = mensajeria;
  }

  @EventListener
  public void alConectar(SessionConnectedEvent evento) {
    StompHeaderAccessor h = StompHeaderAccessor.wrap(evento.getMessage());
    System.out.printf("%n[ECO WS] Nueva conexión WebSocket (sesión %s)%n", h.getSessionId());
  }

  @EventListener
  public void alSuscribir(SessionSubscribeEvent evento) {
    StompHeaderAccessor h = StompHeaderAccessor.wrap(evento.getMessage());
    String destino = h.getDestination();
    Matcher m = destino == null ? null : CANAL.matcher(destino);
    if (m == null || !m.matches()) {
      return;
    }
    int idAudio = Integer.parseInt(m.group(1));
    String usuario = valorONulo(h.getFirstNativeHeader("usuario"), "anónimo");
    suscripciones.put(clave(h.getSessionId(), h.getSubscriptionId()), new Suscripcion(idAudio, usuario));
    int oyentes = contarOyentes(idAudio);
    System.out.printf("[ECO WS] %s se REGISTRÓ en el canal del audio %d (oyentes: %d)%n", usuario, idAudio, oyentes);
    avisar(idAudio, usuario, "ENTRO", oyentes);
  }

  @EventListener
  public void alDesuscribir(SessionUnsubscribeEvent evento) {
    StompHeaderAccessor h = StompHeaderAccessor.wrap(evento.getMessage());
    Suscripcion s = suscripciones.remove(clave(h.getSessionId(), h.getSubscriptionId()));
    if (s != null) {
      int oyentes = contarOyentes(s.idAudio());
      System.out.printf("[ECO WS] %s se DES-REGISTRÓ del canal del audio %d (oyentes: %d)%n", s.usuario(), s.idAudio(), oyentes);
      avisar(s.idAudio(), s.usuario(), "SALIO", oyentes);
    }
  }

  @EventListener
  public void alDesconectar(SessionDisconnectEvent evento) {
    String prefijo = evento.getSessionId() + "|";
    suscripciones.entrySet().removeIf(e -> {
      if (!e.getKey().startsWith(prefijo)) {
        return false;
      }
      Suscripcion s = e.getValue();
      int oyentes = (int) suscripciones.entrySet().stream()
          .filter(o -> !o.getKey().startsWith(prefijo) && o.getValue().idAudio() == s.idAudio()).count();
      System.out.printf("[ECO WS] %s se desconectó: sale del canal del audio %d (oyentes: %d)%n", s.usuario(), s.idAudio(), oyentes);
      avisar(s.idAudio(), s.usuario(), "SALIO", oyentes);
      return true;
    });
    System.out.printf("[ECO WS] Conexión WebSocket cerrada (sesión %s)%n", evento.getSessionId());
  }

  public int contarOyentes(int idAudio) {
    return (int) suscripciones.values().stream().filter(s -> s.idAudio() == idAudio).count();
  }

  private void avisar(int idAudio, String usuario, String evento, int oyentes) {
    mensajeria.convertAndSend(PREFIJO_CANAL + idAudio, new AvisoCanalDTO(idAudio, usuario, evento, oyentes));
  }

  private static String clave(String sesion, String suscripcion) {
    return sesion + "|" + suscripcion;
  }

  private static String valorONulo(String valor, String porDefecto) {
    return valor == null || valor.isBlank() ? porDefecto : valor;
  }
}
