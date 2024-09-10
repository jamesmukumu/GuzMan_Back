package mpesaexpress

import "gorm.io/gorm"

type Customer_Details struct {
gorm.Model
Customer_Number string `json:"customer_no" gorm:"not null"`
Amount string `json:"amount" gorm:"not null"`
Date_Pay string 
Checkout_Request_Id string `gorm:"unique;not null;primaryKey"`
Confirmation_Payment_Mpesa Confirmation_Payment_Mpesa `gorm:"foreignKey:CheckoutRequestID; references:Checkout_Request_Id"`
}

type Mpesa_Payment struct{
MerchantRequestID string `gorm:"not null"`
CheckoutRequestID string `gorm:"unique; not null"`
ResponseCode string `gorm:"not null"`
ResponseDescription string `gorm:"not null"`
CustomerMessage string `gorm:"not null"`
}  

type Confirmation_Payment_Mpesa struct{
gorm.Model
ResponseCode string `gorm:"not null"`
ResponseDescription string `gorm:"not null"`
MerchantRequestID string `gorm:"not null;unique"`
CheckoutRequestID string `gorm:"not null;unique"`
ResultCode string `gorm:"not null"`
ResultDesc string `gorm:"not null"`
}