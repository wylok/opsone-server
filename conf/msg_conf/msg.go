package msg_conf

type RMsg struct {
	MsgType string `json:"msg_type" binding:"required"` //消息类型
	Level   string `json:"level" binding:"required"`    //消息级别
	Content string `json:"content" binding:"required"`  //消息内容
	Title   string `json:"title"`                       //消息标题
}
type DMsg struct {
	MsgIds []string `json:"msg_ids" binding:"required"` //消息ID列表
}
type QMsg struct {
	MsgId   string `form:"msg_id"`
	Title   string `form:"title"`
	MsgType string `form:"msg_type"`
	Level   string `form:"level"`
	Status  string `form:"status"`
	Page    int    `form:"page"`
	PerPage int    `form:"pre_page"`
}
type MsgDetail struct {
	MsgId string `form:"msg_id" binding:"required"`
}
