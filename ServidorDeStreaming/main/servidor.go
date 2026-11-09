// Punto de entrada del Servidor de Streaming (gRPC).
package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	capacontroladores "servidor.local/grpc-servidor/capaControladores"
	pb "servidor.local/grpc-servidor/serviciosAudio"
)

const direccionServidor = ":50051"

func main() {
	lis, err := net.Listen("tcp", direccionServidor)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAudioServiceServer(grpcServer, capacontroladores.NuevoControladorServidor())

	fmt.Println("==================================================")
	fmt.Println("   Servidor de Streaming (gRPC)")
	fmt.Printf("   Escuchando en %s ...\n", direccionServidor)
	fmt.Println("==================================================")
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
