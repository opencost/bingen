package streaming

import "time"

type Category struct {
	Name  string
	Count int
}

type Item struct {
	Name        string
	Description string
	Price       float64
	Created     time.Time
}

type TestModel struct {
	ID         string
	Name       string
	Created    time.Time
	Items      map[string]*Item
	Categories map[string]*Category
}
