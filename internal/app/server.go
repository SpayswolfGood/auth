package app

import (
	auth "github.com/DrusGalkin/forum-auth-grpc/pkg/api/g_rpc"
	"github.com/DrusGalkin/forum-auth-grpc/pkg/logger"
	"github.com/DrusGalkin/forum-auth-grpc/pkg/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

func StartGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Logger.Fatal("Ошибка создания подключения для gRPC", zap.Error(err))
	}

	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, &server.AuthServer{})

	logger.Logger.Debug("gRPC сервер стартует")
	if err := s.Serve(lis); err != nil {
		logger.Logger.Error("Ошибка запуска gRPC сервера", zap.Error(err))
	}
}
