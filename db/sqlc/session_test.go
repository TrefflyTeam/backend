package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"treffly/token"
	"treffly/util"
)

func createRandomSession(t *testing.T, userID int32, tokenMaker token.Maker) Session {
	refreshToken, payload, err := tokenMaker.CreateToken(userID, time.Hour)
	require.NoError(t, err)

	arg := CreateSessionParams{
		UserID:       payload.UserID,
		Uuid:         payload.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    payload.ExpiredAt,
		IsBlocked:    false,
	}

	session, err := testQueries.CreateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, session)
	require.Equal(t, session.UserID, arg.UserID)
	require.Equal(t, session.RefreshToken, arg.RefreshToken)
	require.WithinDuration(t, session.ExpiresAt, arg.ExpiresAt, time.Second)
	require.Equal(t, session.IsBlocked, arg.IsBlocked)

	return session
}

func TestCreateSession(t *testing.T) {
	user := createRandomUser(t)
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	createRandomSession(t, user.ID, tokenMaker)
}

func TestGetSession(t *testing.T) {
	user := createRandomUser(t)
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	session1 := createRandomSession(t, user.ID, tokenMaker)

	session2, err := testQueries.GetSession(context.Background(), session1.Uuid)
	require.NoError(t, err)
	require.NotEmpty(t, session2)

	require.Equal(t, session1.Uuid, session2.Uuid)
	require.Equal(t, session1.UserID, session2.UserID)
	require.Equal(t, session1.RefreshToken, session2.RefreshToken)
	require.WithinDuration(t, session1.ExpiresAt, session2.ExpiresAt, time.Second)
	require.Equal(t, session1.IsBlocked, session2.IsBlocked)
}

func TestUpdateSession(t *testing.T) {
	user := createRandomUser(t)
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	session1 := createRandomSession(t, user.ID, tokenMaker)

	newToken, payload, err := tokenMaker.CreateToken(user.ID, time.Hour)
	require.NoError(t, err)

	arg := UpdateSessionParams{
		payload.ID,
		newToken,
		payload.ExpiredAt,
		session1.Uuid,
	}

	session2, err := testQueries.UpdateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, session2)

	require.Equal(t, session2.Uuid, payload.ID)
	require.Equal(t, session2.UserID, payload.UserID)
	require.Equal(t, session2.RefreshToken, newToken)
	require.WithinDuration(t, session2.ExpiresAt, payload.ExpiredAt, time.Second)
	require.Equal(t, session2.IsBlocked, session2.IsBlocked)
}
