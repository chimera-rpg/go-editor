package widgets

import (
	g "github.com/AllenDang/giu"
)

type KeyBindsWidget struct {
	flags   KeyBindsFlags
	widgets []g.Widget
}

type KeyBindsFlags uint8

const (
	KeyBindsFlagNone          KeyBindFlags = 0
	KeyBindsFlagWindowFocused              = 1 << iota
	KeyBindsFlagWindowHovered
	KeyBindsFlagItemActive
	KeyBindsFlagItemHovered
)

func KeyBinds(flags KeyBindsFlags, widgets ...g.Widget) *KeyBindsWidget {
	return &KeyBindsWidget{
		flags:   flags,
		widgets: widgets,
	}
}

func (k *KeyBindsWidget) Build() {
	if k.flags&KeyBindsFlagWindowFocused != 0 && !g.IsWindowFocused(g.FocusedFlagsChildWindows) {
		return
	}
	if k.flags&KeyBindsFlagWindowHovered != 0 && !g.IsWindowHovered(g.HoveredFlagsChildWindows) {
		return
	}
	if k.flags&KeyBindsFlagItemActive != 0 && !g.IsItemActive() {
		return
	}
	if k.flags&KeyBindsFlagItemHovered != 0 && !g.IsItemHovered() {
		return
	}

	// Collect all bound keys.
	boundKeys := make(map[int]struct{})
	for _, w := range k.widgets {
		keyBindWidget, isKeyBind := w.(*KeyBindWidget)
		if isKeyBind {
			for _, key := range keyBindWidget.keys {
				boundKeys[key] = struct{}{}
			}
		}
	}
	// Collect all bound modifiers.
	boundModifiers := make(map[int]struct{})
	for _, w := range k.widgets {
		keyBindWidget, isKeyBind := w.(*KeyBindWidget)
		if isKeyBind {
			for _, key := range keyBindWidget.modifiers {
				boundModifiers[key] = struct{}{}
			}
		}
	}

	// Get all released and pressed keys so as to reduce keybind collisions.
	var downModifiers []int
	for key := range boundModifiers {
		if g.IsKeyDown(key) {
			downModifiers = append(downModifiers, key)
		}
	}

	var pressedKeys, releasedKeys, downKeys []int
	for key := range boundKeys {
		if g.IsKeyPressed(key) {
			pressedKeys = append(pressedKeys, key)
		} else if g.IsKeyReleased(key) {
			releasedKeys = append(releasedKeys, key)
		}
		if g.IsKeyDown(key) {
			downKeys = append(downKeys, key)
		}
	}
	// Iterate through all widgets.
	for _, w := range k.widgets {
		keyBindWidget, isKeyBind := w.(*KeyBindWidget)

		if isKeyBind {
			if keyBindWidget.flags&KeyBindFlagPressed != 0 {
				if len(keyBindWidget.keys) == len(pressedKeys) && len(keyBindWidget.modifiers) == len(downModifiers) {
					keyBindWidget.Build()
				}
			}
			if keyBindWidget.flags&KeyBindFlagDown != 0 {
				if len(keyBindWidget.keys) == len(downKeys) && len(keyBindWidget.modifiers) == len(downModifiers) {
					keyBindWidget.Build()
				}
			}
			if keyBindWidget.flags&KeyBindFlagReleased != 0 {
				if len(keyBindWidget.keys) == len(releasedKeys) && len(keyBindWidget.modifiers) == len(downModifiers) {
					keyBindWidget.Build()
				}
			}
		} else {
			w.Build()
		}
	}
}

type KeyBindWidget struct {
	flags     KeyBindFlags
	modifiers []int
	keys      []int
	cb        func()
}

type KeyBindFlags uint8

const (
	KeyBindFlagNone    KeyBindFlags = 0
	KeyBindFlagPressed              = 1 << iota
	KeyBindFlagReleased
	KeyBindFlagDown
)

func KeyBind(flags KeyBindFlags, modifiers []int, keys []int, cb func()) *KeyBindWidget {
	return &KeyBindWidget{
		flags:     flags,
		modifiers: modifiers,
		keys:      keys,
		cb:        cb,
	}
}

func (k *KeyBindWidget) Build() {
	keysDown := func(keys []int) bool {
		for _, key := range keys {
			if !g.IsKeyDown(key) {
				return false
			}
		}
		return true
	}
	keysPressed := func(keys []int) bool {
		for _, key := range keys {
			if !g.IsKeyPressed(key) {
				return false
			}
		}
		return true
	}
	keysReleased := func(keys []int) bool {
		for _, key := range keys {
			if !g.IsKeyReleased(key) {
				return false
			}
		}
		return true
	}

	if k.flags&KeyBindFlagPressed != 0 {
		if keysDown(k.modifiers) && keysPressed(k.keys) {
			k.cb()
		}
	}

	if k.flags&KeyBindFlagDown != 0 {
		if keysDown(k.modifiers) && keysDown(k.keys) {
			k.cb()
		}
	}

	if k.flags&KeyBindFlagReleased != 0 {
		if keysDown(k.modifiers) && keysReleased(k.keys) {
			k.cb()
		}
	}
}

type Key int

func Keys(keys ...int) []int {
	return keys
}

const (
	Key0 = iota + 48
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
)
const (
	KeyA = iota + 65
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
)
const (
	KeyEscape = iota + 256
	KeyEnter
	KeyTab
	KeyBackspace
	KeyInsert
	KeyDelete
	KeyRight
	KeyLeft
	KeyDown
	KeyUp
	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd
)
const (
	KeyLeftShift = iota + 340
	KeyLeftControl
	KeyLeftAlt
	KeyLeftSuper
	KeyRightShift
	KeyRightControl
	KeyRightAlt
	KeyRightSuper
	KeyMenu
	KeyShift   = KeyLeftShift
	KeyAlt     = KeyLeftAlt
	KeyControl = KeyLeftControl
)
