package main

// import (
// 	"devices_excel/devices"
// 	"fmt"
// 	//"runtime"
// 	"sync"
// 	"time"

// 	"fyne.io/fyne/v2"
// 	"fyne.io/fyne/v2/app"
// 	"fyne.io/fyne/v2/container"
// 	"fyne.io/fyne/v2/widget"
// )

// // Enums for device and direction
// type Device int
// type Direction int
// type Jump int

// const (
// 	DeviceNone Device = iota
// 	Fluke
// 	Additel
// )

// const (
// 	JumpNone Jump = iota
// 	Jump1
// 	Jump2
// )

// type UI struct {
// 	app     fyne.App
// 	window  fyne.Window
// 	fluke   *devices.Fluke
// 	additel *devices.Additel

// 	excelConnected bool
// 	excelCloseCh   chan struct{}
// 	excelWriteCh   chan WritableData
// 	excelWg        sync.WaitGroup

// 	logRich   *widget.RichText // <--- added log panel
// 	logScroll *container.Scroll
// }

// func NewUi() *UI {
// 	a := app.New()
// 	w := a.NewWindow("Device Control")

// 	ui := &UI{
// 		app:            a,
// 		window:         w,
// 		fluke:          devices.NewFluke(),
// 		additel:        devices.NewAdditel(),
// 		excelConnected: false,
// 		excelCloseCh:   make(chan struct{}),
// 		excelWriteCh:   make(chan WritableData, 10),
// 	}

// 	// ---------------------------------------------------------
// 	// LOG PANEL (right side)
// 	// ---------------------------------------------------------
// 	ui.logRich = widget.NewRichText()
// 	ui.logRich.Wrapping = fyne.TextWrapWord

// 	ui.logScroll = container.NewVScroll(ui.logRich)
// 	ui.logScroll.SetMinSize(fyne.NewSize(250, 200))

// 	log := func(level LogLevel, msg string) {
// 		timestamp := time.Now().Format("15:04:05")

// 		var style widget.RichTextStyle
// 		var prefix string

// 		switch level {
// 		case INFO:
// 			prefix = "INFO "
// 			style = widget.RichTextStyle{
// 				TextStyle: fyne.TextStyle{
// 					Bold: true,
// 				},
// 			}
// 		case ERROR:
// 			prefix = "ERROR"
// 			style = widget.RichTextStyle{
// 				ColorName: "error",
// 				TextStyle: fyne.TextStyle{
// 					Bold: true,
// 				},
// 			}
// 		}

// 		seg := &widget.TextSegment{
// 			Text:  fmt.Sprintf("%s %s: %s", timestamp, prefix, msg),
// 			Style: style,
// 		}

// 		fyne.Do(func() {
// 			if len(ui.logRich.Segments) > 100 {
// 				ui.logRich.Segments = ui.logRich.Segments[1:]
// 			}
// 			ui.logRich.Segments = append(ui.logRich.Segments, seg)
// 			ui.logRich.Refresh()
// 			ui.logScroll.Refresh()
// 			ui.logScroll.ScrollToBottom()
// 		})
// 	}

// 	// ---------------------------------------------------------
// 	// LEFT-SIDE CONTROLS
// 	// ---------------------------------------------------------
// 	var selectedDevice Device = DeviceNone
// 	var selectedDirection Direction = DirectionNone
// 	var selectedJump Jump = JumpNone

// 	flukeButton := widget.NewButton("Fluke", nil)
// 	additelButton := widget.NewButton("Additel", nil)
// 	buttonUp := widget.NewButton("góra", nil)
// 	buttonDown := widget.NewButton("dół", nil)
// 	jump1Button := widget.NewButton("krok: 1", nil)
// 	jump2Button := widget.NewButton("krok: 2", nil)

// 	updateJumpButtons := func(selected Jump) {
// 		selectedJump = selected
// 		jump1Button.Importance = widget.MediumImportance
// 		jump2Button.Importance = widget.MediumImportance

