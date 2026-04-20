package aliases

type Child string

type Info struct {
	Name string
	Age  uint
}

type ChildInfo *Info

type OtherChildInfo []ChildInfo

type Parent struct {
	Name           string
	Age            int
	FirstChild     Child
	FirstChildInfo ChildInfo
	Children       []Child
	ChildrenInfo   OtherChildInfo
}

func NewGeneratedParent() *Parent {
	children := []string{
		"Bob",
		"Alice",
		"Jill",
		"Suzy",
		"Charles",
		"Bill",
	}

	var age uint = 19
	allChildren := []Child{}
	infos := []ChildInfo{}
	for _, c := range children {
		allChildren = append(allChildren, Child(c))
		inf := ChildInfo(&Info{
			Name: c,
			Age:  age,
		})
		age -= 1
		infos = append(infos, inf)
	}

	return &Parent{
		Name:           "Dad",
		Age:            51,
		FirstChild:     allChildren[0],
		FirstChildInfo: infos[0],
		Children:       allChildren,
		ChildrenInfo:   OtherChildInfo(infos),
	}

}
