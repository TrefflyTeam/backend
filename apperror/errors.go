package apperror

import "net/http"

type ErrorResponse struct {
	HTTPCode int    `json:"-"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

var (
	GeneralBadRequest = ErrorResponse{
		HTTPCode: http.StatusBadRequest,
		Title:    "Ошибка запроса",
		Subtitle: "Перезагрузите страницу",
	}

	NotFound = ErrorResponse{
		HTTPCode: http.StatusNotFound,
		Title:    "Ничего не найдено",
		Subtitle: "Запрашиваемый ресурс недоступен или не существует",
	}

	InvalidCredentials = ErrorResponse{
		HTTPCode: http.StatusUnauthorized,
		Title:    "Неверный логин или пароль",
		Subtitle: "Попробуйте ещё раз",
	}

	TokenExpired = ErrorResponse{
		HTTPCode: http.StatusUnauthorized,
		Title: "Сессия завершена",
		Subtitle: "Войдите снова, чтобы продолжить",
	}

	EmailTaken = ErrorResponse{
		HTTPCode: http.StatusUnauthorized,
		Title: "Почта уже занята",
		Subtitle: "Укажите другую почту или войдите в аккаунт",
	}

	BadRequest = ErrorResponse{
		HTTPCode: http.StatusBadRequest,
		Title: "Некорректные данные",
		Subtitle: "Проверьте введённую информацию и попробуйте снова",
	}

	Forbidden = ErrorResponse{
		HTTPCode: http.StatusForbidden,
		Title: "Недостаточно прав",
		Subtitle: "У вас нет доступа к этому разделу",
	}

	InternalServer = ErrorResponse{
		HTTPCode: http.StatusInternalServerError,
		Title: "Ошибка сервера",
		Subtitle: "Что-то пошло не так. Попробуйте позже",
	}

	BadGateway = ErrorResponse{
		HTTPCode: http.StatusBadGateway,
		Title: "Сервер не отвечает",
		Subtitle: "Запрос занял слишком много времени. Попробуйте позже",
	}
)
