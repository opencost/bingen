package core

import "fmt"

// DebugScope is used to output lead and trailing emissions during code generation
type DebugScope interface {
	Begin() DebugScope
	End()
}

// just a default debug emission scope that writes a comment delimitting the start/end
// of an encode/decode scope
type emitScope struct {
	b        Backend
	e        *Emit
	label    string
	header   string
	typeName string
}

// creates a new debug scope emission
func newDebugScope(b Backend, e *Emit, label string, header string, typeName string) DebugScope {
	return &emitScope{
		b:        b,
		e:        e,
		label:    label,
		header:   header,
		typeName: typeName,
	}
}

// BeginDebugScope creates a new DebugScope and calls Begin()
func BeginDebugScope(b Backend, e *Emit, label, header, typeName string) DebugScope {
	return newDebugScope(b, e, label, header, typeName).Begin()
}

func (es *emitScope) Begin() DebugScope {
	t := fmt.Sprintf("--- [begin][%s][%s](%s) ---", es.label, es.header, es.typeName)
	es.b.Comment(es.e, t)
	return es
}

func (es *emitScope) End() {
	t := fmt.Sprintf("--- [end][%s][%s](%s) ---", es.label, es.header, es.typeName)
	es.b.Comment(es.e, t)
}
