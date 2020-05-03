package Basics

import (
	"time"
)

const (
	JYBFL = "https://cs.jiaoyubao.cn/edu/"
	JYB   = "https://cs.jiaoyubao.cn"
	Allo  = "https://github.com/"
)

func InitConf() {
	ConfSql = MySql{
		"root",
		"332214",
		"127.0.0.1",
		"3306",
		"test",
	}
}

var ConfSql MySql

type Model struct {
	ID        uint `gorm:"primary_key" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type MySql struct {
	Name     string
	Password string
	Host     string
	Port     string
	Database string
}

type TrainingSchool struct {
	ID          uint     `json:"id"`
	Url         string   `json:"url"`
	Name        string   `json:"name"`
	TypeID      uint     `json:"type_id"`
	TypeUrl     string   `json:"type_url"`
	TypeName    string   `json:"type_name"`
	BrightSpot  []string `json:"bright_spot"`  //亮点，特色
	Info        string   `json:"info"`         //简介
	Course      []string `json:"course"`       //课程
	Campus      string   `json:"campus"`       //校区
	PhoneNumber string   `json:"phone_number"` //联系电话
}

type Type struct {
	Model
	TypeName string `json:"type_name"`
	TypeUrl  string `json:"type_url"`
	MaxPage  int    `json:"max_page"`
	Count    int    `json:"count"`
}

var EveryType []Type

type TsUrl struct {
	Model
	TypeID uint   `json:"type_id"`
	Url    string `json:"url"`
}
