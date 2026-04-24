package aliases

// @bingen:define[string]:github.com/opencost/bingen/tests/shared.Name
// @bingen:define[*int]:github.com/opencost/bingen/tests/shared.Age
// @bingen:define[[]float64]:github.com/opencost/bingen/tests/shared.FloatList
// @bingen:define[map[string]int]:github.com/opencost/bingen/tests/shared.StrMap
// @bingen:define[[]*uint32]:github.com/opencost/bingen/tests/shared.UIntPtrList
// @bingen:define[[][]map[string]*int]:github.com/opencost/bingen/tests/shared.DoubleSlice

// @bingen:generate[streamable,stringtable]:Parent
// @bingen:generate:Child
// @bingen:generate:Info
// @bingen:generate:ChildInfo
// @bingen:generate:OtherChildInfo
