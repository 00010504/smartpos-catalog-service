package models

type MeasurementUnit struct {
	Id        string `json:"id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
	Precision string `json:"precision"`
}
