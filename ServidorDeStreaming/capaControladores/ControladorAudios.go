// Package capacontroladores implementa el servicio gRPC definido en audio.proto.
package capacontroladores

import (
	"fmt"

	capafachada "servidor.local/grpc-servidor/capaFachada"
	pb "servidor.local/grpc-servidor/serviciosAudio"
)

// ControladorServidor atiende los procedimientos remotos de AudioService.
type ControladorServidor struct {
	pb.UnimplementedAudioServiceServer
	fachada *capafachada.FachadaStreaming
}

// NuevoControladorServidor crea el controlador con su fachada.
func NuevoControladorServidor() *ControladorServidor {
	return &ControladorServidor{fachada: capafachada.NuevaFachadaStreaming()}
}

// AudioStream es la implementación del procedimiento remoto.
func (s *ControladorServidor) AudioStream(req *pb.AudioRequest, stream pb.AudioService_AudioStreamServer) error {
	fmt.Printf("\n[ECO gRPC] Invocado: AudioStream(filename=\"%s\", titulo=\"%s\", tipo=\"%s\", cliente=\"%s\")\n",
		req.Filename, req.Titulo, req.Tipo, req.Cliente)

	// Delegar la lectura y el envío de fragmentos a la fachada
	return s.fachada.EnviarFragmentosAudio(req, stream)
}
