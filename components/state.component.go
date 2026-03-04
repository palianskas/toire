package components

import (
	"fmt"
	"log"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gap "github.com/muesli/go-app-paths"
	repo "github.com/palianskas/toire/internal"
)

var APP_DIR string
var DB_PATH string
var WIN_HEIGHT int
var WIN_WIDTH int

func HandleFatalErr(err error) {
	if err != nil {
		panic(fmt.Sprintf(`Encountered fatal error: %v`, err))
	}
}

type CtxPushMsg struct {
	ctx Context
}

func CtxPushCmd(ctx Context) tea.Cmd {
	return func() tea.Msg {
		return CtxPushMsg{
			ctx: ctx,
		}
	}
}

type CtxPopMsg struct{}

func CtxPopCmd() tea.Cmd {
	return func() tea.Msg {
		return CtxPopMsg{}
	}
}

type State struct {
	ctxStack []Context
}

func (state *State) Init() tea.Cmd {
	return nil
}

func (state *State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	state.ensureCtx()

	_, ok := msg.(CtxPopMsg)
	if ok && len(state.ctxStack) > 1 {
		state.popCtx()
		state.ensureCtx()
		state.ctxStack[len(state.ctxStack)-1].WakeUp()
	}

	pushMsg, ok := msg.(CtxPushMsg)
	if ok {
		state.pushCtx(pushMsg.ctx)
	}

	wsMsg, ok := msg.(tea.WindowSizeMsg)
	if ok {
		state.handleResizeMsg(wsMsg)
	}

	ctx := state.ctxStack[len(state.ctxStack)-1]
	_, cmd := ctx.Update(msg)

	return state, cmd
}

func (state *State) View() string {
	state.ensureCtx()
	ctx := state.ctxStack[len(state.ctxStack)-1]
	return ctx.View()
}

func InitState() *State {
	APP_DIR = initAppDir()
	_, err := initRepo()

	HandleFatalErr(err)

	state := State{
		ctxStack: []Context{},
	}

	ctx := NewOverview()

	state.pushCtx(&ctx)

	return &state
}

func initAppDir() string {
	scope := gap.NewScope(gap.User, "toire")

	ddirs, err := scope.DataDirs()

	HandleFatalErr(err)

	path := strings.Trim(ddirs[0], " ")
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		err = os.Mkdir(path, 0o770)
	}

	HandleFatalErr(err)

	return path
}

func initRepo() (*repo.Repo, error) {
	DB_PATH = APP_DIR + "database.db"

	repo, err := repo.Open(DB_PATH)

	if err != nil {
		return nil, err
	}

	err = repo.Init()

	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (state *State) ensureCtx() {
	if len(state.ctxStack) < 1 {
		log.Fatal("State: Context stack is empty")
	}
}

func (state *State) pushCtx(ctx Context) {
	state.ctxStack = append(state.ctxStack, ctx)
}

func (state *State) popCtx() error {
	if len(state.ctxStack) < 1 {
		return fmt.Errorf("Attempting to pop from empty stack")
	}

	state.ctxStack = state.ctxStack[:len(state.ctxStack)-1]

	return nil
}

func (state *State) handleResizeMsg(msg tea.WindowSizeMsg) {
	WIN_HEIGHT = msg.Height
	WIN_WIDTH = msg.Width
}
