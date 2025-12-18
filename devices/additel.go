package devices

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.bug.st/serial"
)

const (
	//AdtIDNCmd Command = "*IDN?"
	AdtValCmd Command = "255:R:MRMD:1"
)

var Additelmode = &serial.Mode{
	BaudRate: 9600,
	DataBits: 8,
	Parity:   serial.NoParity,
	StopBits: serial.TwoStopBits,
}

var OneStopBitMode = &serial.Mode{
	BaudRate: 9600,
	DataBits: 8,
	Parity:   serial.NoParity,
	StopBits: serial.OneStopBit,
}

var additelRegex = regexp.MustCompile(`(\d+\.\d+)`)
var additelIDNRegex = regexp.MustCompile(`^\d{3}:[A-Za-z]:[A-Za-z]{4}:\s*\d+\.\d+:[A-Za-z]+$`)

type Additel struct {
	Port serial.Port
}

func NewAdditel() *Additel {
	return &Additel{}
}

func (a *Additel) GetResult(log func(int, string)) (string, error) {
	err := writeCmd(a.Port, AdtValCmd)
	if err != nil {
		return "", fmt.Errorf("błąd wysyłania komendy do urządzenia: %w", err)
	}
	value, err := readValWithTimeout(a.Port, 1 * time.Second)
	if err != nil {
		return "", fmt.Errorf("błąd odczytu z urządzenia: %w", err)
	}
	if value == "" {

	}
	log(0, fmt.Sprintf("Odczytana wartość z urządzenia: %s", value))
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

	modes := map[string]*serial.Mode{
		"dwa bity stopu": Additelmode,
		"jeden biy stopu": OneStopBitMode,
	}
	for _, p := range ports {
		for key, mode := range modes {
			log(0, "Próļa połączenia dla: "+key)
			port, err := serial.Open(p, mode)
			if err != nil {
				continue
			}
			log(0, "wysłano komendę: "+string(AdtValCmd))
			err = writeCmd(port, AdtValCmd)
			if err != nil {
				log(1, "błąd wysyłania komendy do urządzenia: "+err.Error())
				port.Close()
				continue
			}

			response, err := readValWithTimeout(port, 2 * time.Second)
			if err != nil {
				log(1, "próba odczytana odpowiedzi zwróciła błąd: "+err.Error()+". Zwrócona odpowiedź: "+response)
				port.Close()
				continue
			}
			log(0, "odczytano z portu: "+response)
			if checkCorrectIDNResponse(response) {
				a.Port = port
				return nil
			}
			port.Close()
		}
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