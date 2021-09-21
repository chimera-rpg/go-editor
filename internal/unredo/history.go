package unredo

// History maintains a stack of States.
type History struct {
	index int
	stack []State
}

// NewUnredoabler returns an Unredoabler interface.
func NewUnredoabler(s State) Unredoabler {
	return &History{
		index: 0,
		stack: []State{s},
	}
}

// State returns the current underlying state.
func (h *History) State() State {
	if h.index >= 0 && h.index < len(h.stack) {
		return h.stack[h.index]
	}
	return nil
}

// Push pushes a new state onto the stack.
func (h *History) Push(s State) {
	h.stack = append(h.stack[:h.index+1], s)
	h.index++
}

// Replace replaces the current state.
func (h *History) Replace(s State) {
	h.stack[h.index] = s
}

// Redo redoes a state if possible.
func (h *History) Redo() bool {
	if !h.Redoable() {
		return false
	}
	h.index++
	return true
}

// Undo undoes a state if possible.
func (h *History) Undo() bool {
	if !h.Undoable() {
		return false
	}
	h.index--
	return true
}

// Undoable returns if the history can be undone.
func (h *History) Undoable() bool {
	if h.index > 0 && len(h.stack) > 0 {
		return true
	}
	return false
}

// Redoable returns if the history can be redone.
func (h *History) Redoable() bool {
	if h.index < len(h.stack)-1 {
		return true
	}
	return false
}
