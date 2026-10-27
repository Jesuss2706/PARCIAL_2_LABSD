package co.edu.unicauca.servidorreacciones.capaControladores;

import org.springframework.messaging.handler.annotation.DestinationVariable;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.stereotype.Controller;

import co.edu.unicauca.servidorreacciones.capaFachada.FachadaReacciones;
import co.edu.unicauca.servidorreacciones.dtos.ReaccionDTO;

/**
 * Recibe las reacciones que los clientes envían a /app/enviarReaccion/{idCanal}
 * y las reenvía a los oyentes del mismo audio.
 */
@Controller
public class ReaccionesController {

	private final FachadaReacciones fachada;

	public ReaccionesController(FachadaReacciones fachada) {
		this.fachada = fachada;
	}

	@MessageMapping("/enviarReaccion/{idCanal}")
	public void recibirReaccion(@DestinationVariable String idCanal, @Payload ReaccionDTO reaccion,
			SimpMessageHeaderAccessor cabeceras) {
		System.out.printf("%n[ECO WS] Invocado: SEND /app/enviarReaccion/%s (sesión %s)%n",
				idCanal, cabeceras.getSessionId());
		fachada.reenviarReaccion(idCanal, reaccion, cabeceras.getSessionId());
	}
}
