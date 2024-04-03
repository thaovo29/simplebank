package gapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/thaovo29/simplebank/db/mock"
	db "github.com/thaovo29/simplebank/db/sqlc"
	pb "github.com/thaovo29/simplebank/pb"
	"github.com/thaovo29/simplebank/util"
	"github.com/thaovo29/simplebank/worker"
	mockwk "github.com/thaovo29/simplebank/worker/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	err = actualArg.AfterCreate(expected.user)

	return err == nil
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password, user}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{
						User: user,
					}, nil)
				taskDistributor.EXPECT().
					DistributeTaskSEndVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)

			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createUser := res.GetUser()
				require.Equal(t, user.Username, createUser.Username)
				require.Equal(t, user.FullName, createUser.FullName)
				require.Equal(t, user.Email, createUser.Email)
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
				taskDistributor.EXPECT().
					DistributeTaskSEndVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			storeCtrl := gomock.NewController(t)
			taskCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			defer taskCtrl.Finish()

			store := mockdb.NewMockStore(storeCtrl)
			// build stubs

			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store, taskDistributor)
			//start test server and send request
			server := newTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(8)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	user = db.User{
		Username:       util.RandomOwner(),
		Role:           util.DepositorRole,
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
