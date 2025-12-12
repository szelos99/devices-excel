package devices

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"go.bug.st/serial"
)

const (
	//AdtIDNCmd Command = "*IDN?"
	AdtValCmd Command = "255:R:MRMD"
)

var Additelmode = &serial.Mode{
	BaudRate: 9600,
	DataBits: 8,
	Parity:   serial.NoParity,
	StopBits: serial.TwoStopBits,
}

var additelRegex = regexp.MustCompile(`(\d+\.\d+)`)
var additelIDNRegex = regexp.MustCompile(`^\d{3}:[A-Za-z]:[A-Za-z]{4}:\s*\d+\.\d+:[A-Za-z]+$`)

type Additel struct {
	Port serial.Port
}

func NewAdditel() *Additel {
	return &Additel{}
}

func (a *Additel) GetResult() (string, error) {
	err := writeCmd(a.Port, AdtValCmd)
	if err != nil {
		return "", fmt.Errorf("błąd wysyłania komendy do urządzenia: %w", err)
	}
	value, err := readValWithTimeout(a.Port, 1 * time.Second)
	if err != nil {
		return "", fmt.Errorf("błąd odczytu z urządzenia: %w", err)
	}

	result := additelRegex.FindString(value)
	return result, nil
}

func (a *Additel) Connect(log func(int, string)) error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return fmt.Errorf("nie udało się pobrać listy portów szeregowych: %v", err)
	}
	if len(ports) == 0 {
		return fmt.Errorf("nie znaleziono portów szeregowych")
	}

	for _, p := range ports {
		port, err := serial.Open(p, Additelmode)
		if err != nil {
			continue
		}
		if err := port.SetReadTimeout(2 * time.Second); err != nil {
			return fmt.Errorf("set timeout: %v", err)
		}
		fyne.Do(func() {log(0, "wysłano komendę: "+string(AdtValCmd))})
		err = writeCmd(port, AdtValCmd)
		if err != nil {
			fyne.Do(func() {log(1, "błąd wysyłania komendy do urządzenia: "+err.Error())})
			port.Close()
			continue
		}

		response, err := readValWithTimeout(port, 1 * time.Second)
		if err != nil {
			fyne.Do(func() {log(1, "próba odczytana odpowiedzi zwróciła błąd: "+err.Error()+". Zwrócona odpowiedź: "+response)})
			port.Close()
			continue
		}
		fyne.Do(func() {log(0, "odczytano z portu: "+response)})
		if checkCorrectIDNResponse(response) {
			a.Port = port
			return nil
		}
		port.Close()
	}

	return fmt.Errorf("nie znaleziono additel na podanych portach")
}

func checkCorrectIDNResponse(response string) bool {
	cleanResponse := strings.ReplaceAll(response, " ", "")
	return additelIDNRegex.MatchString(cleanResponse)
}

func (a *Additel) Close() {
	if a.Port != nil {
		a.Port.Close()
	}
}