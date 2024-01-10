package model

type Statistic struct {
	OrganizationName string
	Employees        []*EmployeeDetail
}

type EmployeeDetail struct {
	FirstName    string
	LastName     string
	Username     string
	OrdersAmount float32
}
