package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thaovo29/simplebank/util"
)

func TestPasetoMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute
	role := util.DepositorRole
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(username, role,  duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, payload.Username, username)
	require.Equal(t, role, payload.Role)

	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredPasetoTokenMaker(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	role := util.DepositorRole

	token, payload, err := maker.CreateToken(username, role, -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token) 
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.EqualError(t, err, ErrExpiredToken.Error())

	require.Nil(t, payload)	
}
