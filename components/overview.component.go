package components

import (
	"fmt"
	"reflect"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/palianskas/toire/internal"
	repo "github.com/palianskas/toire/internal"
	"github.com/palianskas/toire/models"
)

type OverviewItem struct {
	name     string
	onSelect func() tea.Cmd
}

type OverviewCtx struct {
	ContextBase
	repo            *internal.Repo
	items           []OverviewItem
	selectedItemIdx int
}

func (ctx *OverviewCtx) Init() tea.Cmd {
	return nil
}

func (ctx *OverviewCtx) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	keyMsg, ok := msg.(tea.KeyMsg)

	if ok {
		keyCmd := ctx.handleKeyMsg(keyMsg)
		cmd = tea.Batch(cmd, keyCmd)

	}

	return ctx, cmd
}

func (ctx *OverviewCtx) View() string {
	ctx.paint(`Toire: cli media repertoire`)

	for idx, item := range ctx.items {
		prefix := ' '
		if ctx.selectedItemIdx == idx {
			prefix = '>'
		}

		ctx.paint(fmt.Sprintf("%c %s", prefix, item.name))

	}

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}

func NewOverview() OverviewCtx {
	repo, err := repo.GetInstance(DB_PATH)

	HandleFatalErr(err)

	ctx := OverviewCtx{
		ContextBase: NewContextBase(reflect.TypeOf(OverviewCtx{}).Name()),
		repo:        repo,
	}

	seriesListItem := OverviewItem{
		name:     "Open series list",
		onSelect: ctx.openSeriesList,
	}

	items := []OverviewItem{
		seriesListItem,
	}

	lastPlayedItem, err := ctx.getLastPlayedOverviewItem()

	if err != nil {
		ctx.err = err
	}

	if lastPlayedItem != nil {
		items = slices.Insert(items, 0, *lastPlayedItem)
	}

	ctx.items = items

	return ctx
}

func (ctx *OverviewCtx) WakeUp() error {
	// return ctx.setLastPlayedEntities()
	return nil
}

func (ctx *OverviewCtx) getLastPlayedOverviewItem() (*OverviewItem, error) {
	epId, err := ctx.repo.GetLastPlayedEpisodeId()

	if err != nil {
		return nil, nil // ignore
	}

	ep, err := ctx.repo.GetEpisode(epId)

	if err != nil {
		return nil, err
	}

	season, err := ctx.repo.GetSeason(ep.SeasonId)

	if err != nil {
		return nil, err
	}

	season.Episodes, err = ctx.repo.GetEpisodesBySeason(season.Id)

	if err != nil {
		return nil, err
	}

	series, err := ctx.repo.GetSeries(season.SeriesId)

	if err != nil {
		return nil, err
	}

	item := OverviewItem{
		name:     fmt.Sprintf("Open last played: %s S%02dE%02d", series.Name, season.Number, ep.Number),
		onSelect: func() tea.Cmd { return ctx.openLastPlayed(*season) },
	}

	return &item, nil
}

func (ctx *OverviewCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {

	cmd := ctx.handleBaseKeyMsg(msg)

	switch msg.String() {
	case "j":
		ctx.selectedItemIdx = (ctx.selectedItemIdx + 1) % len(ctx.items)
	case "k":
		ctx.selectedItemIdx = (ctx.selectedItemIdx - 1) % len(ctx.items)
	case "enter":
		cmd = tea.Batch(cmd, ctx.onItemSelected())
	}

	return cmd
}

func (ctx *OverviewCtx) openSeriesList() tea.Cmd {
	seriesList := NewSeriesList()
	return CtxPushCmd(&seriesList)
}

func (ctx *OverviewCtx) openLastPlayed(lastPlayedSeason models.Season) tea.Cmd {
	seasonInfo := NewSeasonInfo(&lastPlayedSeason)
	return CtxPushCmd(&seasonInfo)

}

func (ctx *OverviewCtx) onItemSelected() tea.Cmd {
	item := ctx.items[ctx.selectedItemIdx]

	return item.onSelect()
}
