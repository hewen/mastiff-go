package handler

type TestReq struct {
	Name string `json:"name"`
}

type TestResp struct {
	Greet string `json:"greet"`
}
