package components

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type FilePickedMsg struct {
	Path string
}

func (msg FilePickedMsg) String() string {
	return `FilePickedMsg ` + msg.Path
}

func FilePickedCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return FilePickedMsg{
			Path: path,
		}
	}
}

const FILE_PICKER_ITEM_CNT = 20

type FilePickerCtx struct {
	Path           string
	IsDirSelection bool
	focused        bool
	items          []os.DirEntry
	paginator      paginator.Model
	table          table.Model
	err            error
	ContextBase
}

func NewFilePicker(path *string, isDirSelection bool) FilePickerCtx {
	if path == nil {
		homePath, err := os.UserHomeDir()

		if err != nil {
			HandleFatalErr(err)
		}

		path = &homePath
	}

	ctx := FilePickerCtx{
		Path:           *path,
		IsDirSelection: isDirSelection,
		ContextBase:    NewContextBase(reflect.TypeOf(FilePickerCtx{}).Name()),
	}

	ctx.Init()

	return ctx
}

func (ctx *FilePickerCtx) Init() tea.Cmd {

	columns := []table.Column{
		{
			Title: `Type`, Width: 4,
		},
		{
			Title: "Name", Width: 50,
		},
	}
	ctx.table = table.New(table.WithColumns(columns), table.WithFocused(ctx.focused))
	ctx.table.SetHeight(FILE_PICKER_ITEM_CNT)
	ctx.table.SetWidth(WIN_WIDTH)

	ctx.paginator = paginator.New()
	ctx.paginator.Type = paginator.Arabic
	ctx.paginator.PerPage = ctx.table.Height()

	return nil
}

func (ctx *FilePickerCtx) Update(msg tea.Msg) (*FilePickerCtx, tea.Cmd) {

	keyMsg, ok := msg.(tea.KeyMsg)

	var cmd tea.Cmd

	if ok {
		cmd = ctx.handleBaseKeyMsg(keyMsg)

		if cmd != nil {
			return ctx, cmd
		}

		cmd = ctx.handleKeyMsg(keyMsg)
	}

	ctx.items, ctx.err = os.ReadDir(ctx.Path)

	if ctx.err != nil {
		return ctx, nil
	}

	var tcmd tea.Cmd
	var pcmd tea.Cmd
	ctx.table, tcmd = ctx.table.Update(msg)
	ctx.paginator, pcmd = ctx.paginator.Update(msg)

	cmd = tea.Batch(cmd, tcmd, pcmd)

	ctx.paginator.SetTotalPages(len(ctx.items))
	ctx.setRenderedItems()

	return ctx, cmd

}

func (ctx *FilePickerCtx) View() string {
	ctx.paint(`Current path:` + ctx.Path)

	if !ctx.focused {
		return ctx.render()
	}
	if len(ctx.items) > 0 {
		ctx.paint(ctx.table.View())
		ctx.paint(ctx.paginator.View())
	} else {
		ctx.paint(`No items...`)
	}

	ctx.paint("Open dir/Select file: <Enter>; Select: <Space>; Up: <k>; Down: <j>; Next page: <l>; Prev. page: <h>; Back: <->, Quit: <q>")

	ctx.paintError()
	ctx.paintDebug()

	return ctx.render()
}

func (ctx *FilePickerCtx) Focused() bool {
	return ctx.focused
}

func (ctx *FilePickerCtx) Focus() {
	ctx.focused = true
	ctx.table.Focus()
}

func (ctx *FilePickerCtx) Blur() {
	ctx.focused = false
	ctx.table.Blur()
}

func (ctx *FilePickerCtx) getTableRows(items []os.DirEntry) ([]table.Row, error) {
	var rows []table.Row

	for _, item := range items {
		row, err := ctx.toTableRow(item)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	return rows, nil

}

func (ctx *FilePickerCtx) toTableRow(item os.DirEntry) (table.Row, error) {
	_, err := item.Info()

	if err != nil {
		return nil, err
	}

	row := table.Row{
		item.Type().String()[:1],
		item.Name(),
	}

	return row, nil
}

func (ctx *FilePickerCtx) setRenderedItems() {
	ctx.paginator.SetTotalPages(len(ctx.items))
	ctx.paginator.Page = min(ctx.paginator.Page, ctx.paginator.TotalPages-1)

	if len(ctx.items) < 1 {
		return
	}

	l, r := ctx.paginator.GetSliceBounds(len(ctx.items))
	rows, err := ctx.getTableRows(ctx.items[l:r])

	if err != nil {
		ctx.err = err
		return
	}

	ctx.table.SetRows(rows)
}

func (ctx *FilePickerCtx) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		return ctx.onItemSelect(false)
	case " ":
		return ctx.onItemSelect(ctx.IsDirSelection)
	case "-":
		ctx.onPopDir()
	}

	return nil
}

func (ctx *FilePickerCtx) onItemSelect(isDirSelectable bool) tea.Cmd {
	if len(ctx.items) < 1 {
		return nil
	}

	name := ctx.table.SelectedRow()[1]

	var item os.DirEntry = nil

	for i := range ctx.items {
		if ctx.items[i].Name() != name {
			continue
		}
		item = ctx.items[i]
	}

	if item == nil {
		return nil
	}

	fullPath := ctx.Path + "/" + name

	var linkPath string
	linkPath, ctx.err = os.Readlink(fullPath)

	if ctx.err == nil {
		fullPath = linkPath
	}

	fullPath, ctx.err = filepath.Abs(fullPath)
	if ctx.err != nil {
		return nil
	}
	_, ctx.err = os.ReadDir(fullPath)

	isDir := ctx.err == nil

	if !isDir || isDirSelectable {
		return FilePickedCmd(fullPath)
	}

	ctx.err = os.Chdir(fullPath)

	if ctx.err == nil {
		ctx.Path = fullPath
		ctx.table.SetCursor(0)
	}

	return nil

}

func (ctx *FilePickerCtx) onPopDir() {
	idx := 0
	for i, c := range ctx.Path {
		if c == '/' {
			idx = i
		}
	}

	idx = max(1, idx)
	ctx.Path = ctx.Path[:idx]

}
