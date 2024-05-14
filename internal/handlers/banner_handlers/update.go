package bannerhandlers

import (
	errMsg "banner-serivce/internal/api/err"
	"banner-serivce/internal/api/response"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type RequestUpdateBanner struct {
	TagIDs    []int                  `json:"tag_ids" validate:"required"`
	FeatureID int                    `json:"feature_id" validate:"required"`
	Content   map[string]interface{} `json:"content" validate:"required"`
	IsActive  bool                   `json:"is_active" validate:"required"`
}

func NewUpdateBannerHandler(bannerRepo Banners, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bannerID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			logger.Error("Invalid banner ID")
			render.JSON(w, r, response.Error("Invalid banner ID"))
			return
		}

		var req RequestUpdateBanner

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			validateErr := err.(validator.ValidationErrors)
			logger.Error("Invalid request", errMsg.Err(err))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		banner, err := bannerRepo.FindBannerByID(r.Context(), bannerID)
		if err != nil {
			render.Status(r, http.StatusNotFound)
			logger.Error("Failed to find banner")
			render.JSON(w, r, response.Error("Banner not found"))
			return
		}

		banner.TagIDs = req.TagIDs
		banner.FeatureID = req.FeatureID
		banner.Content = req.Content
		banner.IsActive = req.IsActive
		banner.UpdatedAt = time.Now()

		err = bannerRepo.UpdateBanner(r.Context(), &banner)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			logger.Error("Failed to update banner")
			render.JSON(w, r, response.Error("Failed to update banner"))
			return
		}

		responseOK(w, r, banner)
	}

}
