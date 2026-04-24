package aliases

import "github.com/opencost/bingen/tests/shared"

type Child string

type Info struct {
	Name string
	Age  uint
}

type ChildInfo *Info

type OtherChildInfo []ChildInfo

type Parent struct {
	Name            shared.Name
	Age             shared.Age
	FirstChild      Child
	FirstChildInfo  ChildInfo
	Children        []Child
	ChildrenInfo    OtherChildInfo
	FavoriteNumbers shared.FloatList
	NameMap         shared.StrMap
	U32s            shared.UIntPtrList
	Slices          shared.DoubleSlice
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
		Name:            "Dad",
		Age:             toPtr(51),
		FirstChild:      allChildren[0],
		FirstChildInfo:  infos[0],
		Children:        allChildren,
		ChildrenInfo:    OtherChildInfo(infos),
		FavoriteNumbers: shared.FloatList([]float64{1.2, 3.65, 82.3}),
		NameMap: shared.StrMap(map[string]int{
			children[0]: 1,
			children[1]: 2,
			children[2]: 3,
		}),
	}

}

func toPtr[T any](v T) *T {
	return &v
}
