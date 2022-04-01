package main

type startResponse struct {
	*dataResponse
	SessionCode string `json:"session_code"`
}

type innerResponse struct {
	*dataResponse
	Content *string `json:"content,omitempty"`
}

type attrResponse struct {
	*dataResponse
	Content string `json:"content"`
}

type jsResponse struct {
	*dataResponse
	Content interface{} `json:"content"`
}
