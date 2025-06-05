package model

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg,omitempty"`
	Data interface{} `json:"data"`
}

func NewErrResp(err error) *Resp {
	return &Resp{
		Code: 1,
		Msg:  err.Error(),
	}
}
func NewResp(data interface{}) *Resp {
	return &Resp{
		Data: data,
	}
}

func NewSuccessResp() *Resp {
	return &Resp{}
}
