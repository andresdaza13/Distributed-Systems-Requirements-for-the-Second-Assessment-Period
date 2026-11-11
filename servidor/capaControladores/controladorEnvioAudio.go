package capacontroladores

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	capafachadaservices "servidor.local/grpc-servidor/capaFachadaServices"
	"servidor.local/grpc-servidor/capaPublicador"
	"servidor.local/grpc-servidor/capalogger"
	pb "servidor.local/grpc-servidor/serviciosCancion"
)

type ControladorServidor struct {
	pb.UnimplementedAudioServiceServer
	logger     *capalogger.Logger
	publicador *capaPublicador.Publicador
}

func NewControladorServidor(logger *capalogger.Logger) *ControladorServidor {
	return &ControladorServidor{
		logger:     logger,
		publicador: capaPublicador.CrearUnicaInstancia(),
	}
}

// Implementación del procedimiento remoto
func (thisC *ControladorServidor) EnviarCancionMedianteStream(
	req *pb.PeticionDTO, stream pb.AudioService_EnviarCancionMedianteStreamServer) error {

	direcionCliente := ObtenerDireccionCliente(stream.Context())
	usuario := ObtenerUsuario(stream.Context())

	// operación asíncrona: el streaming no espera al logger
	go thisC.logger.AlmacenarSolicitud(req.Titulo, direcionCliente)

	combined := req.GetTitulo() + "." + req.GetFormato()
	err := capafachadaservices.StreamAudioFile(
		combined,
		func(data []byte) error {
			return stream.Send(&pb.FragmentoCancion{Data: data})
		})

	// Solo si el envío terminó SIN error se cuenta la reproducción (opción d).
	// Asíncrono: publicar en la cola no debe frenar el retorno del stream.
	if err == nil {
		go thisC.publicador.Publicar(capaPublicador.EventoReproduccion{
			Usuario: usuario,
			AudioID: req.GetTitulo(),
		})
	}
	return err
}

func ObtenerDireccionCliente(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "desconocido"
}

// ObtenerUsuario lee el metadato x-usuario que el cliente adjunta a la llamada
// gRPC-Web. Así identificamos quién reproduce SIN cambiar el contrato .proto.
func ObtenerUsuario(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		valores := md.Get("x-usuario")
		if len(valores) > 0 && valores[0] != "" {
			return valores[0]
		}
	}
	return "anonimo"
}
