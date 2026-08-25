package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	app              *tview.Application
	reader           *application.Reader
	status, view     *tview.TextView
	root             *tview.Flex
	config           *configstore.Store
	settings         configstore.Config
	bookID, bookPath string
	book             domain.Book
	focusables       []tview.Primitive
	mainFocusables   []tview.Primitive
	mainButtons      []*tview.Button
	themeIndex       int
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
	u := &UI{app: tview.NewApplication(), reader: reader, config: store, settings: settings, status: tview.NewTextView().SetDynamicColors(true), view: tview.NewTextView().SetScrollable(true)}
	u.build()
	if settings.CurrentFile != "" {
		if book, openErr := reader.OpenFile(settings.CurrentFile); openErr == nil {
			u.showBook(settings.CurrentFile, book)
		}
	}
	if err := u.app.SetRoot(u.root, true).EnableMouse(true).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (u *UI) build() {
	u.applyTheme()
	u.view.SetText("Откройте FB2-файл клавишей o или скачайте книгу клавишей d.").SetBorder(true).SetTitle(" Чтение ")
	u.status.SetText(" o открыть   d скачать   b книги   q выйти ").SetBorder(true)
	openButton := tview.NewButton("Открыть (o)").SetSelectedFunc(u.openDialog)
	downloadButton := tview.NewButton("Скачать (d)").SetSelectedFunc(u.downloadDialog)
	booksButton := tview.NewButton("Открытые книги (b)").SetSelectedFunc(u.booksDialog)
	themeButton := tview.NewButton("Темы").SetSelectedFunc(u.nextTheme)
	closeButton := tview.NewButton("Закрыть").SetSelectedFunc(func() { u.saveCurrent(); u.app.Stop() })
	u.mainButtons = []*tview.Button{openButton, downloadButton, booksButton, themeButton, closeButton}
	buttons := tview.NewFlex().AddItem(openButton, 0, 1, false).AddItem(downloadButton, 0, 1, false).AddItem(booksButton, 0, 1, false).AddItem(themeButton, 0, 1, false).AddItem(closeButton, 0, 1, false).AddItem(u.status, 0, 2, false)
	u.root = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(u.view, 0, 1, true).AddItem(buttons, 3, 0, false)
	u.mainFocusables = []tview.Primitive{u.view, openButton, downloadButton, booksButton, themeButton, closeButton}
	u.focusables = u.mainFocusables
	u.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		u.saveCurrent()
		if event.Key() == tcell.KeyF10 {
			u.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTAB || event.Key() == tcell.KeyBacktab {
			u.cycleFocus(event.Key() == tcell.KeyTAB)
			return nil
		}
		switch event.Rune() {
		case 'o':
			u.openDialog()
			return nil
		case 'd':
			u.downloadDialog()
			return nil
		case 'b':
			u.booksDialog()
			return nil
		case 'q':
			u.saveCurrent()
			u.app.Stop()
			return nil
		}
		return event
	})
}

func (u *UI) nextTheme() {
	u.themeIndex = (u.themeIndex + 1) % 3
	u.applyTheme()
}

func (u *UI) applyTheme() {
	themes := []tview.Theme{
		{PrimitiveBackgroundColor: tcell.ColorBlack, ContrastBackgroundColor: tcell.ColorDarkCyan, MoreContrastBackgroundColor: tcell.ColorTeal, BorderColor: tcell.ColorGray, TitleColor: tcell.ColorYellow, GraphicsColor: tcell.ColorWhite, PrimaryTextColor: tcell.ColorWhite, SecondaryTextColor: tcell.ColorSilver, TertiaryTextColor: tcell.ColorGray, InverseTextColor: tcell.ColorBlack, ContrastSecondaryTextColor: tcell.ColorWhite},
		{PrimitiveBackgroundColor: tcell.ColorNavy, ContrastBackgroundColor: tcell.ColorDarkBlue, MoreContrastBackgroundColor: tcell.ColorBlue, BorderColor: tcell.ColorLightSteelBlue, TitleColor: tcell.ColorLightCyan, GraphicsColor: tcell.ColorWhite, PrimaryTextColor: tcell.ColorWhite, SecondaryTextColor: tcell.ColorLightCyan, TertiaryTextColor: tcell.ColorLightSteelBlue, InverseTextColor: tcell.ColorNavy, ContrastSecondaryTextColor: tcell.ColorWhite},
		{PrimitiveBackgroundColor: tcell.ColorWhite, ContrastBackgroundColor: tcell.ColorLightYellow, MoreContrastBackgroundColor: tcell.ColorYellow, BorderColor: tcell.ColorDarkGray, TitleColor: tcell.ColorDarkBlue, GraphicsColor: tcell.ColorBlack, PrimaryTextColor: tcell.ColorBlack, SecondaryTextColor: tcell.ColorDarkGray, TertiaryTextColor: tcell.ColorGray, InverseTextColor: tcell.ColorWhite, ContrastSecondaryTextColor: tcell.ColorBlack},
	}
	tview.Styles = themes[u.themeIndex]
	if u.view != nil {
		u.view.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor).SetBorderColor(tview.Styles.BorderColor)
		u.status.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor).SetBorderColor(tview.Styles.BorderColor)
	}
	for _, button := range u.mainButtons {
		if button != nil {
			button.SetStyle(tcell.StyleDefault.Foreground(tview.Styles.PrimaryTextColor).Background(tview.Styles.PrimitiveBackgroundColor)).SetActivatedStyle(tcell.StyleDefault.Foreground(tview.Styles.InverseTextColor).Background(tview.Styles.MoreContrastBackgroundColor))
		}
	}
}

