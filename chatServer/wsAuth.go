package chatServer

type WsAuthRequest struct {
	Id     int               `json:"id"`
	Type   string            `json:"type"`
	Method string            `json:"method"`
	Data   wsAuthRequestData `json:"data"`
}

type wsAuthRequestData struct {
	AuthKey string `json:"authKey"`
}

type WsAuthSuccessResponse struct {
	Id     int    `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}
