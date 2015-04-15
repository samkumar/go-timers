package timers

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	)

/* HASHTABLE-BASED TIMERS */

var timers map[string]int64 = make(map[string]int64)
var timersEnd map[string]int64 = make(map[string]int64)

func StartTimer(name string) {
	if _, ok := timers[name]; ok {
		panic(fmt.Sprintf("Attempted to start running timer %s", name))
	} else {
		timers[name] = time.Now().UnixNano()
	}
}

func EndTimer(name string) {
	if _, ok := timersEnd[name]; ok {
		panic(fmt.Sprintf("Attempted to end stopped timer %s", name))
	} else {
		timersEnd[name] = time.Now().UnixNano()
	}
}

func GetTimerDelta(name string) int64 {
	if valStart, ok := timers[name]; ok {
		if valEnd, ok := timersEnd[name]; ok {
			return valEnd - valStart
		} else {
			return -2
		}
	} else {
		return -1
	}
}

func ResetTimer(name string) int64 {
	if val, ok := timers[name]; ok {
		now := time.Now().UnixNano()
		timers[name] = now
		return now - val
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

func DeleteTimer(name string) {
	if _, ok := timers[name]; ok {
		delete(timers, name)
	} else {
		panic(fmt.Sprintf("Attempted to stop timer %s, which is not running", name))
	}
	delete(timersEnd, name)
}

/* FILE-BASED TIMERS */

var timerDir string

func SetFileTimerCollection (dirString string) {
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

func expandFilePathStart(name string) string {
	return fmt.Sprintf("%s/%s_start", timerDir, name)
}

func expandFilePathEnd(name string) string {
	return fmt.Sprintf("%s/%s_end", timerDir, name)
}

/** This will overwrite any existing timers. I didn't add error checking here
    because I reasoned that we may see some of the same timers from previous
    runs of the program. */
func StartFileTimer(name string) {
	writeFileTimer(name, expandFilePathStart)
}

func EndFileTimer(name string) {
	writeFileTimer(name, expandFilePathEnd)
}

func writeFileTimer(name string, nameFinder func (string) string) {
	file, err := os.Create(nameFinder(name))
	defer file.Close()
	if err == nil {
		err = binary.Write(file, binary.LittleEndian, time.Now().UnixNano())
		if err != nil {
			panic(fmt.Sprintf("Could not write to file timer %s: %v", nameFinder(name), err))
		}
	} else {
		panic(fmt.Sprintf("Could not write to file timer %s: %v", nameFinder(name), err))
	}
}

func readFileTimer(name string, nameFinder func (string) string) int64 {
	file, err := os.Open(nameFinder(name))
	defer file.Close()
	var fileTime int64
	if err == nil {
		err = binary.Read(file, binary.LittleEndian, &fileTime)
		if err != nil {
			panic(fmt.Sprintf("Could not poll file timer %s: %v", nameFinder(name), err))
		}
		return fileTime
	} else {
		panic(fmt.Sprintf("Could not open file timer %s: %v", nameFinder(name), err))
	}
}

func GetFileTimerDelta(name string) (delta int64) {
	var started bool = false
	defer func () {
			if r := recover(); r != nil {
				if started {
					delta = -2 // indicates timer was started but never ended
				} else {
				 	delta = -1 // indicates timer was never started
				}
			}
		}()
	var startTime int64 = readFileTimer(name, expandFilePathStart)
	started = true
	var endTime int64 = readFileTimer(name, expandFilePathEnd)
	delta = endTime - startTime
	return
}

func PollFileTimer(name string) int64 {
	return time.Now().UnixNano() - readFileTimer(name, expandFilePathStart)
}

func DeleteFileTimer(name string) {
	var err error = os.Remove(expandFilePathStart(name))
	if err != nil {
		panic(fmt.Sprintf("Could not stop file timer %s: %v", name, err))
	}
	os.Remove(expandFilePathEnd(name))
}

func DeleteFileTimerIfExists(name string) {
	os.Remove(expandFilePathStart(name))
	os.Remove(expandFilePathEnd(name))
}
