package aliasnil

// Tags is an alias whose underlying type (a slice) is nilable. It exercises the
// alias-of-nilable path: the nil-flag byte must be written and read symmetrically.
type Tags []string

// Holder embeds a Tags field between two scalars so any nil-flag misalignment
// corrupts the trailing field.
type Holder struct {
	Name  string
	Tags  Tags
	Count int
}
