package bannerhandlers

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/api/response"
	"banner-serivce/internal/structs"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Banners interface {
	CreateBanner(ctx context.Context, banner *structs.Banner) error
	FindBannerByFeatureTag(ctx context.Context, featureID, tagID int) (*structs.Banner, error)
	DeleteBannerByID(ctx context.Context, id int) error
	FindBannersByParameters(ctx context.Context, params RequestGetBanners) ([]structs.Banner, error)
	UpdateBanner(ctx context.Context, banner *structs.Banner) error
	FindBannerByID(ctx context.Context, id int) (structs.Banner, error)
}

type BannerTags interface {
	CreateBannerTag(ctx context.Context, bannerTag *structs.BannerTag) error
}

type RequestBanner struct {
	TagIDs    []int                  `json:"tag_ids" validate:"required"`
	FeatureID int                    `json:"feature_id" validate:"required"`
	Content   map[string]interface{} `json:"content" validate:"required"`
	IsActive  bool                   `json:"is_active" validate:"required"`
}

type ResponseBanner struct {
	response.Response
	ID        int                    `json:"banner_id"`
	TagIDs    []int                  `json:"tag_ids"`
	FeatureID int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func New(log *slog.Logger, bannerRepository Banners, bannerTagsRepository BannerTags) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const loggerOptions = "handlers.banners.CreateBanner.New"
		log = log.With(
			slog.String("options", loggerOptions),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req RequestBanner
		err := json.NewDecoder(r.Body).Decode(&req)
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

		banner := structs.Banner{
			TagIDs:    req.TagIDs,
			FeatureID: req.FeatureID,
			Content:   req.Content,
			IsActive:  req.IsActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = bannerRepository.CreateBanner(r.Context(), &banner)
		if err != nil {
			log.Error("Failed to create banner", errMsg.Err(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("Failed to create banner"))
			return
		}

		log.Info("banner added")
		for _, tagID := range req.TagIDs {
			bannerTag := structs.BannerTag{
				BannerID: banner.ID,
				TagID:    tagID,
			}
			err = bannerTagsRepository.CreateBannerTag(r.Context(), &bannerTag)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				log.Error("Failed to create banner tag", errMsg.Err(err))
				render.JSON(w, r, response.Error("Failed to create banner tag"))
				return
			}
		}
		log.Info("banner-tegs added")
		responseOK(w, r, banner)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, banner structs.Banner) {
	render.JSON(w, r, ResponseBanner{
		Response:  response.OK(),
		ID:        banner.ID,
		TagIDs:    banner.TagIDs,
		FeatureID: banner.FeatureID,
		Content:   banner.Content,
		IsActive:  banner.IsActive,
	})
}
