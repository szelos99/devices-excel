package devices

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.bug.st/serial"
)

const (
	FlukeIDNCmd Command = "*IDN?"
	FlukeValCmd Command = "VAL?"
)

var flukemode = &serial.Mode{
	BaudRate: 9600,
	DataBits: 8,
	Parity:   serial.NoParity,
	StopBits: serial.OneStopBit,
}

var flukeRegex = regexp.MustCompile(`[^0-9\.]`)

type Fluke struct {
	Port serial.Port
}

func NewFluke() *Fluke {
	return &Fluke{}
}

func (f *Fluke) GetResult(log func(int, string)) (string, error) {
	err := writeCmd(f.Port, FlukeValCmd)
    if err != nil {
        return "", fmt.Errorf("błąd wysyłania komendy do urządzenia: %w", err)
    }

    value, err := readValWithTimeout(f.Port, 1*time.Second)
    if err != nil {
        return "",fmt.Errorf("błąd odczytu z urządzenia: %w", err)
    }
	log(0, fmt.Sprintf("Odczytana wartość z urządzenia: %s", value))
	result := flukeRegex.ReplaceAllString(value, "")
	return result, nil
}

func (f *Fluke) Connect() error {
	ports, err := serial.GetPortsList()
	if err != nil {
		return fmt.Errorf("nie udało się pobrać listy portów szeregowych: %v", err)
	}
	if len(ports) == 0 {
		return fmt.Errorf("nie znaleziono portów szeregowych")
	}
	for _, p := range ports {
		port, err := serial.Open(p, flukemode)
		if err != nil {
			continue
		}
		err = writeCmd(port, FlukeIDNCmd)
		if err != nil {
			port.Close()
			continue
		}

		response, err := readValWithTimeout(port, 1 * time.Second)
		if err != nil {
			port.Close()
			continue
		}
		if strings.Contains(strings.ToLower(response), "fluke") {
			f.Port = port
			return nil
		}
		port.Close()
	}

	return fmt.Errorf("nie znaleziono fluke")
}

func (f *Fluke) Close() {
	if f.Port != nil {
		f.Port.Close()
	}
}