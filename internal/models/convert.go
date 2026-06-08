package models

type RequestConvert struct {
	TypeName string `form:"type" json:"type"`
	Filename string `form:"filename" json:"filename"`
}

type ResponseConvert struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
}
