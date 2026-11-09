package co.edu.unicauca.reacciones.filtros;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

/**
 * Igual que el filtro de los microservicios en Go: rechaza toda petición (incluido el
 * handshake del WebSocket) que no traiga la cabecera secreta que agrega el API Gateway.
 */
@Component
public class FiltroGateway extends OncePerRequestFilter {

  public static final String HEADER_GATEWAY = "X-Gateway-Token";

  @Value("${gateway.secreto}")
  private String secreto;

  @Override
  protected boolean shouldNotFilter(HttpServletRequest request) {
    return request.getRequestURI().startsWith("/actuator");
  }

  @Override
  protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
      throws ServletException, IOException {
    String token = request.getHeader(HEADER_GATEWAY);
    boolean valido = token != null && MessageDigest.isEqual(
        token.getBytes(StandardCharsets.UTF_8), secreto.getBytes(StandardCharsets.UTF_8));
    if (!valido) {
      response.setStatus(HttpServletResponse.SC_FORBIDDEN);
      response.setContentType("application/json");
      response.getWriter().write("{\"error\":\"acceso denegado: la petición debe pasar por el API Gateway\"}");
      return;
    }
    chain.doFilter(request, response);
  }
}
