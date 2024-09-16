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
Favourites Favourites_Customers `gorm:"foreignKey:Created_By;references:Users_Name"`
}


type Favourites_Customers struct{
gorm.Model
Phone_Number string `json:"phone_number" gorm:"not null"`
Name string `json:"name" gorm:"not null"`
Created_By string `gorm:"not null"`

}




func (user Users)PresetTodefault(){
var timeNow string  = time.Now().Format(time.ANSIC)
user.Created_At = timeNow
}
