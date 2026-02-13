package models

type DashboardSummary struct {
	CarsCount          int
	CountriesCount     int
	UsersCount         int
	CalculationsCount  int
	PurchasesCount     int
	AveragePurchaseSum float64

	RecentPurchases []Purchase
}
