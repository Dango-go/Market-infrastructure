package http

type categoryCreateRequest struct {
	Name        string `json:"name" validate:"required,max=120"`
	Slug        string `json:"slug" validate:"required,max=120"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

type brandCreateRequest struct {
	Name        string `json:"name" validate:"required,max=120"`
	Slug        string `json:"slug" validate:"required,max=120"`
	Description string `json:"description" validate:"omitempty,max=500"`
	CountryCode string `json:"country_code" validate:"omitempty,len=2"`
}

type specRequest struct { Key string `json:"key" validate:"required,max=80"`; Value string `json:"value" validate:"required,max=300"` }
type mediaRequest struct { URL string `json:"url" validate:"required,url,max=512"`; Type string `json:"type" validate:"required,max=40"`; SortOrder int32 `json:"sort_order"` }
type compatibilityRequest struct { Kind string `json:"kind" validate:"required,max=80"`; Value string `json:"value" validate:"required,max=160"` }

type productCreateRequest struct {
	CategoryID       string                 `json:"category_id" validate:"required,uuid"`
	BrandID          string                 `json:"brand_id" validate:"required,uuid"`
	Slug             string                 `json:"slug" validate:"required,max=160"`
	SKU              string                 `json:"sku" validate:"required,max=80"`
	Name             string                 `json:"name" validate:"required,max=200"`
	ShortDescription string                 `json:"short_description" validate:"omitempty,max=300"`
	Description      string                 `json:"description" validate:"omitempty,max=4000"`
	DatasheetURL     string                 `json:"datasheet_url" validate:"omitempty,url,max=512"`
	ImageURL         string                 `json:"image_url" validate:"omitempty,url,max=512"`
	Status           string                 `json:"status" validate:"omitempty,oneof=draft active archived"`
	Featured         bool                   `json:"featured"`
	Specs            []specRequest          `json:"specs"`
	Media            []mediaRequest         `json:"media"`
	Compatibility    []compatibilityRequest `json:"compatibility"`
}

type productUpdateRequest struct {
	CategoryID          *string                 `json:"category_id" validate:"omitempty,uuid"`
	BrandID             *string                 `json:"brand_id" validate:"omitempty,uuid"`
	Slug                *string                 `json:"slug" validate:"omitempty,max=160"`
	SKU                 *string                 `json:"sku" validate:"omitempty,max=80"`
	Name                *string                 `json:"name" validate:"omitempty,max=200"`
	ShortDescription    *string                 `json:"short_description" validate:"omitempty,max=300"`
	Description         *string                 `json:"description" validate:"omitempty,max=4000"`
	DatasheetURL        *string                 `json:"datasheet_url" validate:"omitempty,url,max=512"`
	ImageURL            *string                 `json:"image_url" validate:"omitempty,url,max=512"`
	Status              *string                 `json:"status" validate:"omitempty,oneof=draft active archived"`
	Featured            *bool                   `json:"featured"`
	Specs               []specRequest           `json:"specs"`
	Media               []mediaRequest          `json:"media"`
	Compatibility       []compatibilityRequest  `json:"compatibility"`
	ReplaceSpecs        bool                    `json:"replace_specs"`
	ReplaceMedia        bool                    `json:"replace_media"`
	ReplaceCompatibility bool                   `json:"replace_compatibility"`
}
