package co.edu.unicauca.servidorreacciones.capaControladores;

import java.util.List;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import co.edu.unicauca.servidorreacciones.capaFachada.FachadaReacciones;
import co.edu.unicauca.servidorreacciones.dtos.CanalDTO;

/** Servicio REST para consultar qué canales están activos y quién está registrado. */
@RestController
@RequestMapping("/reacciones")
public class ControladorCanales {

	private final FachadaReacciones fachada;

	public ControladorCanales(FachadaReacciones fachada) {
		this.fachada = fachada;
	}

	@GetMapping("/canales")
	public List<CanalDTO> listarCanales() {
		List<CanalDTO> canales = fachada.listarCanales();
		System.out.printf("%n[ECO REST] Invocado: GET /reacciones/canales -> %d canal(es) activo(s)%n", canales.size());
		return canales;
	}
}
