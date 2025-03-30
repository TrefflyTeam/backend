package api

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	mockdb "treffly/db/mock"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

func addCookie(
	request *http.Request,
	name string,
	token string,
	expires time.Time,
	path string,
	domain string,
) {
	request.AddCookie(&http.Cookie{
		Name:    name,
		Value:   token,
		Path:    path,
		Domain:  domain,
		Expires: expires,
		Secure: false,
		HttpOnly: true,
	})
}

func randomSession(t *testing.T, tokenMaker token.Maker) db.Session {
	userID := int32(util.RandomInt(1, 100))

	refreshToken, payload, err := tokenMaker.CreateToken(userID, time.Minute)
	require.NoError(t, err)

	return db.Session{
		Uuid:         payload.ID,
		UserID:       payload.UserID,
		RefreshToken: refreshToken,
		ExpiresAt:    payload.ExpiredAt,
		IsBlocked:    false,
	}
}

func TestRefreshTokensAPI(t *testing.T) {
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	session := randomSession(t, tokenMaker)
	testCases := []struct {
		name          string
		setupRequest  func(t *testing.T, request *http.Request)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(session, nil)

				store.EXPECT().
					UpdateSession(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(ctx context.Context, arg db.UpdateSessionParams) (db.Session, error) {
						require.Equal(t, session.Uuid, arg.OldUuid)
						require.NotEmpty(t, arg.NewUuid)
						require.NotEmpty(t, arg.RefreshToken)
						require.True(t, arg.ExpiresAt.After(time.Now()), "ExpiresAt is not in the future")

						return db.Session{
							Uuid:         arg.NewUuid,
							UserID:       session.UserID,
							RefreshToken: arg.RefreshToken,
							ExpiresAt:    arg.ExpiresAt,
							IsBlocked:    false,
						}, nil
					})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)

				cookies := recorder.Result().Cookies()
				require.Len(t, cookies, 2)

				var accessCookie, refreshCookie *http.Cookie
				for _, cookie := range cookies {
					switch cookie.Name {
					case "access_token":
						accessCookie = cookie
					case "refresh_token":
						refreshCookie = cookie
					}
				}
				require.NotNil(t, accessCookie)
				require.NotNil(t, refreshCookie)

				accessPayload, err := tokenMaker.VerifyToken(accessCookie.Value)
				require.NoError(t, err)
				require.True(t, accessPayload.ExpiredAt.After(time.Now()))

				refreshPayload, err := tokenMaker.VerifyToken(refreshCookie.Value)
				require.NoError(t, err)
				require.True(t, refreshPayload.ExpiredAt.After(time.Now()))
			},
		},
		{
			name: "InvalidToken",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", "invalid_token",
					time.Now().Add(time.Hour), refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "BlockedSession",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				blockedSession := session
				blockedSession.IsBlocked = true

				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(blockedSession, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredSession",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				expiredSession := session
				expiredSession.ExpiresAt = time.Now().Add(-time.Hour)

				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(expiredSession, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UserIDMismatch",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				invalidSession := session
				invalidSession.UserID = session.UserID + 1

				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(invalidSession, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SessionNotFound",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(db.Session{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "DatabaseErrorOnGetSession",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(db.Session{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DatabaseErrorOnUpdateSession",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(session, nil)

				store.EXPECT().
					UpdateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DatabaseErrorOnUpdateSession",
			setupRequest: func(t *testing.T, request *http.Request) {
				addCookie(request, "refresh_token", session.RefreshToken,
					session.ExpiresAt, refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), session.Uuid).
					Times(1).
					Return(session, nil)

				store.EXPECT().
					UpdateSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TokenMismatch",
			setupRequest: func(t *testing.T, request *http.Request) {
				otherToken, _, err := tokenMaker.CreateToken(session.UserID, time.Minute)
				require.NoError(t, err)
				addCookie(request, "refresh_token", otherToken,
					time.Now().Add(time.Hour), refreshTokenCookiePath, cookieDomain)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NoTokenCookie",
			setupRequest: func(t *testing.T, request *http.Request) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			server.tokenMaker = tokenMaker
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodPost, "/tokens/refresh", nil)
			require.NoError(t, err)

			tc.setupRequest(t, request)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}