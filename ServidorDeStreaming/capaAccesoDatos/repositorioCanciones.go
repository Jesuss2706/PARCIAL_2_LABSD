// Package capaaccesodatos abre los archivos mp3 guardados en la carpeta de audios.
package capaaccesodatos

import (
	"fmt"
	"os"
	"path/filepath"
)

// carpetaAudios es donde el servidor de audios deja los mp3 (se puede cambiar con RUTA_AUDIOS).
func carpetaAudios() string {
	if carpeta := os.Getenv("RUTA_AUDIOS"); carpeta != "" {
		return carpeta
	}
	return "audios"
}

// AbrirArchivo abre el audio indicado y lo devuelve como *os.File.
func AbrirArchivo(nombreArchivo string) (*os.File, error) {
	// filepath.Base evita que el cliente lea archivos fuera de la carpeta de audios
	ruta := filepath.Join(carpetaAudios(), filepath.Base(nombreArchivo))
	file, err := os.Open(ruta)
	if err != nil {
		fmt.Println("Error abriendo el audio:", ruta)
		return nil, fmt.Errorf("no se pudo abrir el archivo %s: %w", nombreArchivo, err)
	}
	fmt.Println("Audio abierto:", ruta)
	return file, nil
}
