package co.edu.unicauca.reacciones.capaConfiguracion;

import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.simp.config.MessageBrokerRegistry;
import org.springframework.web.socket.config.annotation.*;

/**
 * Bróker STOMP en memoria. Cada audio tiene su propio canal:
 *   - el cliente se suscribe a  /brokerDeReacciones/audio/{idAudio}   (al reproducir)
 *   - y envía sus reacciones a  /app/enviarReaccion/{idAudio}
 */
@Configuration
@EnableWebSocketMessageBroker
public class WebSocketConfig implements WebSocketMessageBrokerConfigurer {

  @Override
  public void configureMessageBroker(MessageBrokerRegistry config) {
    config.enableSimpleBroker("/brokerDeReacciones");
    config.setApplicationDestinationPrefixes("/app");
  }

  @Override
  public void registerStompEndpoints(StompEndpointRegistry registry) {
    // WebSocket nativo (lo usa el cliente web a través del API Gateway)
    registry.addEndpoint("/ws-stomp").setAllowedOriginPatterns("*");
    // SockJS (igual que en la guía 8), por compatibilidad
    registry.addEndpoint("/ws").setAllowedOriginPatterns("*").withSockJS();
  }
}
