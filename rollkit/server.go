package rollkit

import (
	"context"
	"errors"
	"fmt"
	"net"

	proxy "github.com/rollkit/go-da/proxy/grpc"
	"github.com/rs/zerolog/log"
	scontext "github.com/sunriselayer/sunrise-data/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Serve() {
	ctx := context.Background()
	da := NewSunriseDA(ctx, scontext.Config)
	srv := proxy.NewServer(da, grpc.Creds(insecure.NewCredentials()))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", scontext.Config.Rollkit.Port))
	if err != nil {
		log.Err(err).Msgf("failed to create network listener on port %d", scontext.Config.Rollkit.Port)
	}
	log.Info().Msgf("Running rollkit go-da API on localhost: %d", scontext.Config.Rollkit.Port)

	err = srv.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		log.Err(err).Msg("gRPC server stopped with error")
	}
}
