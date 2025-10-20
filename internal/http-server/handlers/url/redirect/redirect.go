package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	resp "rest-api/internal/lib/api/responce"
	_ "rest-api/internal/lib/logger/sl"

	"rest-api/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" `
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		const op = "handler.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(request.Context())),
		)

		alias := chi.URLParam(request, "alias")
		if alias == "" {
			log.Info("alias empty")

			render.JSON(writer, request, resp.Error("empty alias"))

			return
		}

		resURl, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("урл not found", "alias", alias)
			render.JSON(writer, request, resp.Error("not found url"))
			return
		}

		if err != nil {
			log.Info("failed to get url", "alias", alias, "err", err)
			render.JSON(writer, request, resp.Error("internal url"))
			return
		}

		log.Info("get URL", slog.String("url", resURl))

		http.Redirect(writer, request, resURl, http.StatusFound)

	}
}
