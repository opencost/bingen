package failing

type Widget struct {
	widgetType int
	Names      map[string]string

	PropertyNoComma int //@bingen:field[version=2 default=15]
	PropertyCorrect int //@bingen:field[version=2, default=15]
	PropertyNoSet   int //@bingen:field[version=2, default]
}
