package java

// externalType describes how a non-generated go type that bingen encodes as an
// external reference (length-prefixed MarshalBinary output) is handled in Java:
// the Java type a field maps to, the imports that type needs, and the runtime
// class that reproduces the reference's byte format.
type externalType struct {
	javaType     string   // ie:  "OffsetDateTime"
	typeImports  []string // ie:  []string{"java.time.OffsetDateTime"}
	runtimeClass string   // codec class in the runtime package, ie:  "GoTime"
}

// externalTypes maps a go reference type name (as it appears in the IR, ie: 
// "time.Time") to its Java handling. User-defined external types are not yet
// registrable here; that needs a config surface analogous to @bingen:import.
var externalTypes = map[string]externalType{
	"time.Time": {
		javaType:     "OffsetDateTime",
		typeImports:  []string{"java.time.OffsetDateTime"},
		runtimeClass: "GoTime",
	},
}
