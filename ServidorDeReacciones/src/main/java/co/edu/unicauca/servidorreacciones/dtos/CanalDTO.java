package co.edu.unicauca.servidorreacciones.dtos;

import java.util.List;

/** Estado de un canal de reacciones: audio y oyentes registrados. */
public class CanalDTO {

	private final String idCanal;
	private final String tituloAudio;
	private final int cantidadOyentes;
	private final List<String> usuarios;

	public CanalDTO(String idCanal, String tituloAudio, int cantidadOyentes, List<String> usuarios) {
		this.idCanal = idCanal;
		this.tituloAudio = tituloAudio;
		this.cantidadOyentes = cantidadOyentes;
		this.usuarios = usuarios;
	}

	public String getIdCanal() { return idCanal; }
	public String getTituloAudio() { return tituloAudio; }
	public int getCantidadOyentes() { return cantidadOyentes; }
	public List<String> getUsuarios() { return usuarios; }
}
