package co.edu.unicauca.reacciones.dtos;

/** Aviso que el servidor publica en el canal de un audio cuando alguien entra o sale. */
public class AvisoCanalDTO {
  private final String tipo = "AVISO";
  private final int idAudio;
  private final String usuario;
  private final String evento;   // "ENTRO" o "SALIO"
  private final int oyentes;     // clientes registrados en el canal en este momento

  public AvisoCanalDTO(int idAudio, String usuario, String evento, int oyentes) {
    this.idAudio = idAudio;
    this.usuario = usuario;
    this.evento = evento;
    this.oyentes = oyentes;
  }

  public String getTipo() { return tipo; }
  public int getIdAudio() { return idAudio; }
  public String getUsuario() { return usuario; }
  public String getEvento() { return evento; }
  public int getOyentes() { return oyentes; }
}
