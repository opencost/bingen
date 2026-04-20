package generator

// DebugEncodeType is the encode type used in debug scope
type DebugEncodeType string

const (
	DebugEncodeTypeWrite DebugEncodeType = "write"
	DebugEncodeTypeRead  DebugEncodeType = "read"
)

// DebugScope is an implementation prototype that represents a debuggable scope that
// contains specific elements of context based state. Calling End will pop the scope
// and flush the state.
type DebugScope interface {
	// End pops the scope and flushes state
	End()
}

// debugScope is the default GeneratorContext based implementation
type debugScope struct {
	gc         GeneratorContext
	encodeType DebugEncodeType
	genType    string
	typeName   string
}

// BeginDebugScope generates a new debug scope for the provided context and state.
func BeginDebugScope(gc GeneratorContext, encodeType DebugEncodeType, genType, typeName string) DebugScope {
	gc.Writeln("// --- [begin][%s][%s](%s) ---", encodeType, genType, typeName)
	return &debugScope{
		gc:         gc,
		encodeType: encodeType,
		genType:    genType,
		typeName:   typeName,
	}
}

// End pops the scope and flushes state
func (ds *debugScope) End() {
	ds.gc.Writeln("// --- [end][%s][%s](%s) ---\n", ds.encodeType, ds.genType, ds.typeName)
}
