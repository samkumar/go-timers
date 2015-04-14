package timers

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	)

/* HASHTABLE-BASED TIMERS */

var timers map[string]int64 = make(map[string]int64)

func StartTimer(name string) {
	if _, ok := timers[name]; ok {
		panic(fmt.Sprintf("Attempted to start running timer %s", name))
	} else {
		timers[name] = time.Now().UnixNano()
	}
}

func ResetTimer(name string) {
	if _, ok := timers[name]; ok {
		timers[name] = time.Now().UnixNano()
	} else {
		panic(fmt.Sprintf("Attempted to reset timer %s, which is not running", name))
	}
}

func PollTimer(name string) int64 {
	if val, ok := timers[name]; ok {
		return time.Now().UnixNano() - val
	} else {
		panic(fmt.Sprintf("Attempted to poll timer %s, which is not running", name))
	}
}

func StopTimer(name string) int64 {
	if val, ok := timers[name]; ok {
		delete(timers, name)
		return time.Now().UnixNano() - val
	} else {
		panic(fmt.Sprintf("Attempted to stop timer %s, which is not running", name))
	}
}

/* FILE-BASED TIMERS */

var timerDir string

func SetCollection (dirString string) {
	fi, err := os.Stat(dirString)
	if err == nil && fi.IsDir() {
		lastIndex := len(dirString) - 1
		if dirString[lastIndex] == '/' {
			timerDir = dirString[0:lastIndex]
		} else {
			timerDir = dirString
		}
	} else {
		panic(fmt.Sprintf("Attempted to set Timer collection to invalid directory %s", dirString))
	}
}

func expandFilePath(name string) string {
	return fmt.Sprintf("%s/%s", timerDir, name)
}

/** This will overwrite any existing timers. I didn't add error checking here
    because I reasoned that we may see some of the same timers from previous
    runs of the program. */
func StartFileTimer(name string) {
	file, err := os.OpenFile(expandFilePath(name), os.O_RDWR, 0666)
	defer file.Close()
	if err == nil {
		err = binary.Write(file, binary.LittleEndian, time.Now().UnixNano())
		if err != nil {
			panic(fmt.Sprintf("Could not write to file timer %s: %v", name, err))
		}
	} else {
		panic(fmt.Sprintf("Could not start file timer %s: %v", name, err))
	}
}

func PollFileTimer(name string) int64 {
	file, err := os.Open(expandFilePath(name))
	defer file.Close()
	var fileTime int64
	if err == nil {
		err = binary.Read(file, binary.LittleEndian, &fileTime)
		if err != nil {
			panic(fmt.Sprintf("Could not poll file timer %s: %v", name, err))
		}
		return time.Now().UnixNano() - fileTime
	} else {
		panic(fmt.Sprintf("Could not open file timer %s: %v", name, err))
	}
}

func StopFileTimer(name string) {
	var err error = os.Remove(expandFilePath(name))
	if err != nil {
		panic(fmt.Sprintf("Could not stop file timer %s: %v", name, err))
	}
}
