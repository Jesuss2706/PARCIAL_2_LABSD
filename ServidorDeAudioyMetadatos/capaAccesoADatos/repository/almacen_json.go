package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Los audios registrados se guardan en un archivo JSON (igual que en la guía 10)
// para que sobrevivan a los reinicios del servidor.

func cargarJSON(ruta string, destino any) (bool, error) {
	contenido, err := os.ReadFile(ruta)
	if os.IsNotExist(err) || (err == nil && len(contenido) == 0) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(contenido, destino)
}

func guardarJSON(ruta string, datos any) error {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	contenido, err := json.MarshalIndent(datos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ruta, contenido, 0o644)
}
