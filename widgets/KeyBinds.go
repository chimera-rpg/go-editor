package widgets

import (
	g "github.com/AllenDang/giu"
)

type KeyBindsWidget struct {
	onWindowFocused, onWindowHovered bool
	onItemActive, onItemHovered      bool
	widgets                          []g.Widget
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
		onWindowFocused: flags&KeyBindsFlagWindowFocused != 0,
		onWindowHovered: flags&KeyBindsFlagWindowHovered != 0,
		onItemActive:    flags&KeyBindsFlagItemActive != 0,
		onItemHovered:   flags&KeyBindsFlagItemHovered != 0,
		widgets:         widgets,
	}
}

func (k *KeyBindsWidget) Build() {
	if k.onWindowFocused && !g.IsWindowFocused(g.FocusedFlagsChildWindows) {
		return
	}
	if k.onWindowHovered && !g.IsWindowHovered(g.HoveredFlagsChildWindows) {
		return
	}
	if k.onItemActive && !g.IsItemActive() {
		return
	}
	if k.onItemHovered && !g.IsItemHovered() {
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
			if keyBindWidget.onPressed {
				if len(keyBindWidget.keys) == len(pressedKeys) && len(keyBindWidget.modifiers) == len(downModifiers) {
					keyBindWidget.Build()
				}
			}
			if keyBindWidget.onDown {
				if len(keyBindWidget.keys) == len(downKeys) && len(keyBindWidget.modifiers) == len(downModifiers) {
					keyBindWidget.Build()
				}
			}
			if keyBindWidget.onReleased {
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
	onPressed  bool
	onDown     bool
	onReleased bool
	modifiers  []int
	keys       []int
	cb         func()
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
		onPressed:  flags&KeyBindFlagPressed != 0,
		onDown:     flags&KeyBindFlagDown != 0,
		onReleased: flags&KeyBindFlagReleased != 0,
		modifiers:  modifiers,
		keys:       keys,
		cb:         cb,
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

	if k.onPressed {
		if keysDown(k.modifiers) && keysPressed(k.keys) {
			k.cb()
		}
	}

	if k.onDown {
		if keysDown(k.modifiers) && keysDown(k.keys) {
			k.cb()
		}
	}

	if k.onReleased {
		if keysDown(k.modifiers) && keysReleased(k.keys) {
			k.cb()
		}
	}
}

func Keys(keys ...int) []int {
	return keys
}
