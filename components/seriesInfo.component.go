package components

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	repo "github.com/palianskas/toire/internal"
	"github.com/palianskas/toire/models"
)

type SeriesInfoCtx struct {
	series *models.Series
	pager  paginator.Model
	table  table.Model
	repo   *repo.Repo
	ContextBase
}

var seriesInfoColumns = []table.Column{
	{
		Title: "Id", Width: 0,
	},
	{
		Title: "Number", Width: 10,
	},
	{
		Title: "Episodes", Width: 10,
	},
	{
		Title: "Name", Width: 80,
	},
}

func NewSeriesInfo(series *models.Series) SeriesInfoCtx {
	repo, err := repo.GetInstance(DB_PATH)

	HandleFatalErr(err)

	ctx := SeriesInfoCtx{
		series:      series,
		repo:        repo,
		ContextBase: NewContextBase(reflect.TypeOf(SeriesInfoCtx{}).Name()),
	}

	ctx.pager = paginator.New()
	ctx.pager.Type = paginator.Arabic
	ctx.pager.PerPage = 15

	rows := ctx.getTablePage()

	HandleFatalErr(err)

	ctx.table = table.New(
		table.WithColumns(seriesInfoColumns),
		table.WithRows(rows[0:min(len(rows), ctx.pager.PerPage)]),
		table.WithFocused(true),
	)

	ctx.updateTableSize()
	ctx.setTableRows()

	return ctx
}

func (ctx *SeriesInfoCtx) Init() tea.Cmd {
	return nil
}

func (ctx *SeriesInfoCtx) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)

	var cmd tea.Cmd = nil

	if ok {
		cmd = ctx.handleKeyMsg(keyMsg)
	}

	ctx.sortSeriesSeasons()

	ctx.setTableRows()

	ctx.updateTableSize()

	return ctx, cmd
}

func (ctx *SeriesInfoCtx) updateTableSize() {
	reservedHeight := 6
	if ctx.err != nil {
		reservedHeight += 2
	}

	ctx.table.SetHeight(WIN_HEIGHT - reservedHeight)
	ctx.table.SetWidth(WIN_WIDTH)
	ctx.pager.PerPage = ctx.table.Height()
	ctx.pager.SetTotalPages(ctx.series.SeasonsCnt)
}

func (ctx *SeriesInfoCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	cmd = ctx.handleBaseKeyMsg(msg)

	if cmd != nil {
		return cmd
	}

	switch msg.String() {
	case `a`, "A":
		addSeasonCtx := NewAddSeason(ctx.series)
		return CtxPushCmd(&addSeasonCtx)
	case "D":
		ctx.err = ctx.deleteSelected()
	case "enter", " ":
		cmd, ctx.err = ctx.openSelected()
		if cmd != nil {
			return cmd
		}
	}

	cmds := []tea.Cmd{}
	var pcmd tea.Cmd
	var tcmd tea.Cmd
	ctx.pager, pcmd = ctx.pager.Update(msg)
	ctx.table, tcmd = ctx.table.Update(msg)

	cmds = append(cmds, pcmd)
	cmds = append(cmds, tcmd)

	wrap := min(ctx.pager.ItemsOnPage(ctx.series.SeasonsCnt), ctx.pager.PerPage)
	wrap = max(wrap, 1)

	return tea.Batch(cmds...)
}

func (ctx *SeriesInfoCtx) View() string {
	ctx.paint(fmt.Sprintf("Series: %v", ctx.series.Name))
	ctx.paint("Seasons:")

	if ctx.series.SeasonsCnt > 0 {
		ctx.paint(ctx.table.View())
		ctx.paint(ctx.pager.View())
	} else {
		ctx.paint("\n<No seasons found>\n")
	}

	ctx.paint(`Add season: <a>; Delete season: <D>; Up: <k>; Down: <j>; Next page: <l>; Prev. page: <h>; Quit: <q>`)

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}

func (ctx *SeriesInfoCtx) WakeUp() error {
	ctx.series, ctx.err = ctx.repo.GetSeries(ctx.series.Id)

	if ctx.err != nil {
		return ctx.err
	}

	ctx.series.Seasons, ctx.err = ctx.repo.GetSeasonsBySeries(ctx.series.Id)
	ctx.sortSeriesSeasons()

	return ctx.err
}

func (ctx *SeriesInfoCtx) getTablePage() []table.Row {
	page := ctx.pager.Page
	size := ctx.pager.PerPage
	offset := page * size
	rows := make([]table.Row, min(size, ctx.series.SeasonsCnt))

	for i := offset; i < min(offset+size, ctx.series.SeasonsCnt); i++ {
		rows = append(rows, toSeasonRow((*ctx.series.Seasons)[i]))
	}

	return rows
}

func toSeasonRow(season models.Season) table.Row {
	return table.Row{
		strconv.Itoa(season.Id),
		strconv.Itoa(season.Number),
		strconv.Itoa(season.EpisodeCnt),
		season.Name,
	}
}

func (ctx *SeriesInfoCtx) setTableRows() {
	ctx.pager.SetTotalPages(ctx.series.SeasonsCnt)
	l, r := ctx.pager.GetSliceBounds(ctx.series.SeasonsCnt)

	rows := make([]table.Row, r-l)
	for i, season := range (*ctx.series.Seasons)[l:r] {
		rows[i] = toSeasonRow(season)
	}

	ctx.table.SetRows(rows)
}

func (ctx *SeriesInfoCtx) deleteSelected() error {
	row := ctx.table.SelectedRow()

	id, err := strconv.Atoi(row[0])

	if err != nil {
		return err
	}

	err = ctx.repo.DeleteSeason(id)

	if err != nil {
		return err
	}

	ctx.series.Seasons, err = ctx.repo.GetSeasonsBySeries(ctx.series.Id)
	ctx.sortSeriesSeasons()

	if err != nil {
		return err
	}

	ctx.series.SeasonsCnt = len(*ctx.series.Seasons)
	ctx.table.SetCursor(ctx.table.Cursor() - 1)

	return nil
}

func (ctx *SeriesInfoCtx) openSelected() (tea.Cmd, error) {
	row := ctx.table.SelectedRow()

	id, err := strconv.Atoi(row[0])

	if err != nil {
		return nil, err
	}

	season, err := ctx.repo.GetSeason(id)

	if err != nil {
		return nil, err
	}

	season.Episodes, err = ctx.repo.GetEpisodesBySeason(id)

	if err != nil {
		return nil, err
	}

	infoCtx := NewSeasonInfo(season)

	return CtxPushCmd(&infoCtx), nil
}

func (ctx *SeriesInfoCtx) sortSeriesSeasons() {
	slices.SortFunc(*ctx.series.Seasons, func(x models.Season, y models.Season) int { return x.CreateDate.Compare(y.CreateDate) })
}
