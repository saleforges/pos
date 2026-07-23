package category

type createCategoryReq struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parentId,omitempty"`
}

type updateCategoryReq struct {
	Name     *string `json:"name,omitempty"`
	ParentID *int64  `json:"parentId,omitempty"`
}
