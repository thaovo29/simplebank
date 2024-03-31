package gapi

import (
	"fmt"

	db "github.com/thaovo29/simplebank/db/sqlc"
	pb "github.com/thaovo29/simplebank/pb"
	"github.com/thaovo29/simplebank/token"
	"github.com/thaovo29/simplebank/util"
	"github.com/thaovo29/simplebank/worker"
)

// Server serves gRPC requests for our banking service
type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	taskDistributor worker.TaskDistributor
}

// new server creates new gRPC server
func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		taskDistributor: taskDistributor,
	}
	return server, nil
}
