package database

import "fmt"

// DBConf 数据库连接配置
type DBConf struct {
	Host     string `json:",default=127.0.0.1"`
	Port     int    `json:",default=3306"`
	User     string `json:",default=root"`
	Password string `json:",default="`
	Database string `json:",default=business"`
	MaxConns int    `json:",default=20"`
	MaxIdle  int    `json:",default=10"`
}

// DSN 返回 MySQL 连接字符串
func (c DBConf) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database)
}
