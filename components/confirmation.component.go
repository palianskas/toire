package components

import (
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmationState int

const (
	Pending ConfirmationState = iota
	Confirmed
	Rejected
)

type ConfirmationMsg struct {
	state ConfirmationState
}

func ConfirmationCmd(state ConfirmationState) tea.Cmd {
	return func() tea.Msg {
		return ConfirmationMsg{state: state}
	}
}

type ConfirmationCtx struct {
	ContextBase
	state         ConfirmationState
	input         textinput.Model
	stringBuilder strings.Builder
}

func NewConfirmation() ConfirmationCtx {
	input := textinput.New()
	input.Focus()
	return ConfirmationCtx{
		state:       Pending,
		input:       input,
		ContextBase: NewContextBase(reflect.TypeOf(ConfirmationCtx{}).Name()),
	}
}

func (ctx *ConfirmationCtx) Init() tea.Cmd {
	return nil
}

func (ctx *ConfirmationCtx) View() string {
	ctx.paint("Confirm? Y/n")
	if ctx.input.Focused() {
		ctx.paint(ctx.input.View())
	}

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}

func (ctx *ConfirmationCtx) Update(msg tea.Msg) (*ConfirmationCtx, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)

	var cmd tea.Cmd = nil

	if ok {
		cmd = ctx.handleKeyMsg(keyMsg)
	}

	return ctx, cmd
}

func (ctx *ConfirmationCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if !ctx.input.Focused() {
		ctx.input.Focus()
		return nil
	}

	switch msg.String() {
	case "enter":
		val := strings.ToLower(strings.TrimSpace(ctx.input.Value()))
		if len(val) < 1 || val[0] == 'y' {
			ctx.state = Confirmed
		} else if val[0] == 'n' {
			ctx.state = Rejected
		}
	case "esc":
		ctx.state = Rejected
	}

	if ctx.state != Pending {
		return ConfirmationCmd(ctx.state)
	}

	var cmd tea.Cmd
	ctx.input, cmd = ctx.input.Update(msg)
	return cmd
}
