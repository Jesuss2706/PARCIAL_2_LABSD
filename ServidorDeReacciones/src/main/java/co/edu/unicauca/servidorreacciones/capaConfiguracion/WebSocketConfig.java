package co.edu.unicauca.servidorreacciones.capaConfiguracion;

import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.simp.config.MessageBrokerRegistry;
import org.springframework.web.socket.config.annotation.EnableWebSocketMessageBroker;
import org.springframework.web.socket.config.annotation.StompEndpointRegistry;
import org.springframework.web.socket.config.annotation.WebSocketMessageBrokerConfigurer;

/**
 * Configuración del broker STOMP.
 *
 * <ul>
 *   <li>Endpoint de conexión: /reacciones/ws (SockJS). Por el gateway: /api/reacciones/ws</li>
 *   <li>Canales (uno por audio): /canalesReacciones/audio/{idCanal}</li>
 *   <li>Envío de reacciones: /app/enviarReaccion/{idCanal}</li>
 * </ul>
 */
@Configuration
@EnableWebSocketMessageBroker
public class WebSocketConfig implements WebSocketMessageBrokerConfigurer {

	public static final String PREFIJO_CANALES = "/canalesReacciones";
	public static final String PREFIJO_CANAL_AUDIO = PREFIJO_CANALES + "/audio/";

	@Override
	public void configureMessageBroker(MessageBrokerRegistry config) {
		config.enableSimpleBroker(PREFIJO_CANALES);
		config.setApplicationDestinationPrefixes("/app");
	}

	@Override
	public void registerStompEndpoints(StompEndpointRegistry registry) {
		registry.addEndpoint("/reacciones/ws")
				.setAllowedOriginPatterns("*")
				.withSockJS();
	}
}
