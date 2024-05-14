package bannerhandlers

import (
	"banner-serivce/internal/api/response"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

type RequestGetBanners struct {
	FeatureID *int `json:"feature_id"`
	TagID     *int `json:"tag_id"`
	Limit     *int `json:"limit"`
	Offset    *int `json:"offset"`
}

func NewGetBannersHandler(bannerRepo Banners, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := parseGetBannersRequest(r)

		banners, err := bannerRepo.FindBannersByParameters(r.Context(), req)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			logger.Error("Failed to get banners")
			render.JSON(w, r, response.Error("Failed to get banners"))
			return
		}

		render.JSON(w, r, banners)
	}
}

func parseGetBannersRequest(r *http.Request) RequestGetBanners {
	req := RequestGetBanners{}

	if featureIDStr := r.URL.Query().Get("feature_id"); featureIDStr != "" {
		featureID, _ := strconv.Atoi(featureIDStr)
		req.FeatureID = &featureID
	}

	if tagIDStr := r.URL.Query().Get("tag_id"); tagIDStr != "" {
		tagID, _ := strconv.Atoi(tagIDStr)
		req.TagID = &tagID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		req.Limit = &limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, _ := strconv.Atoi(offsetStr)
		req.Offset = &offset
	}

	return req
}
