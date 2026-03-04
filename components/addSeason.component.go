package components

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/palianskas/toire/internal"
	repo "github.com/palianskas/toire/internal"
	"github.com/palianskas/toire/models"
)

type AddSeasonCtx struct {
	season            *models.Season
	series            *models.Series
	nameInput         textinput.Model
	filePicker        *FilePickerCtx
	filePickerFocused bool
	repo              *repo.Repo
	ContextBase
}

func NewAddSeason(series *models.Series) AddSeasonCtx {
	repo, err := repo.GetInstance(DB_PATH)

	HandleFatalErr(err)

	var path string
	path, err = os.UserHomeDir()

	HandleFatalErr(err)

	picker := NewFilePicker(&path, true)

	ctx := AddSeasonCtx{
		series:      series,
		season:      &models.Season{Number: series.SeasonsCnt + 1, SeriesId: series.Id},
		nameInput:   textinput.New(),
		filePicker:  &picker,
		repo:        repo,
		ContextBase: NewContextBase(reflect.TypeOf(AddSeasonCtx{}).Name()),
	}

	ctx.nameInput.Placeholder = "(empty)"
	ctx.nameInput.Focus()

	return ctx
}

func (ctx *AddSeasonCtx) Init() tea.Cmd {
	return ctx.filePicker.Init()
}

func (ctx *AddSeasonCtx) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)

	var cmd tea.Cmd = nil

	if ok {
		cmd = ctx.handleKeyMsg(keyMsg)
	}

	fileMsg, ok := msg.(FilePickedMsg)

	if ok {
		resolver := internal.Resolver{}
		eps, err := resolver.ReadDir(fileMsg.Path)

		if err != nil {
			ctx.err = err
			return ctx, nil
		}

		name := ctx.nameInput.Value()

		if len(name) < 1 {
			pathTokens := strings.Split(fileMsg.Path, "/")
			name = pathTokens[len(pathTokens)-1]
		}

		season := models.Season{
			Name:     name,
			Number:   ctx.series.SeasonsCnt + 1,
			SeriesId: ctx.series.Id,
		}

		err = ctx.repo.SaveSeason(&season)

		if err != nil {
			ctx.err = err
			return ctx, nil
		}

		slices.Sort(eps)

		for idx, epPath := range eps {
			pathTokens := strings.Split(epPath, "/")
			ep := models.Episode{
				Name:     pathTokens[len(pathTokens)-1],
				Number:   idx + 1,
				Path:     epPath,
				IsPlayed: false,
				SeasonId: season.Id,
			}

			ctx.repo.SaveEpisode(&ep)
		}

		return ctx, CtxPopCmd()
	}

	ctx.updateTableSize()

	return ctx, cmd
}

func (ctx *AddSeasonCtx) updateTableSize() {
	reservedHeight := 7
	if ctx.err != nil {
		reservedHeight += 2
	}
}

func (ctx *AddSeasonCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	cmd = ctx.handleBaseKeyMsg(msg)

	if cmd != nil {
		return cmd
	}

	switch msg.String() {
	case "tab", "shift+tab", "up", "down":
		if ctx.nameInput.Focused() {
			ctx.nameInput.Blur()
			ctx.filePicker.Focus()
		} else if ctx.filePicker.Focused() {
			ctx.nameInput.Focus()
			ctx.filePicker.Blur()
		}
	case "enter":
		if ctx.nameInput.Focused() {
			ctx.nameInput.Blur()
			ctx.filePicker.Focus()
		}
	}

	if ctx.nameInput.Focused() {
		var iCmd tea.Cmd
		ctx.nameInput, iCmd = ctx.nameInput.Update(msg)
		cmd = tea.Batch(cmd, iCmd)
	}

	if ctx.filePicker.Focused() {
		var pCmd tea.Cmd
		ctx.filePicker, pCmd = ctx.filePicker.Update(msg)
		cmd = tea.Batch(cmd, pCmd)
	}

	return cmd
}

func (ctx *AddSeasonCtx) View() string {
	ctx.paint(fmt.Sprintf("%v", ctx.series.Name))
	ctx.paint(`Add season`)

	ctx.paint(`Name:`)
	ctx.paint(ctx.nameInput.View())
	ctx.paint(ctx.filePicker.View())

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}
