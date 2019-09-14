//go:generate goversioninfo

package main

import (
	"fmt"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/mastahyeti/cms"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

const programTitle = "p7s Extract"
const p7sExt = ".p7s"

var textColorRed walk.Color = walk.RGB(255, 0, 0)
var textColorDefault walk.Color

func main() {

	iconCache := initIconCache()
	defer iconCache.Treat()

	mw := new(MyMainWindow)

	_ = MainWindow{
		AssignTo:    &mw.MainWindow,
		Title:       programTitle,
		Icon:        iconCache.Get("main"),
		MinSize:     Size{Width: 400, Height: 60},
		Size:        Size{Width: 700, Height: 60},
		Persistent:  true,
		OnDropFiles: mw.onDropFiles,

		Layout: HBox{},
		Children: []Widget{
			PushButton{
				Text:        "Отваряне",
				ToolTipText: "Отваряне на .p7s файл",
				Image:       iconCache.Get("open"),
				OnClicked:   mw.openBtn_OnClicked,
			},
			LineEdit{
				AssignTo:      &mw.p7sFileNameLE,
				Text:          "",
				ToolTipText:   "Път до .p7s файл",
				OnTextChanged: mw.p7sFileLE_OnTextChanged,
			},
			PushButton{
				AssignTo:    &mw.extractBtn,
				Text:        "Извличане",
				ToolTipText: "Извличане на съдържащите се данни",
				Image:       iconCache.Get("extract"),
				Enabled:     false,
				OnClicked:   mw.extractBtn_OnClicked,
			},
		},
	}.Create()

	textColorDefault = mw.p7sFileNameLE.TextColor()

	mw.initHelpButton()

	// os.Args[0] == program name, os.Args[1] == first argument
	args := os.Args
	if len(args) > 1 {
		p7sFileName := args[1]
		mw.initP7SFileName(p7sFileName)
	}

	mw.Run()
}

func initIconCache() *NamedIconCache {

	iconCache := NewNamedIconCache()
	defer iconCache.Treat()

	mainIcon, err := walk.NewIconFromResourceId(2)
	if err != nil {
		mainIcon, _ = walk.NewIconFromSysDLL("SHELL32", 45)
	}
	iconCache.AddNamed("main", mainIcon)

	openIcon, _ := walk.NewIconFromSysDLL("SHELL32", 4)
	iconCache.AddNamed("open", openIcon)

	extractIcon, _ := walk.NewIconFromSysDLL("SHELL32", 132)
	iconCache.AddNamed("extract", extractIcon)

	helpIcon, _ := walk.NewIconFromSysDLL("SHELL32", 154)
	iconCache.AddNamed("help", helpIcon)

	return iconCache
}

type WndProcFunc func(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr

type MyMainWindow struct {
	*walk.MainWindow
	prevP7SFilePath string
	prevDataFileDir string
	p7sFileNameLE   *walk.LineEdit
	extractBtn      *walk.PushButton
	origWndProc     *WndProcFunc
	origWndProcPtr  uintptr
}

func (mw *MyMainWindow) initHelpButton() {

	hWnd := mw.Handle()
	var index, value int32

	newWndProcPtr := syscall.NewCallback(mw.WndProc)
	mw.origWndProcPtr = win.SetWindowLongPtr(hWnd, win.GWLP_WNDPROC, newWndProcPtr)

	index = win.GWL_STYLE
	value = win.GetWindowLong(hWnd, index)
	if newValue := value &^ (win.WS_MINIMIZEBOX | win.WS_MAXIMIZEBOX); newValue != value {
		win.SetLastError(0)
		win.SetWindowLong(hWnd, index, newValue)
	}

	index = win.GWL_EXSTYLE
	value = win.GetWindowLong(hWnd, index)
	if newValue := value | win.WS_EX_CONTEXTHELP; newValue != value {
		win.SetLastError(0)
		win.SetWindowLong(hWnd, index, newValue)
	}
}

func (mw *MyMainWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {

	switch msg {
	case win.WM_NCLBUTTONDOWN:
		if wParam == win.HTHELP {
			mw.showAbout()
			return 0
		}
	case win.WM_HELP:
		mw.showAbout()
	}

	return win.CallWindowProc(mw.origWndProcPtr, hwnd, msg, wParam, lParam)
}

func (mw *MyMainWindow) p7sFileLE_OnTextChanged() {

	p7sFileName := mw.p7sFileNameLE.Text()

	exists := fileExists(p7sFileName)
	if !exists {
		if p7sFileName != "" {
			_ = mw.p7sFileNameLE.SetToolTipText("Файлът не съществува")
			mw.p7sFileNameLE.SetTextColor(textColorRed)
		}
	} else {
		_ = mw.p7sFileNameLE.SetToolTipText("Път до .p7s файл")
		mw.p7sFileNameLE.SetTextColor(textColorDefault)
	}
	mw.extractBtn.SetEnabled(exists)
}

func (mw *MyMainWindow) onDropFiles(files []string) {

	if len(files) > 0 {
		p7sFileName := files[0]
		mw.initP7SFileName(p7sFileName)
	}
}

func (mw *MyMainWindow) initP7SFileName(p7sFileName string) {

	//if path.Ext(p7sFileName) != p7sExt {
	//	walk.MsgBox(nil, programTitle,
	//		"Грешен формат на файла.\r\nОчаква се файл с разширение .p7s",
	//		walk.MsgBoxOK | walk.MsgBoxIconWarning)
	//	return
	//}

	if filepath.Ext(p7sFileName) == p7sExt {
		mw.prevP7SFilePath = p7sFileName
	} else {
		baseDir := filepath.Dir(p7sFileName)
		mw.prevP7SFilePath = filepath.Join(baseDir, p7sExt)
	}

	_ = mw.p7sFileNameLE.SetText(p7sFileName)
}

func (mw *MyMainWindow) openBtn_OnClicked() {

	p7sFileName, err := mw.selectP7SFile()
	if err != nil {
		walk.MsgBox(mw, programTitle,
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	if p7sFileName != "" {
		mw.initP7SFileName(p7sFileName)
	}
}

func (mw *MyMainWindow) selectP7SFile() (p7sFileName string, err error) {

	openDialog := new(walk.FileDialog)
	openDialog.FilePath = mw.prevP7SFilePath
	openDialog.Filter = "p7s Файлове (*.p7s)|*.p7s"
	openDialog.Title = "Избор на .p7s файл"

	if ok, err := openDialog.ShowOpen(mw); err != nil {
		return "", fmt.Errorf("Системна грешка\r\n\r\n" + err.Error())
	} else if !ok {
		return "", nil
	}

	p7sFileName = openDialog.FilePath
	return p7sFileName, nil
}

func (mw *MyMainWindow) extractBtn_OnClicked() {

	p7sFileName := mw.p7sFileNameLE.Text()

	if p7sFileName == "" {
		walk.MsgBox(mw, programTitle,
			"Не е избран .p7s файл",
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	p7sContents, err := mw.readP7SContents(p7sFileName)
	if err != nil {
		walk.MsgBox(mw, programTitle,
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	data, err := mw.extractP7SData(p7sContents)
	if err != nil {
		walk.MsgBox(mw, programTitle,
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	dataFileName, err := mw.selectDataFile(p7sFileName)
	if err != nil {
		walk.MsgBox(mw, programTitle,
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	if dataFileName == "" {
		return
	}

	mw.prevDataFileDir = filepath.Dir(dataFileName)

	err = mw.writeData(dataFileName, data)
	if err != nil {
		walk.MsgBox(mw, programTitle,
			err.Error(),
			walk.MsgBoxOK|walk.MsgBoxIconError)
		return
	}

	walk.MsgBox(mw, programTitle,
		"Извличането завърши успешно",
		walk.MsgBoxOK|walk.MsgBoxIconExclamation)
}

func (mw *MyMainWindow) readP7SContents(p7sFileName string) (p7sContents []byte, err error) {

	p7sFileName = absolutePath(p7sFileName)

	p7sContents, err = ioutil.ReadFile(p7sFileName)
	if err != nil {
		return nil, fmt.Errorf("Грешка при четене на .p7s файл\r\n\r\n" + err.Error())
	}

	return p7sContents, nil
}

func (mw *MyMainWindow) extractP7SData(p7sFileContents []byte) (data []byte, err error) {

	signedData, err := cms.ParseSignedData(p7sFileContents)
	if err != nil {
		return nil, fmt.Errorf("Некоректен формат на .p7s файл\r\n\r\n" + err.Error())
	}

	if signedData.IsDetached() {
		return nil, fmt.Errorf("p7s файлът не съдържа данни")
	}

	data, err = signedData.GetData()
	if err != nil {
		return nil, fmt.Errorf("Грешка при извличане на данни от .p7s файл\r\n\r\n" + err.Error())
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("p7s файлът не съдържа данни")
	}

	//certificates, err := signedData.GetCertificates()
	//if err != nil {
	//	return err
	//}

	return data, nil
}

func (mw *MyMainWindow) selectDataFile(p7sFileName string) (dataFileName string, err error) {

	if filepath.Ext(p7sFileName) == p7sExt {
		dataFileName = p7sFileName[0 : len(p7sFileName)-len(p7sExt)]
	}

	if mw.prevDataFileDir != "" {
		_, fileName := filepath.Split(dataFileName)
		dataFileName = filepath.Join(mw.prevDataFileDir, fileName)
	}

	saveDialog := new(walk.FileDialog)
	saveDialog.FilePath = dataFileName
	saveDialog.Filter = "Всички Файлове (*.*)|*.*"
	saveDialog.Title = "Запис на файл с данни"
	saveDialog.Flags = win.OFN_OVERWRITEPROMPT

	if ok, err := saveDialog.ShowSave(mw); err != nil {
		return "", fmt.Errorf("Системна грешка\r\n\r\n" + err.Error())
	} else if !ok {
		return "", nil
	}

	dataFileName = saveDialog.FilePath
	return dataFileName, nil
}

func (mw *MyMainWindow) writeData(dataFileName string, data []byte) error {

	dataFileName = absolutePath(dataFileName)

	err := ioutil.WriteFile(dataFileName, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Грешка при запис на извлечените данни\r\n\r\n" + err.Error())
	}

	return nil
}

func (mw *MyMainWindow) showAbout() {
	walk.MsgBox(mw, programTitle,
		programTitle+" v1.0.3\r\n\r\n© 2019 Д-р Дамян Митев\r\n\r\ndamyan_mitev@mail.bg",
		walk.MsgBoxOK|walk.MsgBoxIconQuestion)
}

func absolutePath(fileName string) string {

	if filepath.IsAbs(fileName) {
		return fileName
	}

	currentDir, _ := os.Getwd()
	return filepath.Join(currentDir, fileName)
}

func fileExists(fileName string) bool {

	fileName = absolutePath(fileName)
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
