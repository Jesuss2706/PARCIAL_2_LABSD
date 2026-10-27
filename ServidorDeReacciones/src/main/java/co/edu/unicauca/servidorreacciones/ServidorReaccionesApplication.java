package co.edu.unicauca.servidorreacciones;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

/**
 * Microservicio de reacciones: los clientes se conectan por WebSocket (STOMP sobre SockJS),
 * se registran en el canal del audio que están reproduciendo y reciben (callback) las
 * reacciones que envían los demás oyentes de ese mismo audio.
 */
@SpringBootApplication
@EnableScheduling
public class ServidorReaccionesApplication {

	public static void main(String[] args) {
		SpringApplication.run(ServidorReaccionesApplication.class, args);
	}
}
