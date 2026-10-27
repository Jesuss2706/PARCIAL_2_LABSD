package co.edu.unicauca.servidorreacciones.dtos;

/**
 * Reacción que un cliente envía sobre el audio que está reproduciendo.
 * Viaja en formato JSON por el WebSocket.
 */
public class ReaccionDTO {

	private String idCanal;
	private String tituloAudio;
	private String usuario;
	private String tipo; // like, heart, sad, fun
	private String fechaHora;

	public ReaccionDTO() {
	}

	public String getIdCanal() { return idCanal; }
	public void setIdCanal(String idCanal) { this.idCanal = idCanal; }

	public String getTituloAudio() { return tituloAudio; }
	public void setTituloAudio(String tituloAudio) { this.tituloAudio = tituloAudio; }

	public String getUsuario() { return usuario; }
	public void setUsuario(String usuario) { this.usuario = usuario; }

	public String getTipo() { return tipo; }
	public void setTipo(String tipo) { this.tipo = tipo; }

	public String getFechaHora() { return fechaHora; }
	public void setFechaHora(String fechaHora) { this.fechaHora = fechaHora; }
}
