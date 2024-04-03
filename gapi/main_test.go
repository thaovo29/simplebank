package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/thaovo29/simplebank/db/sqlc"
	"github.com/thaovo29/simplebank/token"
	"github.com/thaovo29/simplebank/util"
	"github.com/thaovo29/simplebank/worker"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistribution worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistribution)
	require.NoError(t, err)
	return server
}

func newContextWithBearerToken(t *testing.T, token token.Maker, username string, role string, duration time.Duration) context.Context {
	ctx := context.Background()
	accessToken, _, err := token.CreateToken(username, role, duration)
	require.NoError(t, err)
	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{
			bearerToken,
		},
	}
	return metadata.NewIncomingContext(ctx, md)
}
