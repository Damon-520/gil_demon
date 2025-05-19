package gorm_builder

type Cond struct {
	Field    string
	Operator string
	Value    interface{}
}

type Query struct {
	Where *[]Cond
	Order string
	Sort  string
}
