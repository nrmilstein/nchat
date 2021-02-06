package chatServer

type wsRequest struct {
	Id     int               `json:"id"`
	Type   string            `json:"type"`
	Method string            `json:"method"`
	Data   map[string]string `json:"data"`
}

type wsSuccessResponse struct {
	Id     int         `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type wsErrorResponse struct {
	Id      int         `json:"id"`
	Type    string      `json:"type"`
	Status  string      `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type wsNotification struct {
	Type   string      `json:"type"`
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}
