package taghandlers

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/api/response"
	"banner-serivce/internal/structs"
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type RequestTag struct {
	Name string `json:"name" validate:"required"`
}

type ResponseTag struct {
	response.Response
	ID   int    `json:"tag_id"`
	Name string `json:"name"`
}

type Tag interface {
	CreateTag(ctx context.Context, tag *structs.Tag) error
}
func New(log *slog.Logger, tagRepository Tag) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const loggerOptions = "handlers.features.createTag.New"
		log = log.With(
			slog.String("options", loggerOptions),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req RequestTag
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request body", errMsg.Err(err))
			render.JSON(w, r, response.Error("Failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("Invalid request", errMsg.Err(err))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}
		tag := structs.Tag{Name: req.Name}
		err = tagRepository.CreateTag(r.Context(), &tag)
		if err != nil {
			log.Error("Failed to create tag", errMsg.Err(err))
			render.JSON(w, r, response.Error("Failed to create tag"))
			return
		}
		log.Info("Tag added")
		responseOK(w, r, req.Name, tag.ID)
	}
}
func responseOK(w http.ResponseWriter, r *http.Request, name string, tag_id int) {
	render.JSON(w, r, ResponseTag{Response: response.OK(),
		Name: name, ID: tag_id})
}
