package unredo

// Unredoabler is a generic interface to an undo/redo stack.
type Unredoabler interface {
	State() State
	Replace(State)
	Push(State)
	Undo() bool
	Redo() bool
	Undoable() bool
	Redoable() bool
}
