package logs

import (
	"fmt"
	"log"
)

type Logs struct {

}

func (l Logs) Debug(format string, a ...interface{}) {
	log.Printf("[DEBUG] %s", fmt.Sprintf(format, a))
}

func (l Logs) Error(format string, a ...interface{}) {
	log.Printf("[ERROR] %s", fmt.Sprintf(format, a))
}