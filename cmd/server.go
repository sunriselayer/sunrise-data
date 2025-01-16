package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	proxy "github.com/rollkit/go-da/proxy/grpc"
	"github.com/sunriselayer/sunrise-data/config"
	"github.com/sunriselayer/sunrise-data/rollkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Serve(conf config.Config) {
	ctx := context.Background()
	da := rollkit.NewSunriseDA(ctx, conf)
	srv := proxy.NewServer(da, grpc.Creds(insecure.NewCredentials()))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Rollkit.Port))
	if err != nil {
		log.Fatalln("failed to create network listener:", err)
	}
	log.Println("serving avail-da over gRPC on:", lis.Addr())
	err = srv.Serve(lis)
	if !errors.Is(err, grpc.ErrServerStopped) {
		log.Fatalln("gRPC server stopped with error:", err)
	}
}
