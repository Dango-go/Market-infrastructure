package http

type createReviewRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Rating    int32  `json:"rating"`
	Title     string `json:"title" validate:"omitempty,max=255"`
	Body      string `json:"body" validate:"required,max=4000"`
}

type updateReviewRequest struct {
	Rating int32  `json:"rating"`
	Title  string `json:"title" validate:"omitempty,max=255"`
	Body   string `json:"body" validate:"required,max=4000"`
}
