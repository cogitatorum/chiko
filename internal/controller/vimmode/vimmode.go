package vimmode

import (
	"github.com/felangga/chiko/internal/entity"
	"github.com/felangga/chiko/internal/events"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type VimMode struct {
	App        *tview.Application
	ActiveMode entity.VimMode
}

func (vm *VimMode) SetMode(mode entity.VimMode) {
	vm.ActiveMode = mode
	events.GetEventManager().Emit(&events.VimModeChanged{
		NewMode: mode,
	})
}

func NewVimMode(app *tview.Application) *VimMode {
	vm := new(VimMode)
	vm.App = app
	vm.SetMode(entity.VimModeNormal)

	return vm
}

// OnTextWidget reports whether the focused primitive is vim-managed.
func (vm *VimMode) OnTextWidget() bool {
	switch vm.App.GetFocus().(type) {
	case *tview.TextArea, *tview.InputField:
		return true
	default:
		return false
	}
}

// TranslateInput applies the active mode's keymap and nvim-style mode switches.
// Returns nil when the key was consumed (e.g. mode change).
func (vm *VimMode) TranslateInput(event *tcell.EventKey) *tcell.EventKey {
	if !vm.OnTextWidget() {
		return event
	}

	if handled := vm.handleModeSwitch(event); handled {
		return nil
	}

	// Ctrl-hjkl must be handled before the keymap: TextArea binds KeyCtrlL to
	// select-all (and KeyCtrlK to delete). On key-repeat that freezes the UI.
	// Modern terminals may also send KeyRune+'l'+ModCtrl instead of KeyCtrlL.
	if vm.ActiveMode == entity.VimModeNormal || vm.ActiveMode == entity.VimModeVisual {
		if ev := translateCtrlHJKL(event, vm.ActiveMode == entity.VimModeVisual); ev != nil {
			return ev
		}
	}

	switch vm.ActiveMode {
	case entity.VimModeNormal:
		return NormalKeyMap.GetBinding(event).To
	case entity.VimModeEdit:
		return EditKeyMap.GetBinding(event).To
	case entity.VimModeVisual:
		return VisualKeyMap.GetBinding(event).To
	}
	return event
}

// translateCtrlHJKL maps Ctrl-h/j/k/l to the same motions as h/j/k/l.
// Returns nil if the event is not a Ctrl-hjkl variant.
func translateCtrlHJKL(event *tcell.EventKey, visual bool) *tcell.EventKey {
	mod := tcell.ModNone
	if visual {
		mod = tcell.ModShift
	}

	switch event.Key() {
	case tcell.KeyCtrlH: // == KeyBackspace
		return tcell.NewEventKey(tcell.KeyLeft, 0, mod)
	case tcell.KeyCtrlJ:
		return tcell.NewEventKey(tcell.KeyDown, 0, mod)
	case tcell.KeyCtrlK:
		return tcell.NewEventKey(tcell.KeyUp, 0, mod)
	case tcell.KeyCtrlL:
		return tcell.NewEventKey(tcell.KeyRight, 0, mod)
	}

	if event.Key() == tcell.KeyRune && event.Modifiers()&tcell.ModCtrl != 0 {
		switch event.Rune() {
		case 'h', 'H':
			return tcell.NewEventKey(tcell.KeyLeft, 0, mod)
		case 'j', 'J':
			return tcell.NewEventKey(tcell.KeyDown, 0, mod)
		case 'k', 'K':
			return tcell.NewEventKey(tcell.KeyUp, 0, mod)
		case 'l', 'L':
			return tcell.NewEventKey(tcell.KeyRight, 0, mod)
		}
	}
	return nil
}

// handleModeSwitch implements Esc / i / a / v like nvim. Returns true if consumed.
func (vm *VimMode) handleModeSwitch(event *tcell.EventKey) bool {
	if event.Key() == tcell.KeyEscape {
		if vm.ActiveMode != entity.VimModeNormal {
			vm.SetMode(entity.VimModeNormal)
			return true
		}
		return false
	}

	// Mode-entry keys only apply from Normal (nvim: i/a/v).
	if vm.ActiveMode != entity.VimModeNormal || event.Key() != tcell.KeyRune {
		return false
	}

	switch event.Rune() {
	case 'i', 'a':
		vm.SetMode(entity.VimModeEdit)
		return true
	case 'v':
		vm.SetMode(entity.VimModeVisual)
		return true
	}
	return false
}
