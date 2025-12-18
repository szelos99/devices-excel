package main

import (
	"fmt"
	"runtime"

	"fyne.io/fyne/v2"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type WritableData struct {
	Value     string
	Direction Direction
	Step      Step
}

// ExcelApp holds the COM connection to Excel
type ExcelApp struct {
	App     *ole.IDispatch
	unknown *ole.IUnknown
}

func RunExcelWriter(log func(LogLevel, string), writeChan <-chan WritableData, closeChan <-chan struct{}, connected *bool) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	e, err := ConnectToExcel()
	if err != nil {
		fyne.Do(func() { log(ERROR, "Błąd połączenia z Excelem: "+err.Error()) })
		return
	}
	fyne.Do(func() { log(INFO, "Połączono z Excelem") })
	*connected = true
	for {
		select {
		case writeVar := <-writeChan:
			err := e.WriteAndMove(writeVar.Value, writeVar.Direction, writeVar.Step)
			if err != nil {
				fyne.Do(func() { log(ERROR, "Błąd zapisu do Excela: "+err.Error()) })
			}
			fyne.Do(func() { log(INFO, "Zapisano do excela warosść:"+writeVar.Value) })
		case <-closeChan:
			e.Close()
			return
		}
	}
}

// Connect attaches to an already running Excel instance
func ConnectToExcel() (*ExcelApp, error) {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return nil, err
	}

	unknown, err := oleutil.GetActiveObject("Excel.Application")
	if err != nil {
		ole.CoUninitialize()
		return nil, fmt.Errorf("no active Excel instance found: %v", err)
	}

	app, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		ole.CoUninitialize()
		return nil, err
	}

	return &ExcelApp{App: app, unknown: unknown}, nil
}

// Close releases COM + Excel reference
func (e *ExcelApp) Close() {
	if e.App != nil {
		e.App.Release()
	}
	if e.unknown != nil {
		e.unknown.Release()
	}
	ole.CoUninitialize()
}

// WriteToSelectedCell writes a string to the currently selected Excel cell
func (e *ExcelApp) WriteToSelectedCell(value string) (string, error) {
	selectedVar, err := oleutil.GetProperty(e.App, "Selection")
	if err != nil {
		return "", err
	}
	defer selectedVar.Clear()

	selection := selectedVar.ToIDispatch()
	defer selection.Release()

	addressVar, err := oleutil.GetProperty(selection, "Address", 0, 0)
	if err != nil {
		return "", err
	}
	defer addressVar.Clear()
	address := addressVar.ToString()

	if _, err := oleutil.PutProperty(selection, "Value", value); err != nil {
		return "", err
	}

	return address, nil
}

// MoveSelection moves the active cell by rowOffset (+1 down, -1 up) and colOffset
func (e *ExcelApp) MoveSelectionv1(rowOffset, colOffset int) error {
	selectionVar, err := oleutil.GetProperty(e.App, "Selection")
	if err != nil {
		return err
	}
	defer selectionVar.Clear()

	selection := selectionVar.ToIDispatch()
	defer selection.Release()

	nextVar, err := oleutil.CallMethod(selection, "Offset", rowOffset, colOffset)
	if err != nil {
		return err
	}
	defer nextVar.Clear()

	next := nextVar.ToIDispatch()
	defer next.Release()

	_, err = oleutil.CallMethod(next, "Select")
	return err
}

func (e *ExcelApp) MoveSelection(rowOffset, colOffset int) error {
	selectionVar, err := oleutil.GetProperty(e.App, "Selection")
	if err != nil {
		return err
	}
	defer selectionVar.Clear()
	selection := selectionVar.ToIDispatch()
	defer selection.Release()

	// WAŻNE: Offset to Property, nie Method
	nextVar, err := oleutil.GetProperty(selection, "Offset", rowOffset, colOffset)
	if err != nil {
		return fmt.Errorf("offset failed: %w", err)
	}
	defer nextVar.Clear()

	next := nextVar.ToIDispatch()
	defer next.Release()

	_, err = oleutil.CallMethod(next, "Select")
	if err != nil {
		return fmt.Errorf("select failed: %w", err)
	}

	return nil
}

func (e *ExcelApp) WriteAndMove(value string, direction Direction, step Step) error {
	_, err := e.WriteToSelectedCell(value)
	if err != nil {
		return err
	}

	switch direction {
	case DirectionUp:
		err = e.MoveSelection(-1*int(step), 0)
	case DirectionDown:
		err = e.MoveSelection(1*int(step), 0)
	default:
		err = nil
	}

	if err != nil {
		return err
	}

	return nil
}
