package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	mockdb "treffly/db/mock"
	db "treffly/db/sqlc"
	"treffly/token"
	"treffly/util"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.PasswordHash)
	if err != nil {
		return false
	}

	e.arg.PasswordHash = arg.PasswordHash
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)
	user.ID = int32(util.RandomInt(0,100))
	log.Println(user.Username)
	log.Println(user.Email)
	log.Println(password)
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
						require.NotEmpty(t, arg.Uuid)
						require.NotEmpty(t, arg.UserID)
						require.NotEmpty(t, arg.RefreshToken)
						require.True(t, arg.ExpiresAt.After(time.Now()))
						require.False(t, arg.IsBlocked)

						return db.Session{
							Uuid:         arg.Uuid,
							UserID:       arg.UserID,
							RefreshToken: arg.RefreshToken,
							ExpiresAt:    arg.ExpiresAt,
							IsBlocked:    arg.IsBlocked,
							CreatedAt:    time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "@@#$%",
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"email":    "invalid-email",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": "1",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidUsernameLength",
			body: gin.H{
				"username": "a",
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest("POST", url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:     util.RandomUsername(),
		Email:        util.RandomEmail(),
		PasswordHash: hashedPassword,
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
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.PasswordHash)
}

func TestLoginUser(t *testing.T) {
	user, password := randomUser(t)
	user.ID = int32(util.RandomInt(0,100))
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), user.Email).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
						require.NotEmpty(t, arg.Uuid)
						require.NotEmpty(t, arg.UserID)
						require.NotEmpty(t, arg.RefreshToken)
						require.True(t, arg.ExpiresAt.After(time.Now()))
						require.False(t, arg.IsBlocked)

						return db.Session{
							Uuid:         arg.Uuid,
							UserID:       arg.UserID,
							RefreshToken: arg.RefreshToken,
							ExpiresAt:    arg.ExpiresAt,
							IsBlocked:    arg.IsBlocked,
							CreatedAt:    time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
						}, nil
					})

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var response loginUserResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				require.NoError(t, err)

				require.Equal(t, user.Username, response.Username)
				require.Equal(t, user.Email, response.Email)
				require.WithinDuration(t, user.CreatedAt, response.CreatedAt, time.Second)

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
			name: "UserNotFound",
			body: gin.H{
				"email":    "notfound@example.com",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), "notfound@example.com").
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InvalidPassword",
			body: gin.H{
				"email":    user.Email,
				"password": "wrong_password",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), user.Email).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalErrorOnGetUser",
			body: gin.H{
				"email":    user.Email,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), user.Email).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidEmailFormat",
			body: gin.H{
				"email":    "invalid-email",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/login"
			request, err := http.NewRequest("POST", url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}


func TestLogoutUser(t *testing.T) {
	userID := int32(util.RandomInt(0, 100))
	tokenMaker, err := token.NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	testCases := []struct {
		name          string
		setupRequest  func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(*httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupRequest: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, userID, time.Hour)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)

				cookies := recorder.Result().Cookies()
				require.Len(t, cookies, 2)

				for _, cookie := range cookies {
					require.Equal(t, -1, cookie.MaxAge)
				}
			},
		},
		{
			name: "Unauthorized",
			setupRequest: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

				cookies := recorder.Result().Cookies()
				require.Len(t, cookies, 0)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			router := gin.New()
			router.Use(ErrorHandler())
			server := newTestServer(t, store)

			router.POST(
				"/logout",
				authMiddleware(tokenMaker),
				server.logoutUser,
			)

			recorder := httptest.NewRecorder()
			request, _ := http.NewRequest("POST", "/logout", nil)
			tc.setupRequest(t, request, tokenMaker)

			router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}