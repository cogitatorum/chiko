package vimmode

import (
	"github.com/felangga/chiko/internal/entity"
	"github.com/felangga/chiko/internal/events"
)

type VimMode struct {
	ActiveMode entity.VimMode
}

func (vm *VimMode) SetMode(mode entity.VimMode) {
	vm.ActiveMode = mode
	events.GetEventManager().Emit(&events.VimModeChanged{
		NewMode: mode,
	})
}

func NewVimMode() *VimMode {
	vm := new(VimMode)
	vm.SetMode(entity.VimModeNormal)

	return vm
}
