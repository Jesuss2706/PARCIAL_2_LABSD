package co.edu.unicauca.servidorreacciones.capaConfiguracion;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

/**
 * Rechaza toda petición (HTTP o handshake WebSocket) que no haya pasado por el API Gateway.
 * El gateway agrega la cabecera X-Gateway-Token con un secreto compartido.
 */
@Component
public class FiltroGateway extends OncePerRequestFilter {

	public static final String CABECERA_GATEWAY = "X-Gateway-Token";

	private final String secreto;
	private final boolean exigirToken;

	public FiltroGateway(@Value("${gateway.secreto}") String secreto,
			@Value("${gateway.exigir-token}") boolean exigirToken) {
		this.secreto = secreto;
		this.exigirToken = exigirToken;
	}

	@Override
	protected void doFilterInternal(HttpServletRequest peticion, HttpServletResponse respuesta, FilterChain cadena)
			throws ServletException, IOException {
		if (!exigirToken || peticion.getRequestURI().startsWith("/actuator") || tokenValido(peticion)) {
			cadena.doFilter(peticion, respuesta);
			return;
		}
		System.out.printf("[ECO HTTP] Petición %s %s RECHAZADA: no pasó por el API Gateway%n",
				peticion.getMethod(), peticion.getRequestURI());
		respuesta.setStatus(HttpServletResponse.SC_FORBIDDEN);
		respuesta.setContentType("application/json;charset=UTF-8");
		respuesta.getWriter().write("{\"error\":\"acceso denegado: la petición debe pasar por el API Gateway\"}");
	}

	private boolean tokenValido(HttpServletRequest peticion) {
		String token = peticion.getHeader(CABECERA_GATEWAY);
		return token != null && MessageDigest.isEqual(
				token.getBytes(StandardCharsets.UTF_8), secreto.getBytes(StandardCharsets.UTF_8));
	}
}
