package co.edu.unicauca.reacciones.dtos;

/** Reacción que un cliente envía sobre el audio que está reproduciendo. */
public class ReaccionDTO {
  private int idAudio;
  private String tituloAudio;
  private String usuario;
  private String reaccion;   // p. ej. "❤️", "🔥", "👏" o un texto corto
  private String fechaHora;

  public ReaccionDTO() {}

  public int getIdAudio() { return idAudio; }
  public void setIdAudio(int idAudio) { this.idAudio = idAudio; }
  public String getTituloAudio() { return tituloAudio; }
  public void setTituloAudio(String tituloAudio) { this.tituloAudio = tituloAudio; }
  public String getUsuario() { return usuario; }
  public void setUsuario(String usuario) { this.usuario = usuario; }
  public String getReaccion() { return reaccion; }
  public void setReaccion(String reaccion) { this.reaccion = reaccion; }
  public String getFechaHora() { return fechaHora; }
  public void setFechaHora(String fechaHora) { this.fechaHora = fechaHora; }
}
