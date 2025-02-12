package models

type DataModel struct {
	ID    string `json:"id" bson:"_id"`
	Value string `json:"value" bson:"value"`
}
