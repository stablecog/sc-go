// * Responses from admin endpoints
package responses

type AdminGalleryResponseBody struct {
	Updated int `json:"updated"`
}

type AdminDeleteResponseBody struct {
	Deleted int `json:"deleted"`
}
