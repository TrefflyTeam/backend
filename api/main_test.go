package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
	db "treffly/db/sqlc"
	"treffly/util"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration: time.Minute,
		RefreshTokenDuration: time.Hour * 24 * 7,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
