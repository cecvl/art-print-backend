package models

type FulfillmentMode string

const (
	FulfillmentAuto   FulfillmentMode = "auto"
	FulfillmentManual FulfillmentMode = "manual"
	FulfillmentSmart  FulfillmentMode = "smart"
)
