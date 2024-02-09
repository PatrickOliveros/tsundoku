package models

import "tsundoku/driver"

type AppConfig struct {
	DataChannel chan BookData
	DB          *driver.DB
}
