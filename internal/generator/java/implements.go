package java

import (
	"fmt"
	"go/types"
	"sort"

	bingentypes "github.com/opencost/bingen/internal/types"
	"golang.org/x/tools/go/packages"
)

// interfaceInfo captures the interface implementation graph derived from the go
// type checker for a single generated package.
type interfaceInfo struct {
	// PkgPath is the fully-qualified import path of the source package, used to
	// build the wire type strings for interface values.
	PkgPath string

	// Implements maps each generated struct name to the sorted set of generated
	// interface names it satisfies.
	Implements map[string][]string
}

// wireName returns the go wire type string written for a value of the named
// struct when it appears as an interface value. Generated bingen types use
// pointer receivers, so the pointer form is emitted (matching go typeToString).
func (ii *interfaceInfo) wireName(structName string) string {
	return "*" + ii.PkgPath + "." + structName
}

// neededInterfaces returns the sorted set of named interface types referenced by
// any generated struct field (recursing through slices, maps, and aliases). The
// empty interface{} is excluded; it maps to Object and needs no marker type.
func neededInterfaces(tc bingentypes.TypeCollection) []string {
	found := make(map[string]bool)

	for _, t := range tc.Types() {
		st, ok := t.(*bingentypes.StructType)
		if !ok {
			continue
		}
		for _, f := range st.Fields {
			bingentypes.WalkType(f.Type, func(gt bingentypes.GenType) {
				if gt.Code() == bingentypes.TypeInterface {
					if name := gt.Name(); name != "interface{}" {
						found[name] = true
					}
				}
			})
		}
	}

	out := make([]string, 0, len(found))
	for name := range found {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// structNames returns the sorted names of all generated struct types, giving a
// deterministic registry order independent of source declaration order.
func structNames(tc bingentypes.TypeCollection) []string {
	var names []string
	for _, t := range tc.Types() {
		if st, ok := t.(*bingentypes.StructType); ok {
			names = append(names, st.Name())
		}
	}
	sort.Strings(names)
	return names
}

// loadInterfaceInfo type-checks the source package and computes the interface
// implementation graph using the go type checker, which resolves embedded and
// cross-package interfaces exactly.
func loadInterfaceInfo(sourceDir string, structs []string, interfaces []string) (*interfaceInfo, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  sourceDir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("loading go types for %q: %w", sourceDir, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no go package found in %q", sourceDir)
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("type errors in %q: %v", sourceDir, pkg.Errors[0])
	}

	scope := pkg.Types.Scope()

	ifaceTypes := make(map[string]*types.Interface, len(interfaces))
	for _, name := range interfaces {
		obj := scope.Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("interface %q not found in package %q", name, pkg.PkgPath)
		}
		iface, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			return nil, fmt.Errorf("type %q is not an interface", name)
		}
		ifaceTypes[name] = iface
	}

	implements := make(map[string][]string)
	for _, structName := range structs {
		obj := scope.Lookup(structName)
		if obj == nil {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		ptr := types.NewPointer(named)

		var satisfied []string
		for _, ifaceName := range interfaces {
			if types.Implements(ptr, ifaceTypes[ifaceName]) {
				satisfied = append(satisfied, ifaceName)
			}
		}
		if len(satisfied) > 0 {
			sort.Strings(satisfied)
			implements[structName] = satisfied
		}
	}

	return &interfaceInfo{
		PkgPath:    pkg.PkgPath,
		Implements: implements,
	}, nil
}
