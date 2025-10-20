package save

import (
	"errors"
	"log/slog"
	"net/http"
	"rest-api/internal/lib/api/responce"
	resp "rest-api/internal/lib/api/responce"
	"rest-api/internal/lib/logger/sl"
	"rest-api/internal/lib/random"
	"rest-api/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" `
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

const aliasLength = 4

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		var req Request

		err := render.DecodeJSON(request.Body, &req)
		if err != nil {
			log.Error("не смог декодировать тело запроса", sl.Err(err))

			render.JSON(writer, request, responce.Error("не смог распарсить тело запроса"))

			return
		}

		log.Info("Запрос тело декодер", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(writer, request, responce.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("урла такой есть уже", slog.String("url", req.URL))
			render.JSON(writer, request, responce.Error("урла такой есть уже существует"))
			return
		}

		if err != nil {
			log.Error("НЕ получилось сохранить урл", sl.Err(err))

			render.JSON(writer, request, responce.Error("НЕ получилось сохранить урл"))
			return
		}

		log.Info("URL добавлен", slog.Int64("id", id))

		responseOK(writer, request, alias)

	}
}

func responseOK(writer http.ResponseWriter, request *http.Request, alias string) {
	render.JSON(writer, request, responce.Response{
		Status: responce.StatusOk,
		Alias:  alias,
	})
}
