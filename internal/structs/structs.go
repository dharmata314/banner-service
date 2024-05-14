package structs

import "time"

type Banner struct {
	ID        int                    `json:"banner_id"`
	TagIDs    []int                  `json:"tag_ids"`
	FeatureID int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type BannerTag struct {
	BannerID int `json:"banner_id"`
	TagID    int `json:"tag_id"`
}

type Feature struct {
	ID   int    `json:"feature_id"`
	Name string `json:"name"`
}

type Tag struct {
	ID   int    `json:"tag_id"`
	Name string `json:"name"`
}

type User struct {
	ID       int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
