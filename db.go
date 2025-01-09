package main

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type ConfinementCell struct {
	Ip   string    `gorm:"column:ip"`
	Time time.Time `gorm:"column:time"`
}

func OpenDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("db"))
	if err != nil {
		return nil, err
	}
	return db, db.AutoMigrate(&ConfinementCell{})
}
