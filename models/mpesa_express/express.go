package mpesaexpress

import "gorm.io/gorm"

type Customer_Details struct {
gorm.Model
Customer_Number string `json:"customer_no" gorm:"not null"`
Amount string `json:"amount" gorm:"not null"`
Date_Pay string 
Checkout_Request_Id string `gorm:"unique;not null;primaryKey"`
Confirmation_Payment_Mpesa Confirmation_Payment_Mpesa `gorm:"foreignKey:TransactionID;references:Checkout_Request_Id"`
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
ResultCode string `gorm:"not null"`
ResultDesc string `gorm:"not null"`
TransactionStatus string `gorm:"not null"`
TransactionCode string `gorm:"not null;"`
TransactionID string `gorm:"not null;unique"`
TransactionAmount string `gorm:"not null"`
TransactionReceipt string `gorm:"not_null"`
TransactionDate string `gorm:"not null"`
TransactionReference string `gorm:"not null"`
Msisdn string `gorm:"not null"`
}
        