func (u *UI) cycleFocus(forward bool) {
	if len(u.focusables) == 0 {
		return
	}
	current := u.app.GetFocus()
	index := 0
	for i, primitive := range u.focusables {
		if primitive == current {
			index = i
			break
		}
	}
	if forward {
		index = (index + 1) % len(u.focusables)
	} else {
		index = (index - 1 + len(u.focusables)) % len(u.focusables)
	}
	u.app.SetFocus(u.focusables[index])
}

func (u *UI) openDialog() {
	initialDir := u.settings.LastDir
	if initialDir == "" {
		initialDir, _ = os.UserHomeDir()
	}
	initialDir, _ = filepath.Abs(initialDir)
	folders, files := tview.NewList(), tview.NewList()
	for _, list := range []*tview.List{folders, files} {
		list.ShowSecondaryText(false)
		list.SetBorder(true)
	}
	folders.SetTitle(" Папки ")
	files.SetTitle(" FB2 файлы ")
	var populate func(string)
	populate = func(path string) { u.populateFolders(folders, path, populate); u.populateFiles(files, path) }
	populate(initialDir)
	close := func() { u.focusables = u.mainFocusables; u.app.SetRoot(u.root, true).SetFocus(u.view) }
	panels := tview.NewFlex().AddItem(folders, 0, 1, true).AddItem(files, 0, 2, false)
	closeButton := tview.NewButton("Закрыть").SetSelectedFunc(close)
	u.focusables = []tview.Primitive{folders, files, closeButton}
	u.app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).AddItem(panels, 0, 1, true).AddItem(closeButton, 2, 0, false), true).SetFocus(folders)
}

func (u *UI) populateFolders(list *tview.List, path string, enter func(string)) {
	list.Clear()
	list.AddItem("..", "", 0, func() { enter(filepath.Dir(path)) })
	entries, err := os.ReadDir(path)
	if err != nil {
		list.AddItem("Ошибка: "+err.Error(), "", 0, nil)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			next := filepath.Join(path, entry.Name())
			list.AddItem(entry.Name(), "", 0, func() { enter(next) })
		}
	}
}
func (u *UI) populateFiles(list *tview.List, path string) {
	list.Clear()
	entries, err := os.ReadDir(path)
	if err != nil {
		list.AddItem("Ошибка: "+err.Error(), "", 0, nil)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "" && filepath.Ext(entry.Name()) == ".fb2" {
			file := filepath.Join(path, entry.Name())
			list.AddItem(entry.Name(), "", 0, func() { u.load(file, func() (domain.Book, error) { return u.reader.OpenFile(file) }) })
		}
	}
}

func (u *UI) downloadDialog() {
	url := tview.NewInputField().SetLabel("URL: ")
	close := func() { u.focusables = u.mainFocusables; u.app.SetRoot(u.root, true).SetFocus(u.view) }
	form := tview.NewForm().AddButton("Скачать и открыть", func() {
		value := url.GetText()
		u.load(value, func() (domain.Book, error) {
			book, _, err := u.reader.Download(context.Background(), value)
			return book, err
		})
	}).AddButton("Закрыть", close)
	u.focusables = []tview.Primitive{url, form}
	u.app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).AddItem(url, 2, 0, true).AddItem(form, 3, 0, false), true).SetFocus(url)
}
func (u *UI) booksDialog() {
	list := tview.NewList()
	list.ShowSecondaryText(true)
	list.SetBorder(true).SetTitle(" Открытые книги ")
	books, err := u.config.ListBooks()
	if err != nil {
		list.AddItem("Ошибка: "+err.Error(), "", 0, nil)
	}
	sort.Slice(books, func(i, j int) bool { return books[i].Title < books[j].Title })
	for _, state := range books {
		current := state
		label := current.Title
		if label == "" {
			label = current.FileName
		}
		list.AddItem(label, current.Path, 0, func() { u.load(current.Path, func() (domain.Book, error) { return u.reader.OpenFile(current.Path) }) })
	}
	closeButton := tview.NewButton("Закрыть").SetSelectedFunc(func() { u.focusables = u.mainFocusables; u.app.SetRoot(u.root, true).SetFocus(u.view) })
	u.focusables = []tview.Primitive{list, closeButton}
	u.app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).AddItem(list, 0, 1, true).AddItem(closeButton, 2, 0, false), true).SetFocus(list)
}

func (u *UI) load(path string, read func() (domain.Book, error)) {
	u.status.SetText(" Загрузка... ")
	go func() {
		book, err := read()
		u.app.QueueUpdateDraw(func() {
			if err != nil {
				u.status.SetText(" Ошибка: " + err.Error() + " ")
				return
			}
			u.showBook(path, book)
		})
	}()
}
func (u *UI) showBook(path string, book domain.Book) {
	id, state, err := u.config.StateFor(path)
	if err != nil {
		u.status.SetText(" Ошибка: " + err.Error() + " ")
		return
	}
	u.bookID, u.bookPath, u.book = id, path, book
	u.settings.CurrentFile, u.settings.CurrentUUID, u.settings.LastDir = path, id, filepath.Dir(path)
	if err := u.config.SaveConfig(u.settings); err != nil {
		u.status.SetText(" Ошибка: " + err.Error() + " ")
		return
	}
	u.view.SetText(book.Text)
	u.view.ScrollTo(state.Line, 0)
	u.focusables = u.mainFocusables
	u.app.SetRoot(u.root, true).SetFocus(u.view)
	u.status.SetText(" o открыть   d скачать   b книги   q выйти ")
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
