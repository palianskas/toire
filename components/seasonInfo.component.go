package components

import (
	"errors"
	"fmt"
	"os/exec"
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

type SeasonInfoCtx struct {
	season *models.Season
	input  textinput.Model
	pager  paginator.Model
	table  table.Model
	repo   *repo.Repo
	ContextBase
}

var seasonInfoColumns = []table.Column{
	{
		Title: "Id", Width: 0,
	},
	{
		Title: "Number", Width: 10,
	},
	{
		Title: `Played`, Width: 10,
	},
	{
		Title: "Name", Width: 50,
	},
	{
		Title: "Path", Width: 50,
	},
}

func NewSeasonInfo(season *models.Season) SeasonInfoCtx {
	repo, err := repo.GetInstance(DB_PATH)

	HandleFatalErr(err)

	ctx := SeasonInfoCtx{
		season:      season,
		input:       textinput.New(),
		repo:        repo,
		ContextBase: NewContextBase(reflect.TypeOf(SeasonInfoCtx{}).Name()),
	}

	ctx.input.Placeholder = "(empty)"

	ctx.pager = paginator.New()
	ctx.pager.Type = paginator.Arabic
	ctx.pager.PerPage = 15

	rows := ctx.getTablePage()

	HandleFatalErr(err)

	ctx.table = table.New(
		table.WithColumns(seasonInfoColumns),
		table.WithRows(rows[0:min(len(rows), ctx.pager.PerPage)]),
		table.WithFocused(true),
	)

	ctx.updateTableSize()
	ctx.setTableRows()

	return ctx
}

func (ctx *SeasonInfoCtx) Init() tea.Cmd {
	return nil
}

func (ctx *SeasonInfoCtx) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)

	var cmd tea.Cmd = nil

	if ok {
		cmd = ctx.handleKeyMsg(keyMsg)
	}

	ctx.setTableRows()

	ctx.updateTableSize()

	return ctx, cmd
}

func (ctx *SeasonInfoCtx) updateTableSize() {
	reservedHeight := 7
	if ctx.err != nil {
		reservedHeight += 2
	}

	ctx.table.SetHeight(WIN_HEIGHT - reservedHeight)
	ctx.table.SetWidth(WIN_WIDTH)
	ctx.pager.PerPage = ctx.table.Height()
	ctx.pager.SetTotalPages(ctx.season.EpisodeCnt)
}

func (ctx *SeasonInfoCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	cmd := ctx.handleBaseKeyMsg(msg)

	if cmd != nil {
		return cmd
	}

	switch msg.String() {
	case "enter":
		ctx.err = ctx.playSelected()
		return nil
	case "D":
		ctx.err = ctx.deleteSelected()
		return nil
	case "t":
		ctx.err = ctx.toggleIsPlayedSelected()
		return nil
	}

	cmds := []tea.Cmd{}
	var pcmd tea.Cmd
	var tcmd tea.Cmd
	ctx.pager, pcmd = ctx.pager.Update(msg)
	ctx.table, tcmd = ctx.table.Update(msg)

	cmds = append(cmds, pcmd)
	cmds = append(cmds, tcmd)

	wrap := min(ctx.pager.ItemsOnPage(ctx.season.EpisodeCnt), ctx.pager.PerPage)
	wrap = max(wrap, 1)

	return tea.Batch(cmds...)
}

func (ctx *SeasonInfoCtx) View() string {
	ctx.paint(fmt.Sprintf("%v", ctx.season.Series.Name))
	ctx.paint(fmt.Sprintf("Season %v: %v", ctx.season.Number, ctx.season.Name))
	ctx.paint("Episodes:")

	if ctx.season.EpisodeCnt > 0 {
		ctx.paint(ctx.table.View())
		ctx.paint(ctx.pager.View())
	} else {
		ctx.paint("\n<No episodes found>\n")
	}

	if ctx.input.Focused() {
		ctx.paint(ctx.input.View())
	} else {
		ctx.paint(`Play: <Enter>; Toggle status: <t>; Delete episode: <D>; Up: <k>; Down: <j>; Next page: <l>; Prev. page: <h>; Quit: <q>`)
	}

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}

