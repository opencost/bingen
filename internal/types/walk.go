package types

// WalkType invokes visit for t and for every element type nested within it,
// descending into slice elements, map keys and values, alias targets, and
// resolved references. References that do not resolve are visited as-is (the
// leaf reference), letting callers decide how to treat external types.
//
// It centralizes the structural recursion shared by the code generators so that
// adding a new nested kind is a single change.
func WalkType(t GenType, visit func(GenType)) {
	visit(t)

	switch t.Code() {
	case TypeSlice:
		WalkType(t.(*SliceType).InnerType, visit)
	case TypeMap:
		m := t.(*MapType)
		WalkType(m.KeyType, visit)
		WalkType(m.ValueType, visit)
	case TypeAlias:
		WalkType(t.(*AliasType).Alias, visit)
	case TypeReference:
		if resolved := t.(*ReferenceType).Resolve(); resolved != nil {
			WalkType(resolved, visit)
		}
	}
}