// 		switch selected {
// 		case Jump1:
// 			jump1Button.Importance = widget.HighImportance
// 		case Jump2:
// 			jump2Button.Importance = widget.HighImportance
// 		}
// 		jump1Button.Refresh()
// 		jump2Button.Refresh()
// 	}

// 	updateDeviceButtons := func(selected Device) {
// 		selectedDevice = selected
// 		flukeButton.Importance = widget.MediumImportance
// 		additelButton.Importance = widget.MediumImportance

// 		switch selected {
// 		case Fluke:
// 			flukeButton.Importance = widget.HighImportance
// 		case Additel:
// 			additelButton.Importance = widget.HighImportance
// 		}
// 		flukeButton.Refresh()
// 		additelButton.Refresh()
// 	}

// 	updateDirectionButtons := func(selected Direction) {
// 		selectedDirection = selected
// 		buttonUp.Importance = widget.MediumImportance
// 		buttonDown.Importance = widget.MediumImportance

// 		switch selected {
// 		case DirectionUp:
// 			buttonUp.Importance = widget.HighImportance
// 		case DirectionDown:
// 			buttonDown.Importance = widget.HighImportance
// 		}
// 		buttonUp.Refresh()
// 		buttonDown.Refresh()
// 	}

// 	flukeButton.OnTapped = func() {
// 		updateDeviceButtons(Fluke)
// 	}

// 	additelButton.OnTapped = func() {
// 		updateDeviceButtons(Additel)
// 	}

// 	buttonUp.OnTapped = func() {
// 		updateDirectionButtons(DirectionUp)
// 	}

// 	buttonDown.OnTapped = func() {
// 		updateDirectionButtons(DirectionDown)
// 	}

// 	jump1Button.OnTapped = func() {
// 		updateJumpButtons(Jump1)
// 	}
// 	jump2Button.OnTapped = func() {
// 		updateJumpButtons(Jump2)
// 	}

// 	connectToFlukeButton := widget.NewButton("Połącz z Fluke", func() {
// 		go func() {
// 			fyne.Do(func() {
// 				log(INFO, "Próba połączenia z Fluke...")
// 			})

// 			err := ui.fluke.Connect()
// 			if err != nil {
// 				fyne.Do(func() {
// 					log(ERROR, "Błąd połączenia z Fluke: "+err.Error())
// 				})
// 			} else {
// 				fyne.Do(func() {
// 					log(INFO, "Połączono z Fluke")
// 				})
// 			}
// 		}()
// 	})

// 	connectAdditelButton := widget.NewButton("Połącz z Additelem", func() {
// 		go func() {
// 			fyne.Do(func() {
// 				log(INFO, "Próba połączenia z Additelem...")
// 			})

// 			err := ui.additel.Connect(log)
// 			if err != nil {
// 				fyne.Do(func() {
// 					log(ERROR, "Błąd połączenia z Additelem: "+err.Error())
// 				})
// 			} else {
// 				fyne.Do(func() {
// 					log(INFO, "Połączono z Additelem")
// 				})
// 			}
// 		}()
// 	})

// 	connectToExcelv2 := widget.NewButton("Połącz z Excelem v2", func() {
// 		if ui.excelConnected {
// 			return
// 		}
// 		ui.excelWg.Go(func() {
// 			RunExcelWriter(log, ui.excelWriteCh, ui.excelCloseCh, &ui.excelConnected)
// 		})
// 		// ui.excelWg.Add(1)
// 		// go func() {
// 		// 	runtime.LockOSThread()
// 		// 	defer runtime.UnlockOSThread()
// 		// 	defer ui.excelWg.Done()

// 		// 	e, err := ConnectToExcel()
// 		// 	if err != nil {
// 		// 		fyne.Do(func() { log(ERROR, "Błąd połączenia z Excelem: "+err.Error()) })
// 		// 		return
// 		// 	}
// 		// 	fyne.Do(func() { log(INFO, "Połączono z Excelem") })
// 		// 	ui.excelConnected = true

