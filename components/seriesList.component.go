package components

import (
	"errors"
	"reflect"
	"slices"
	"strconv"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	repo "github.com/palianskas/toire/internal"
	"github.com/palianskas/toire/models"
)

type SeriesListCtx struct {
	ContextBase
	items []table.Row
	pager paginator.Model
	table table.Model
	input textinput.Model
	conf  *ConfirmationCtx
	repo  *repo.Repo
}

func (ctx *SeriesListCtx) Init() tea.Cmd {
	return nil
}

func (ctx *SeriesListCtx) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)

	var cmd tea.Cmd

	if ok {
		cmd = ctx.handleKeyMsg(keyMsg)
	}

	ctx.setRenderedItems()

	return ctx, cmd
}

func (ctx *SeriesListCtx) View() string {
	ctx.paint("Series:")

	if len(ctx.items) > 0 {
		ctx.paint(ctx.table.View())
		ctx.paint(ctx.pager.View())
	} else {
		ctx.paint("\n<No series found>\n")
	}

	if ctx.conf != nil {
		ctx.paint(ctx.conf.View())
	}

	if ctx.input.Focused() {
		ctx.paint(ctx.input.View())
	} else {
		ctx.paint("Add series: <a>; Delete series: <D>; Up: <k>; Down: <j>; Next page: <l>; Prev. page: <h>; Quit: <q>")
	}

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}

func (ctx *SeriesListCtx) WakeUp() error {
	rows, err := ctx.getRows()
	ctx.items = rows

	if err != nil {
		return err
	}

	return nil
}

func NewSeriesList() SeriesListCtx {
	repo, err := repo.GetInstance(DB_PATH)

	HandleFatalErr(err)

	ctx := SeriesListCtx{
		repo:        repo,
		ContextBase: NewContextBase(reflect.TypeOf(SeriesListCtx{}).Name()),
	}

	p := paginator.New()
	ctx.pager = p

	p.Type = paginator.Arabic
	p.PerPage = 15

	rowCnt, err := ctx.repo.GetTableRowCount(ctx.repo.Tables.Series)

	HandleFatalErr(err)

	p.SetTotalPages(rowCnt)

	columns := []table.Column{
		{
			Title: "Id", Width: 0,
		},
		{
			Title: "Name", Width: 25,
		},
		{
			Title: "Seasons", Width: 10,
		},
	}

	var rows []table.Row
	rows, err = ctx.getRows()
	ctx.items = rows

	HandleFatalErr(err)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows[0:min(len(rows), p.PerPage)]),
		table.WithFocused(true),
	)

	ctx.table = t

	i := textinput.New()
	i.Validate = ctx.validateInput

	ctx.input = i

	return ctx
}

func (ctx *SeriesListCtx) updateTableSize() {
}

func (ctx *SeriesListCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if ctx.conf != nil {
		var cmd tea.Cmd
		ctx.conf, cmd = ctx.conf.Update(msg)

		if cmd == nil {
			return nil
		}

		confMsg, ok := cmd().(ConfirmationMsg)

		if !ok {
			return cmd
		}

		ctx.conf = nil

		if confMsg.state != Confirmed {
			return nil
		}

		ctx.err = ctx.deleteSelectedSeries()

		return nil
	}

	if !ctx.input.Focused() {
		switch msg.String() {
		case "a", "A":
			if !ctx.input.Focused() {
				ctx.input.Focus()
				return nil
			}
		case "h", "l":
			ctx.table.SetCursor(0)
		case "D":
			if ctx.conf == nil {
				c := NewConfirmation()
				ctx.conf = &c
				return nil
			}
		case "enter", " ":
			row := ctx.table.SelectedRow()

			id, err := strconv.Atoi(row[0])

			if err != nil {
				ctx.err = err
				return nil
			}

			series, err := ctx.repo.GetSeries(id)

			if err != nil {
				ctx.err = err
				return nil
			}

			series.Seasons, err = ctx.repo.GetSeasonsBySeries(id)

			if err != nil {
				ctx.err = err
				return nil
			}

			infoCtx := NewSeriesInfo(series)

			return CtxPushCmd(&infoCtx)
		}
	} else {
		switch msg.String() {
		case "esc":
			ctx.input.Blur()
			ctx.input.Reset()
			ctx.table.Focus()
		case "enter":
			ctx.input.Err = ctx.input.Validate(ctx.input.Value())
			if ctx.input.Err != nil {
				ctx.err = ctx.input.Err
				ctx.input.Focus()
				return nil
			}

			model := &models.Series{
				Name: ctx.input.Value(),
			}

			ctx.err = ctx.repo.SaveSeries(model)

			if ctx.err != nil {
				return nil
			}

			model, ctx.err = ctx.repo.GetSeries(model.Id)

			if ctx.err != nil {
				return nil
			}

			ctx.items = slices.Insert(ctx.items, 0, toSeriesRow(*model))

			ctx.input.Blur()
			ctx.input.Reset()
		}
	}

	cmd := ctx.handleBaseKeyMsg(msg)
	if cmd != nil {
		return cmd
	}

	cmds := []tea.Cmd{}
	if ctx.input.Focused() {
		var icmd tea.Cmd
		ctx.input, icmd = ctx.input.Update(msg)
		cmds = append(cmds, icmd)
	} else {
		var pcmd tea.Cmd
		var tcmd tea.Cmd
		ctx.pager, pcmd = ctx.pager.Update(msg)
		ctx.table, tcmd = ctx.table.Update(msg)

		cmds = append(cmds, pcmd)
		cmds = append(cmds, tcmd)
	}

	return tea.Batch(cmds...)
}

func (ctx *SeriesListCtx) setRenderedItems() {
	ctx.pager.SetTotalPages(len(ctx.items))
	l, r := ctx.pager.GetSliceBounds(len(ctx.items))
	ctx.table.SetRows(ctx.items[l:r])
}

func (ctx *SeriesListCtx) getRows() ([]table.Row, error) {
	var rows []table.Row

	series, err := ctx.repo.GetSeriesMany()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(*series, func(x models.Series, y models.Series) int { return -x.CreateDate.Compare(y.CreateDate) })

	for i := 0; i < len(*series); i++ {
		rows = append(rows, toSeriesRow((*series)[i]))
	}

	return rows, err

}

func toSeriesRow(series models.Series) table.Row {
	return table.Row{
		strconv.Itoa(series.Id),
		series.Name,
		strconv.Itoa(series.SeasonsCnt),
	}
}

func (ctx *SeriesListCtx) validateInput(input string) error {
	if len([]rune(input)) > 0 {
		return nil
	}

	return errors.New("Input cannot be empty!")
}

func (ctx *SeriesListCtx) deleteSelectedSeries() error {
	row := ctx.table.SelectedRow()

	id, err := strconv.Atoi(row[0])

	if err != nil {
		return err
	}

	err = ctx.repo.DeleteSeries(id)

	if err != nil {
		ctx.err = err
		return err
	}

	idx := -1
	for i := 0; i < len(ctx.items); i++ {
		if ctx.items[i][0] == row[0] {
			idx = i
			break
		}
	}

	if idx >= 0 {
		ctx.items = append(ctx.items[:idx], ctx.items[idx+1:]...)
	}

	return nil

}
