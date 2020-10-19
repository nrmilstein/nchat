package chatServer

type wsAuthRequest struct {
	Id     int               `json:"id"`
	Type   string            `json:"type"`
	Method string            `json:"method"`
	Data   wsAuthRequestData `json:"data"`
}

type wsAuthRequestData struct {
	AuthKey string `json:"authKey"`
}

type wsAuthSuccessResponse struct {
	Id     int    `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}