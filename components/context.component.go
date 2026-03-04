package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Context interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	WakeUp() error
	// ShouldPop() bool
	// SetState(state *State)
}

type ContextBase struct {
	name          string
	stringBuilder *strings.Builder
	debugLog      *[]string
	err           error
}

func NewContextBase(name string) ContextBase {
	return ContextBase{
		name:          name,
		err:           nil,
		debugLog:      &([]string{}),
		stringBuilder: &(strings.Builder{}),
	}
}

func (ctx ContextBase) WakeUp() error {
	return nil
}

func (ctx ContextBase) handleBaseKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", `q`:
		return tea.Quit
	case "esc":
		return CtxPopCmd()
	}

	return nil
}

func (ctx ContextBase) debug(value ...any) {
	sval := []string{}
	for _, v := range value {
		sval = append(sval, fmt.Sprintf("%v", v))
	}

	*ctx.debugLog = append(*ctx.debugLog, strings.Join(sval, ` `))
}

func (ctx ContextBase) paint(value string) {
	if ctx.stringBuilder.Len() > 0 {
		ctx.stringBuilder.WriteString("\n")
	}
	ctx.stringBuilder.WriteString(value)
}

func (ctx ContextBase) paintDebug() {
	if len(*ctx.debugLog) < 1 {
		return
	}

	ctx.paint(fmt.Sprintf("%v debug:", ctx.name))

	limit := 5
	if len(*ctx.debugLog) >= limit {
		*ctx.debugLog = (*ctx.debugLog)[len(*ctx.debugLog)-limit:]
	}

	for _, line := range *ctx.debugLog {
		ctx.paint(line)
	}
}

func (ctx ContextBase) paintError() {
	if ctx.err == nil {
		return
	}

	ctx.paint(fmt.Sprintf("%v error:", ctx.name))
	ctx.paint(error.Error(ctx.err))
}

func (ctx ContextBase) render() string {
	res := ctx.stringBuilder.String()
	ctx.stringBuilder.Reset()

	return res
}
