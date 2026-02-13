package models

type Car struct {
	ID           int64
	Brand        string
	Model        string
	Year         int
	EngineVolume int
	EngineType   string
	BasePrice    float64
	CountryID    int64
	CountryName  string
	ImageURL     string
}
