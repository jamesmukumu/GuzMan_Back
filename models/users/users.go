package users

import (
	"time"

	"gorm.io/gorm"
)

type Users struct{
gorm.Model
Users_Name string `json:"user_name" gorm:"not null;unique"`
Pin string `json:"Pin" gorm:"not null"`
Created_At string `gorm:"not null"`
Can_Access_All_Functions bool `json:"access_all"`
}



func (user Users)PresetTodefault(){
var timeNow string  = time.Now().Format(time.ANSIC)
user.Created_At = timeNow
}
