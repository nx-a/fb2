package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	configstore "github.com/nx-a/fb2/internal/adapters/config"
	"github.com/nx-a/fb2/internal/adapters/fb2"
	"github.com/nx-a/fb2/internal/adapters/filesystem"
	"github.com/nx-a/fb2/internal/adapters/httpclient"
	"github.com/nx-a/fb2/internal/application"
	"github.com/nx-a/fb2/internal/domain"
	"github.com/rivo/tview"
)

type UI struct {
	app           *tview.Application
	reader        *application.Reader
	title, status *tview.TextView
	view          *tview.TextView
	root          *tview.Flex
	config        *configstore.Store
	settings      configstore.Config
	bookID        string
	bookPath      string
	book          domain.Book
}

func main() {
	reader := application.NewReader(filesystem.Store{}, httpclient.New(), fb2.Parser{})
	store, err := configstore.NewStore()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	settings, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	u := &UI{app: tview.NewApplication(), reader: reader, config: store, settings: settings, title: tview.NewTextView().SetDynamicColors(true), status: tview.NewTextView().SetDynamicColors(true), view: tview.NewTextView().SetScrollable(true)}
	u.build()
	if settings.CurrentFile != "" {
		if book, openErr := reader.OpenFile(settings.CurrentFile); openErr == nil {
			u.showBook(settings.CurrentFile, book)
		}
	}
	if err := u.app.SetRoot(u.root, true).EnableMouse(false).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (u *UI) build() {
	u.title.SetText("[yellow::b] FB2 READER").SetBorder(true).SetBorderColor(tcell.ColorDarkCyan)
	u.view.SetText("Откройте FB2-файл клавишей [yellow]o[-] или скачайте книгу клавишей [yellow]d[-].").SetBorder(true).SetTitle(" Чтение ")
	u.status.SetText("[gray] o открыть   d скачать   q выйти   ↑/↓ прокрутка[-]").SetBorder(true)
	u.root = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(u.title, 3, 0, false).AddItem(u.view, 0, 1, true).AddItem(u.status, 1, 0, false)
	u.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		u.saveCurrent()
		switch event.Rune() {
		case 'o':
			u.openDialog()
			return nil
		case 'd':
			u.downloadDialog()
			return nil
		case 'q':
			u.saveCurrent()
			u.app.Stop()
			return nil
		}
		return event
	})
}

func (u *UI) openDialog() {
	initialDir := u.settings.LastDir
	if initialDir == "" {
		initialDir = "."
	}
	input := tview.NewInputField().SetLabel("Каталог или файл: ").SetText(initialDir)
	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" FB2 файлы ")
	input.SetChangedFunc(func(text string) { u.populate(list, text) })
	u.populate(list, initialDir)
	form := tview.NewForm().AddButton("Открыть", func() {
		path, _ := list.GetItemText(list.GetCurrentItem())
		if path != "" {
			u.load(path, func() (domain.Book, error) { return u.reader.OpenFile(path) })
		}
	}).AddButton("Закрыть", func() { u.app.SetRoot(u.root, true).SetFocus(u.view) })
	box := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(input, 2, 0, true).AddItem(list, 0, 1, false).AddItem(form, 3, 0, false)
	u.app.SetRoot(box, true).SetFocus(input)
}

func (u *UI) populate(list *tview.List, path string) {
	files, err := u.reader.Search(path)
	list.Clear()
	if err != nil {
		list.AddItem("Ошибка: "+err.Error(), "", 0, nil)
		return
	}
	for _, file := range files {
		f := file
		list.AddItem(f, "", 0, func() { u.load(f, func() (domain.Book, error) { return u.reader.OpenFile(f) }) })
	}
}
func (u *UI) downloadDialog() {
	url := tview.NewInputField().SetLabel("URL: ")
	form := tview.NewForm().AddButton("Скачать и открыть", func() {
		value := url.GetText()
		u.load(value, func() (domain.Book, error) {
			book, _, err := u.reader.Download(context.Background(), value)
			return book, err
		})
	}).AddButton("Закрыть", func() { u.app.SetRoot(u.root, true).SetFocus(u.view) })
	u.app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).AddItem(url, 2, 0, true).AddItem(form, 3, 0, false), true).SetFocus(url)
}
func (u *UI) load(path string, read func() (domain.Book, error)) {
	u.status.SetText("[yellow]Загрузка...[-]")
	go func() {
		book, err := read()
		u.app.QueueUpdateDraw(func() {
			if err != nil {
				u.status.SetText("[red]" + err.Error() + "[-]")
				return
			}
			u.showBook(path, book)
		})
	}()
}
func (u *UI) showBook(path string, book domain.Book) {
	id, state, err := u.config.StateFor(path)
	if err != nil {
		u.status.SetText("[red]" + err.Error() + "[-]")
		return
	}
	u.bookID, u.bookPath, u.book = id, path, book
	u.settings.CurrentFile, u.settings.CurrentUUID, u.settings.LastDir = path, id, filepath.Dir(path)
	if err := u.config.SaveConfig(u.settings); err != nil {
		u.status.SetText("[red]" + err.Error() + "[-]")
		return
	}
	u.title.SetText("[yellow::b] " + book.Title)
	u.view.SetText(book.Text)
	u.view.ScrollTo(state.Line, 0)
	u.app.SetRoot(u.root, true).SetFocus(u.view)
	u.status.SetText("[gray] o открыть   d скачать   q выйти   ↑/↓ прокрутка[-]")
}

func (u *UI) saveCurrent() {
	if u.bookID == "" {
		return
	}
	line, _ := u.view.GetScrollOffset()
	_ = u.config.SaveBook(u.bookID, u.book, u.bookPath, line)
	u.settings.CurrentFile, u.settings.CurrentUUID, u.settings.LastDir = u.bookPath, u.bookID, filepath.Dir(u.bookPath)
	_ = u.config.SaveConfig(u.settings)
}
