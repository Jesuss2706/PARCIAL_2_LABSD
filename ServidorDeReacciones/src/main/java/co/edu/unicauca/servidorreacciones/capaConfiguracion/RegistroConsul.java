package co.edu.unicauca.servidorreacciones.capaConfiguracion;

import java.util.List;
import java.util.Map;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.http.MediaType;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import jakarta.annotation.PreDestroy;

/**
 * Registra el microservicio en el directorio Consul (API HTTP) con un health check TTL
 * que se renueva cada 5 segundos. Si Consul no está disponible, se reintenta.
 */
@Component
public class RegistroConsul {

	private final RestClient clienteConsul;
	private final String nombreServicio;
	private final String hostServicio;
	private final int puerto;
	private final String direccionConsul;
	private final String idServicio;
	private final String idCheck;
	private volatile boolean registrado = false;

	public RegistroConsul(@Value("${consul.direccion}") String direccionConsul,
			@Value("${consul.host-servicio}") String hostServicio,
			@Value("${spring.application.name}") String nombreServicio,
			@Value("${server.port}") int puerto) {
		this.direccionConsul = direccionConsul;
		this.hostServicio = hostServicio;
		this.nombreServicio = nombreServicio;
		this.puerto = puerto;
		this.idServicio = nombreServicio + "-" + hostServicio + "-" + puerto;
		this.idCheck = "check-" + idServicio;
		this.clienteConsul = RestClient.builder().baseUrl("http://" + direccionConsul).build();
	}

	@EventListener(ApplicationReadyEvent.class)
	public void alIniciar() {
		registrar();
	}

	/** Renueva el TTL; si el registro se perdió (Consul reiniciado), se vuelve a registrar. */
	@Scheduled(fixedDelay = 5000, initialDelay = 5000)
	public void latido() {
		if (!registrado) {
			registrar();
			return;
		}
		try {
			clienteConsul.put().uri("/v1/agent/check/pass/{id}", idCheck).retrieve().toBodilessEntity();
		} catch (Exception e) {
			System.out.printf("[CONSUL] Error renovando el TTL: %s%n", e.getMessage());
			registrado = false;
		}
	}

	private void registrar() {
		Map<String, Object> registro = Map.of(
				"ID", idServicio,
				"Name", nombreServicio,
				"Address", hostServicio,
				"Port", puerto,
				"Tags", List.of("spring", "websocket", "reacciones"),
				"Check", Map.of(
						"CheckID", idCheck,
						"TTL", "15s",
						"DeregisterCriticalServiceAfter", "1m"));
		try {
			clienteConsul.put().uri("/v1/agent/service/register")
					.contentType(MediaType.APPLICATION_JSON)
					.body(registro)
					.retrieve().toBodilessEntity();
			clienteConsul.put().uri("/v1/agent/check/pass/{id}", idCheck).retrieve().toBodilessEntity();
			registrado = true;
			System.out.printf("[CONSUL] Servicio \"%s\" registrado en Consul (%s) como %s%n",
					nombreServicio, direccionConsul, idServicio);
		} catch (Exception e) {
			System.out.printf("[CONSUL] No fue posible registrarse en %s: %s (se reintenta en 5 s)%n",
					direccionConsul, e.getMessage());
		}
	}

	@PreDestroy
	public void desregistrar() {
		if (!registrado) {
			return;
		}
		try {
			clienteConsul.put().uri("/v1/agent/service/deregister/{id}", idServicio).retrieve().toBodilessEntity();
			System.out.printf("[CONSUL] Servicio %s desregistrado de Consul%n", idServicio);
		} catch (Exception e) {
			System.out.printf("[CONSUL] Error desregistrando: %s%n", e.getMessage());
		}
	}
}
