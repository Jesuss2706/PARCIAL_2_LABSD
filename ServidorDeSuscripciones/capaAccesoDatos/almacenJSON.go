// Package capaAccesoDatos persiste planes, tokens y transacciones.
// Tokens y transacciones se guardan en archivos JSON para que el historial
// sobreviva a los reinicios del servidor entre un caso de prueba y otro.
package capaAccesoDatos

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func cargarJSON(ruta string, destino any) error {
	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) || (err == nil && len(contenido) == 0) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(contenido, destino)
}

// guardarJSON escribe primero en un archivo temporal y luego lo renombra,
// para no dejar el archivo a medio escribir si el proceso se interrumpe.
func guardarJSON(ruta string, datos any) error {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	contenido, err := json.MarshalIndent(datos, "", "  ")
	if err != nil {
		return err
	}
	temporal := ruta + ".tmp"
	if err := os.WriteFile(temporal, contenido, 0o644); err != nil {
		return err
	}
	return os.Rename(temporal, ruta)
}