func (ctx *SeasonInfoCtx) WakeUp() error {
	return nil
}

func (ctx *SeasonInfoCtx) getTablePage() []table.Row {
	page := ctx.pager.Page
	size := ctx.pager.PerPage
	offset := page * size

	rows := make([]table.Row, min(size, ctx.season.EpisodeCnt))

	for i := offset; i < min(offset+size, ctx.season.EpisodeCnt); i++ {
		rows = append(rows, toEpisodeRow((*ctx.season.Episodes)[i]))
	}

	return rows
}

func toEpisodeRow(episode models.Episode) table.Row {
	isPlayed := `-`
	if episode.IsPlayed {
		isPlayed = "+"
	}
	return table.Row{
		strconv.Itoa(episode.Id),
		strconv.Itoa(episode.Number),
		isPlayed,
		episode.Name,
		episode.Path,
	}
}

func (ctx *SeasonInfoCtx) setTableRows() {
	ctx.pager.SetTotalPages(ctx.season.EpisodeCnt)
	l, r := ctx.pager.GetSliceBounds(ctx.season.EpisodeCnt)

	rows := make([]table.Row, r-l)
	for i, episode := range (*ctx.season.Episodes)[l:r] {
		rows[i] = toEpisodeRow(episode)
	}

	ctx.table.SetRows(rows)
}

func (ctx *SeasonInfoCtx) validateInput(input string) error {
	if len([]rune(input)) > 0 {
		return nil
	}

	return errors.New("Input cannot be empty!")
}

func (ctx *SeasonInfoCtx) playSelected() error {
	if len(ctx.table.Rows()) < 1 {
		return nil
	}

	row := ctx.table.SelectedRow()
	if row == nil {
		return errors.New(`No row selected`)
	}

	epId, err := strconv.Atoi(row[0])

	if err != nil {
		return err
	}

	epIdx := slices.IndexFunc(*ctx.season.Episodes, func(ep models.Episode) bool { return ep.Id == epId })

	if epIdx < 0 {
		return fmt.Errorf(`Failed to find episode by id %d`, epId)
	}

	ep := (*ctx.season.Episodes)[epIdx]

	ep.IsPlayed = true

	err = ctx.repo.SaveEpisode(&ep)

	if err != nil {
		return err
	}

	err = ctx.repo.UpdateLastPlayedEpisodeId(ep.Id)

	if err != nil {
		return err
	}

	ctx.season.Episodes, err = ctx.repo.GetEpisodesBySeason(ctx.season.Id)

	if err != nil {
		return err
	}

	task := exec.Command("mpv", ep.Path)

	go func() {
		task.Run()
	}()

	return nil
}

func (ctx *SeasonInfoCtx) deleteSelected() error {
	if len(ctx.table.Rows()) < 1 {
		return nil
	}

	row := ctx.table.SelectedRow()

	epId, err := strconv.Atoi(row[0])

	if err != nil {
		return err
	}

	err = ctx.repo.DeleteEpisode(epId)

	if err != nil {
		return err
	}

	ctx.season.Episodes, err = ctx.repo.GetEpisodesBySeason(ctx.season.Id)

	if err != nil {
		return err
	}

	ctx.season.EpisodeCnt = len(*ctx.season.Episodes)
	ctx.table.SetCursor(ctx.table.Cursor() - 1)

	return nil
}

func (ctx *SeasonInfoCtx) toggleIsPlayedSelected() error {
	if len(ctx.table.Rows()) < 1 {
		return nil
	}

	row := ctx.table.SelectedRow()

	epId, err := strconv.Atoi(row[0])

	if err != nil {
		return err
	}

	ep, err := ctx.repo.GetEpisode(epId)

	if err != nil {
		return err
	}

	ep.IsPlayed = !ep.IsPlayed

	err = ctx.repo.SaveEpisode(ep)

	if err != nil {
		return err
	}

	ctx.season.Episodes, err = ctx.repo.GetEpisodesBySeason(ctx.season.Id)

	if err != nil {
		return err
	}

	ctx.season.EpisodeCnt = len(*ctx.season.Episodes)

	return err
}
