package widgets

import (
	g "github.com/AllenDang/giu"
)

type KeyBindsWidget struct {
	isGlobal bool
	widgets  []g.Widget
}

type KeyBindsFlags uint8

const (
	KeyBindsFlagNone   KeyBindFlags = 0
	KeyBindsFlagGlobal              = 1 << iota
)

func KeyBinds(flags KeyBindsFlags, widgets ...g.Widget) *KeyBindsWidget {
	return &KeyBindsWidget{
		isGlobal: flags&KeyBindsFlagGlobal != 0,
		widgets:  widgets,
	}
}

func (k *KeyBindsWidget) Build() {
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
	// Get all released and pressed keys so as to reduce keybind collisions.
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
	if k.isGlobal || g.IsItemHovered() {
		for _, w := range k.widgets {
			keyBindWidget, isKeyBind := w.(*KeyBindWidget)

			if isKeyBind {
				if keyBindWidget.onPressed {
					if len(keyBindWidget.keys) == len(pressedKeys) {
						keyBindWidget.Build()
					}
				}
				if keyBindWidget.onDown {
					if len(keyBindWidget.keys) == len(downKeys) {
						keyBindWidget.Build()
					}
				}
				if keyBindWidget.onReleased {
					if len(keyBindWidget.keys) == len(releasedKeys) {
						keyBindWidget.Build()
					}
				}
			} else {
				w.Build()
			}
		}
	}
}

type KeyBindWidget struct {
	onPressed  bool
	onDown     bool
	onReleased bool
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

func KeyBind(flags KeyBindFlags, keys []int, cb func()) *KeyBindWidget {
	return &KeyBindWidget{
		onPressed:  flags&KeyBindFlagPressed != 0,
		onDown:     flags&KeyBindFlagDown != 0,
		onReleased: flags&KeyBindFlagReleased != 0,
		keys:       keys,
		cb:         cb,
	}
}

func (k *KeyBindWidget) Build() {
	if k.onPressed {
		match := true
		for _, key := range k.keys {
			if !g.IsKeyPressed(key) {
				match = false
				break
			}
		}
		if match {
			k.cb()
		}
	}

	if k.onDown {
		match := true
		for _, key := range k.keys {
			if !g.IsKeyDown(key) {
				match = false
				break
			}
		}
		if match {
			k.cb()
		}
	}

	if k.onReleased {
		match := true
		for _, key := range k.keys {
			if !g.IsKeyReleased(key) {
				match = false
				break
			}
		}
		if match {
			k.cb()
		}
	}
}
