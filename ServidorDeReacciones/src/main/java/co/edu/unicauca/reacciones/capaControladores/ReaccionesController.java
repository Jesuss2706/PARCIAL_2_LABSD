package co.edu.unicauca.reacciones.capaControladores;

import co.edu.unicauca.reacciones.capaFachada.GestorCanales;
import co.edu.unicauca.reacciones.dtos.ReaccionDTO;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import org.springframework.messaging.handler.annotation.DestinationVariable;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.SendTo;
import org.springframework.stereotype.Controller;

@Controller
public class ReaccionesController {

  private static final DateTimeFormatter FORMATO = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
  private final GestorCanales gestorCanales;

  public ReaccionesController(GestorCanales gestorCanales) {
    this.gestorCanales = gestorCanales;
  }

  /**
   * El cliente envía la reacción a /app/enviarReaccion/{idAudio}; el servidor la reenvía
   * (callback) por la conexión ya abierta a todos los suscritos al canal de ESE audio.
   */
  @MessageMapping("/enviarReaccion/{idAudio}")
  @SendTo("/brokerDeReacciones/audio/{idAudio}")
  public ReaccionDTO procesarReaccion(@DestinationVariable int idAudio, ReaccionDTO reaccion) {
    reaccion.setIdAudio(idAudio);
    reaccion.setFechaHora(LocalDateTime.now().format(FORMATO));
    System.out.printf("[ECO WS] Reacción '%s' de %s en el audio %d (\"%s\") -> reenviada a %d oyentes%n",
        reaccion.getReaccion(), reaccion.getUsuario(), idAudio, reaccion.getTituloAudio(),
        gestorCanales.contarOyentes(idAudio));
    return reaccion;
  }
}