// 		// 	for {
// 		// 		select {
// 		// 		case writeVar := <-ui.excelWriteCh:
// 		// 			err := e.WriteAndMove(writeVar.Value, writeVar.Direction, writeVar.Jump)
// 		// 			if err != nil {
// 		// 				fyne.Do(func() { log(ERROR, "Błąd zapisu do Excela: "+err.Error()) })
// 		// 			}
// 		// 			fyne.Do(func() { log(INFO, "Zapisano do excela") })
// 		// 		case <-ui.excelCloseCh:
// 		// 			e.Close()
// 		// 			return
// 		// 		}
// 		// 	}
// 		// }()
// 	})
// 	// SAVE BUTTON
// 	save := widget.NewButton("zapisz", func() {
// 		ui.save(log, selectedDevice, selectedDirection, selectedJump)
// 	})

// 	leftSide := container.NewVBox(
// 		connectToExcelv2,
// 		container.NewGridWithColumns(2, connectToFlukeButton, connectAdditelButton),
// 		container.NewGridWithColumns(2, flukeButton, additelButton),
// 		container.NewGridWithColumns(2, buttonUp, buttonDown),
// 		container.NewGridWithColumns(2, jump1Button, jump2Button),
// 		save,
// 	)

// 	// ---------------------------------------------------------
// 	// SPLIT VIEW: LEFT BUTTONS | RIGHT LOG PANEL
// 	// ---------------------------------------------------------
// 	split := container.NewHSplit(leftSide, ui.logScroll)
// 	split.SetOffset(0.35) // 35% left panel, 65% right panel

// 	w.SetContent(split)
// 	w.Resize(fyne.NewSize(700, 350))

// 	return ui
// }

// func (ui *UI) CloseUI() {
// 	if ui.excelConnected {
// 		close(ui.excelCloseCh)
// 		ui.excelWg.Wait()
// 	}
// 	if ui.fluke.Port != nil {
// 		ui.fluke.Close()
// 	}
// 	if ui.additel.Port != nil {
// 		ui.additel.Close()
// 	}
// }

// func (ui *UI) save(log func(LogLevel, string), device Device, direction Direction, jump Jump) {
// 	err := ui.validate(device, direction, jump)
// 	if err != nil {
// 		log(ERROR, err.Error())
// 		return
// 	}

// 	switch device {
// 	case Fluke:
// 		go func() {
// 			result, err := ui.fluke.GetResult()
// 			if err != nil {
// 				fyne.Do(func() { log(ERROR, "Błąd odczytu z urządzenia: "+err.Error()) })
// 				return
// 			}

// 			ui.excelWriteCh <- WritableData{
// 				Value:     result,
// 				Direction: direction,
// 				Jump:      jump,
// 			}
// 		}()
// 	case Additel:
// 		go func() {
// 			result, err := ui.additel.GetResult()
// 			if err != nil {
// 				fyne.Do(func() { log(ERROR, "Błąd odczytu z urządzenia: "+err.Error()) })
// 				return
// 			}

// 			ui.excelWriteCh <- WritableData{
// 				Value:     result,
// 				Direction: direction,
// 				Jump:      jump,
// 			}
// 		}()
// 	}
// }

// func (ui *UI) validate(device Device, direction Direction, jump Jump) error {
// 	if !ui.excelConnected {
// 		return fmt.Errorf("nie połączono z Excelem")
// 	}

// 	if device == DeviceNone {
// 		return fmt.Errorf("nie wybrano urządzenia")
// 	}

// 	if direction == DirectionNone {
// 		return fmt.Errorf("nie wybrano kierunku")
// 	}

// 	if jump == JumpNone {
// 		return fmt.Errorf("nie wybrano kroku")
// 	}

// 	switch device {
// 	case Fluke:
// 		if ui.fluke.Port == nil {
// 			return fmt.Errorf("nie połączono z Fluke")
// 		}
// 	case Additel:
// 		if ui.additel.Port == nil {
// 			return fmt.Errorf("nie połączono z Additelem")
// 		}
// 	}

// 	return nil
// }
