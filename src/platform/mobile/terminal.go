//go:build android || ios

package mobile

import (
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const IsMobile = true

var window fyne.Window

func InitTerminal() {
	readPipe, pipeErr := pipeStdStreams()

	var label *widget.Entry
	window, label = constructTerminal()

	if pipeErr != nil {
		label.SetText(pipeErr.Error())
		label.Refresh()
	} else {
		go redirectPipeToTerminal(readPipe, label)
	}
}

func RunTerminal() {
	window.ShowAndRun()
}

func pipeStdStreams() (*os.File, error) {
	readPipe, writePipe, pipeErr := os.Pipe()
	os.Stdout = writePipe
	os.Stderr = writePipe
	return readPipe, pipeErr
}

func constructTerminal() (fyne.Window, *widget.Entry) {
	a := app.New()
	a.Settings().SetTheme(&terminalTheme{})

	window = a.NewWindow("db1000n_Mobile")
	label := widget.NewMultiLineEntry()
	label.Disable()
	window.SetContent(label)

	return window, label
}

func redirectPipeToTerminal(readPipe *os.File, label *widget.Entry) {
	const bufferSize = 100
	buf := make([]byte, bufferSize, bufferSize)
	for {
		numRead, _ := readPipe.Read(buf)
		label.SetText(label.Text + string(buf[:numRead]))
		label.Refresh()
	}
}

type terminalTheme struct{}

func (terminalTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		switch name {
		case theme.ColorNameInputBackground:
			return color.White
		case theme.ColorNameDisabled:
			return color.Black
		}
	}

	switch name {
	case theme.ColorNameInputBackground:
		return color.Black
	case theme.ColorNameDisabled:
		return color.White
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (terminalTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (terminalTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (terminalTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
