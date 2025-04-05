package apperror

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
)

type ErrorResponse struct {
	HTTPCode int    `json:"-"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Cause    error  `json:"-"`
}

var (
	GeneralBadRequest = ErrorTemplate{
		HTTPCode: http.StatusBadRequest,
		Title:    "Ошибка запроса",
		Subtitle: "Перезагрузите страницу",
	}

	NotFound = ErrorTemplate{
		HTTPCode: http.StatusNotFound,
		Title:    "Ничего не найдено",
		Subtitle: "Запрашиваемый ресурс недоступен или не существует",
	}

	InvalidCredentials = ErrorTemplate{
		HTTPCode: http.StatusUnauthorized,
		Title:    "Неверный логин или пароль",
		Subtitle: "Попробуйте ещё раз",
	}

	TokenExpired = ErrorTemplate{
		HTTPCode: http.StatusUnauthorized,
		Title:    "Сессия завершена",
		Subtitle: "Войдите снова, чтобы продолжить",
	}

	EmailTaken = ErrorTemplate{
		HTTPCode: http.StatusBadRequest,
		Title:    "Почта уже занята",
		Subtitle: "Укажите другую почту или войдите в аккаунт",
	}

	BadRequest = ErrorTemplate{
		HTTPCode: http.StatusBadRequest,
		Title:    "Некорректные данные",
		Subtitle: "Проверьте введённую информацию и попробуйте снова",
	}

	Forbidden = ErrorTemplate{
		HTTPCode: http.StatusForbidden,
		Title:    "Недостаточно прав",
		Subtitle: "У вас нет доступа к этому разделу",
	}

	InternalServer = ErrorTemplate{
		HTTPCode: http.StatusInternalServerError,
		Title:    "Ошибка сервера",
		Subtitle: "Что-то пошло не так. Попробуйте позже",
	}

	BadGateway = ErrorTemplate{
		HTTPCode: http.StatusBadGateway,
		Title:    "Сервер не отвечает",
		Subtitle: "Запрос занял слишком много времени. Попробуйте позже",
	}
)

type ErrorTemplate struct {
	HTTPCode int
	Title    string
	Subtitle string
}

func (t ErrorTemplate) WithCause(cause error) ErrorResponse {
	return ErrorResponse{
		HTTPCode: t.HTTPCode,
		Title:    t.Title,
		Subtitle: t.Subtitle,
		Cause:    cause,
	}
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Subtitle)
}

func (e ErrorResponse) Unwrap() error {
	return e.Cause
}

func WrapDBError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound.WithCause(err)
	}
	if errors.Is(err, sql.ErrConnDone) {
		return InternalServer.WithCause(err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "users_email_key" {
				return EmailTaken.WithCause(err)
			}
			if pgErr.ConstraintName == "user_tags_pkey" {
				return BadRequest.WithCause(err)
			}
			if pgErr.ConstraintName == "event_user_pkey" {
				return BadRequest.WithCause(err)
			}
		case pgerrcode.ForeignKeyViolation:
			return BadRequest.WithCause(err)
		}
	}

	return GeneralBadRequest.WithCause(err)
}
