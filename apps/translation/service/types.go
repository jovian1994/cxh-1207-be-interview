package service

type TaskData struct {
	Id         int    `json:"id"`
	Status     int    `json:"status"`
	CreateBy   string `json:"create_by"`
	ResultKey  string `json:"result_key"`
	IsOss      int    `json:"is_oss"`
	Content    string `json:"content"`
	Lang       string `json:"lang"`
	TargetLang string `json:"target_lang"`
	Result     string `json:"result"`
}
