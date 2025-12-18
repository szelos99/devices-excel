package devices

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Command string

func writeCmd(port serial.Port, cmd Command) error {
	full := string(cmd) + "\r\n"
	_, err := port.Write([]byte(full))
	return err
}

func readValWithTimeout(port serial.Port, timeout time.Duration) (string, error) {
	reader := bufio.NewReader(port)
	resultChan := make(chan string)
	errChan := make(chan error)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		fmt.Println("running:" + time.Now().Format("15:04:05"))
		for {
			select {
			case <-ctx.Done():
				return
			default:
				s, err := reader.ReadString('\r')
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- strings.TrimSpace(s)
				return
			}
		}
	}()

	select {
	case res := <-resultChan:
		return res, nil
	case err := <-errChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("read timeout")
	}
}