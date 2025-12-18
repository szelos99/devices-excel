package main

import (
	"devices_excel/devices"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	app    fyne.App
	window fyne.Window

	fluke   *devices.Fluke
	additel *devices.Additel

	// state
	selectedDevice    Device
	selectedDirection Direction
	selectedStep      Step

	// excel
	excelConnected bool
	excelCloseCh   chan struct{}
	excelWriteCh   chan WritableData
	excelWg        sync.WaitGroup

	// logging
	logRich   *widget.RichText
	logScroll *container.Scroll
	log       func(LogLevel, string)
}

func NewUi() *UI {
	ui := &UI{
		app:            app.New(),
		fluke:          devices.NewFluke(),
		additel:        devices.NewAdditel(),
		excelCloseCh:   make(chan struct{}),
		excelWriteCh:   make(chan WritableData, 10),
		selectedDevice: DeviceNone,
		selectedStep:   StepNone,
	}

	ui.window = ui.app.NewWindow("Device Control")
	ui.initLogger()

	left := ui.buildControls()
	split := container.NewHSplit(left, ui.logScroll)
	split.SetOffset(0.35)

	ui.window.SetContent(split)
	ui.window.Resize(fyne.NewSize(700, 350))

	return ui
}

func (ui *UI) initLogger() {
	ui.logRich = widget.NewRichText()
	ui.logRich.Wrapping = fyne.TextWrapWord

	ui.logScroll = container.NewVScroll(ui.logRich)
	ui.logScroll.SetMinSize(fyne.NewSize(250, 200))

	ui.log = func(level LogLevel, msg string) {
		ts := time.Now().Format("15:04:05")
		style := widget.RichTextStyle{}
		prefix := "INFO"

		if level == ERROR {
			prefix = "ERROR"
			style.ColorName = "error"
			style.TextStyle.Bold = true
		}

		seg := &widget.TextSegment{
			Text:  fmt.Sprintf("%s %s: %s", ts, prefix, msg),
			Style: style,
		}

		fyne.Do(func() {
			if len(ui.logRich.Segments) > 100 {
				ui.logRich.Segments = append([]widget.RichTextSegment(nil), ui.logRich.Segments[1:]...)
			}
			ui.logRich.Segments = append(ui.logRich.Segments, seg)
			ui.logRich.Refresh()
			ui.logScroll.ScrollToBottom()
		})
	}
}

func (ui *UI) buildControls() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewButton("Połącz z Excelem", ui.connectExcel),
		container.NewGridWithColumns(2,
			widget.NewButton("Połącz z Fluke", ui.connectFluke),
			widget.NewButton("Połącz z Additelem", ui.connectAdditel),
		),
		ui.deviceButtons(),
		ui.directionButtons(),
		ui.stepButtons(),
		widget.NewButton("Zapisz", ui.save),
	)
}

func highlightSelected[T comparable](value T, buttons map[T]*widget.Button) {
	for k, b := range buttons {
		b.Importance = widget.MediumImportance
		if k == value {
			b.Importance = widget.HighImportance
		}
		b.Refresh()
	}
}

func (ui *UI) deviceButtons() fyne.CanvasObject {
	buttons := map[Device]*widget.Button{}

	buttons[Fluke] = widget.NewButton("Fluke", func() {
		ui.selectedDevice = Fluke
		highlightSelected(Fluke, buttons)
	})

	buttons[Additel] = widget.NewButton("Additel", func() {
		ui.selectedDevice = Additel
		highlightSelected(Additel, buttons)
	})

	return container.NewGridWithColumns(2,
		buttons[Fluke],
		buttons[Additel],
	)
}

func (ui *UI) directionButtons() fyne.CanvasObject {
	buttons := map[Direction]*widget.Button{}

	buttons[DirectionUp] = widget.NewButton("góra", func() {
		ui.selectedDirection = DirectionUp
		highlightSelected(DirectionUp, buttons)
	})

	buttons[DirectionDown] = widget.NewButton("dół", func() {
		ui.selectedDirection = DirectionDown
		highlightSelected(DirectionDown, buttons)
	})

	return container.NewGridWithColumns(2,
		buttons[DirectionUp],
		buttons[DirectionDown],
	)
}

func (ui *UI) stepButtons() fyne.CanvasObject {
	buttons := map[Step]*widget.Button{}

	buttons[Step1] = widget.NewButton("krok: 1", func() {
		ui.selectedStep = Step1
		highlightSelected(Step1, buttons)
	})

	buttons[Step2] = widget.NewButton("krok: 2", func() {
		ui.selectedStep = Step2
		highlightSelected(Step2, buttons)
	})

	return container.NewGridWithColumns(2,
		buttons[Step1],
		buttons[Step2],
	)
}

func (ui *UI) connectFluke() {
	go func() {
		ui.log(INFO, "Próba połączenia z Fluke...")
		if err := ui.fluke.Connect(); err != nil {
			ui.log(ERROR, err.Error())
			return
		}
		ui.log(INFO, "Połączono z Fluke")
	}()
}

func (ui *UI) connectAdditel() {
	go func() {
		ui.log(INFO, "Próba połączenia z Additelem...")
		if err := ui.additel.Connect(ui.log); err != nil {
			ui.log(ERROR, err.Error())
			return
		}
		ui.log(INFO, "Połączono z Additelem")
	}()
}

func (ui *UI) connectExcel() {
	if ui.excelConnected {
		return
	}

	ui.excelWg.Go(func() {
		RunExcelWriter(
			ui.log,
			ui.excelWriteCh,
			ui.excelCloseCh,
			&ui.excelConnected,
		)
	})
}

func (ui *UI) save() {
	if err := ui.validate(); err != nil {
		ui.log(ERROR, err.Error())
		return
	}

	go func() {
		var (
			value string
			err   error
		)

		switch ui.selectedDevice {
		case Fluke:
			value, err = ui.fluke.GetResult(ui.log)
		case Additel:
			value, err = ui.additel.GetResult(ui.log)
		}

		if err != nil {
			ui.log(ERROR, err.Error())
			return
		}

		ui.excelWriteCh <- WritableData{
			Value:     value,
			Direction: ui.selectedDirection,
			Step:      ui.selectedStep,
		}
	}()
}

func (ui *UI) validate() error {
	if !ui.excelConnected {
		return fmt.Errorf("nie połączono z Excelem")
	}
	if ui.selectedDevice == DeviceNone {
		return fmt.Errorf("nie wybrano urządzenia")
	}
	if ui.selectedDirection == DirectionNone {
		return fmt.Errorf("nie wybrano kierunku")
	}
	if ui.selectedStep == StepNone {
		return fmt.Errorf("nie wybrano kroku")
	}

	switch ui.selectedDevice {
	case Fluke:
		if ui.fluke.Port == nil {
			return fmt.Errorf("nie połączono z Fluke")
		}
	case Additel:
		if ui.additel.Port == nil {
			return fmt.Errorf("nie połączono z Additelem")
		}
	}
	return nil
}

func (ui *UI) CloseUI() {
	if ui.excelConnected {
		close(ui.excelCloseCh)
		ui.excelWg.Wait()
	}
	if ui.fluke.Port != nil {
		ui.fluke.Close()
	}
	if ui.additel.Port != nil {
		ui.additel.Close()
	}
}
