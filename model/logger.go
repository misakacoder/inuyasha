package model

type LogLevel struct {
	//调用器
	Caller string `json:"caller" binding:"trim"`
	//表名
	Table string `json:"table" binding:"trim"`
	//日志级别
	Level string `json:"level" binding:"required" enums:"DEBUG,INFO,WARN,ERROR,PANIC"`
	//过期时间 例如：1s 1m 1h
	Time string `json:"time" binding:"required"`
}
