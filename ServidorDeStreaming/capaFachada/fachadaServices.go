// Package fachada contiene la lógica del streaming: leer el mp3 en fragmentos,
// enviarlos por gRPC y notificar la reproducción a la cola de estadísticas.
package fachada

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	capaaccesodatos "servidor.local/grpc-servidor/capaAccesoDatos"
	componenteconexioncola "servidor.local/grpc-servidor/componenteConexionCola"
	"servidor.local/grpc-servidor/dtos"
	pb "servidor.local/grpc-servidor/serviciosAudio"
)

// tamanoFragmento es el tamaño de cada chunk enviado (32 KB).
const tamanoFragmento = 32 * 1024

// FachadaStreaming agrupa la conexión a la cola usada para las estadísticas.
type FachadaStreaming struct {
	mu           sync.Mutex
	conexionCola *componenteconexioncola.RabbitPublisher
}

// NuevaFachadaStreaming crea la fachada. Si RabbitMQ no está disponible el
// streaming sigue funcionando, solo que no se envían estadísticas.
func NuevaFachadaStreaming() *FachadaStreaming {
	fmt.Println("Inicializando fachada de streaming ...")
	conexionCola, err := componenteconexioncola.NewRabbitPublisher()
	if err != nil {
		fmt.Println("Advertencia: no se pudo conectar con RabbitMQ:", err)
		conexionCola = nil
	} else {
		fmt.Println("Conectado a RabbitMQ, cola:", componenteconexioncola.NombreCola)
	}
	return &FachadaStreaming{conexionCola: conexionCola}
}

// abrirArchivo obtiene el archivo del audio usando la capa de acceso a datos.
func abrirArchivo(nombreArchivo string) (*os.File, error) {
	return capaaccesodatos.AbrirArchivo(nombreArchivo)
}

// EnviarFragmentosAudio lee el archivo en chunks y los envía al stream gRPC.
func (f *FachadaStreaming) EnviarFragmentosAudio(req *pb.AudioRequest, stream pb.AudioService_AudioStreamServer) error {
	file, err := abrirArchivo(req.Filename)
	if err != nil {
		// codes.NotFound le permite al cliente mostrar un mensaje claro
		return status.Errorf(codes.NotFound, "el audio %s no está almacenado en el servidor de streaming", req.Filename)
	}
	defer file.Close()

	// Operación asíncrona: se notifica la reproducción sin esperar al consumidor
	go f.notificarReproduccion(req)

	buf := make([]byte, tamanoFragmento)
	chunkNum := 0
	for {
		// n es la cantidad de bytes leídos; en buf queda el fragmento
		n, err := file.Read(buf)
		if err == io.EOF {
			log.Printf("Audio \"%s\" enviado completo (%d fragmentos)", req.Filename, chunkNum)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error al leer el archivo: %w", err)
		}
		if n == 0 {
			continue
		}

		chunkNum++
		if err := stream.Send(&pb.AudioChunk{Data: buf[:n]}); err != nil {
			// Ocurre cuando el cliente elige "Salir" durante la reproducción
			log.Printf("El cliente detuvo la reproducción de \"%s\" en el fragmento #%d", req.Filename, chunkNum)
			return nil
		}
		log.Printf("Chunk #%d enviado (%d bytes)", chunkNum, n)
	}
}

// notificarReproduccion publica en la cola los datos de la reproducción. Si RabbitMQ
// no estaba listo al arrancar (o se cayó), intenta reconectarse antes de publicar.
func (f *FachadaStreaming) notificarReproduccion(req *pb.AudioRequest) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.conexionCola == nil {
		conexion, err := componenteconexioncola.NewRabbitPublisher()
		if err != nil {
			fmt.Println("Advertencia: sin conexión a RabbitMQ, no se envió la estadística:", err)
			return
		}
		fmt.Println("Reconectado a RabbitMQ, cola:", componenteconexioncola.NombreCola)
		f.conexionCola = conexion
	}
	notificacion := dtos.NotificacionReproduccionDTO{
		IdAudio:   req.IdAudio,
		Titulo:    req.Titulo,
		Tipo:      req.Tipo,
		Archivo:   req.Filename,
		Cliente:   req.Cliente,
		FechaHora: time.Now().Format("2006-01-02 15:04:05"),
		Mensaje:   "Nueva reproducción: " + req.Titulo,
	}
	if err := f.conexionCola.PublicarNotificacion(notificacion); err != nil {
		fmt.Println("Error publicando la estadística (se reconectará en la próxima reproducción):", err)
		f.conexionCola.Cerrar()
		f.conexionCola = nil
	}
}
