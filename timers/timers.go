package timers

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

/* LOG-BASED TIMERS */

var file *os.File = nil

func SetLogFile(filepath string) {
	if file != nil {
		file.Close()
	}
	var err error
	file, err = os.Create(filepath)
	if err != nil {
		panic(fmt.Sprintf("Attempted to set log to invalid filepath %v", err))
	}
}

func CloseLogFile() {
	if file == nil {
		panic(fmt.Sprintf("Attempted to close log file, but not log file is active"))
	} else {
		file.Close()
		file = nil
	}
}

func logEvent(name string, tag string) {
	_, err := file.WriteString(fmt.Sprintf("\x00%s\x00%s\x00", name, tag))
	var currTime int64 = time.Now().UnixNano()
	if err == nil {
		err = binary.Write(file, binary.LittleEndian, currTime)
		if err != nil {
			panic(fmt.Sprintf("Failed to write current time to file: %v", err))
		}
	} else {
		panic(fmt.Sprintf("Failed to write timer name to file: %v", err))
	}	
}

const (
	START_SYMBOL string = "s"
	END_SYMBOL string = "e"
	)

/** Name can't contain \0. */
func StartLogTimer(name string) {
	logEvent(name, START_SYMBOL)
}

func EndLogTimer(name string) {
	logEvent(name, END_SYMBOL)
}

type TimerSummary struct {
	starts []int64
	ends []int64
}

func ParseFileToMap(filenames []string) map[string]*TimerSummary {
	var data [][]byte = make([][]byte, len(filenames))
	for i := 0; i < len(filenames); i++ {
		f, err := os.Open(filenames[i])
		if err != nil {
			f.Close()
			panic(fmt.Sprintf("Attempted to parse file at invalid filepath %s", filenames[i]))
		}
		data[i], err = ioutil.ReadAll(f) // it's OK to buffer everything in memory since I'm constructing a hashtable out of it anyway
		f.Close()
		if err != nil {
			panic(fmt.Sprintf("Could not read file at filepath %s", filenames[i]))
		}
	}
	var tmap map[string]*TimerSummary = make(map[string]*TimerSummary)
	var fragments []string
	var name string
	var summary *TimerSummary
	var ok bool
	var time int64
	for i := 0; i < len(filenames); i++ {
		fragments = strings.Split(string(data[i]), "\x00")
		if (len(fragments) % 3 == 1) {
			for j := 1; j < len(fragments); j += 3 {
				name = fragments[j]
				summary, ok = tmap[name]
				if !ok {
					summary = &TimerSummary{make([]int64, 0, 1), make([]int64, 0, 1)}
					tmap[name] = summary
				}
				binary.Read(strings.NewReader(fragments[j + 2]), binary.LittleEndian, &time)
				if fragments[j + 1] == START_SYMBOL {
					summary.starts = append(summary.starts, time)
				} else {
					summary.ends = append(summary.ends, time)
				}
			}
		} else {
			panic(fmt.Sprintf("Log file %s is malformed: has %v fragments", filenames[i], len(fragments)))
		}
	}
	return tmap
}

func ParseMapToDeltas(tmap map[string]*TimerSummary) map[string][]int64 {
	var tname string
	var tsummary *TimerSummary
	var deltamap map[string][]int64 = make(map[string][]int64)
	var i int
	
	var deltas []int64
	
	TimerLoop:
		for tname, tsummary = range tmap {
			if len(tsummary.starts) != len(tsummary.ends) {
				fmt.Printf("Timer %s has a different number of starts than ends\n", tname)
				continue
			} else if len(tsummary.starts) == 0 {
				fmt.Printf("Timer %s was ended but never started\n", tname)
				continue
			} else if len(tsummary.ends) == 0 {
				fmt.Printf("Timer %s was started but never ended\n", tname)
				continue
			}
			deltas = make([]int64, len(tsummary.starts))
			for i = 0; i < len(tsummary.ends); i++ {
				if false && tsummary.starts[i] > tsummary.ends[i] {
					fmt.Printf("Timer %s has an end time preceding start time\n", tname)
					continue TimerLoop
				}
				if false && i > 1 && tsummary.starts[i] < tsummary.ends[i - 1] {
					fmt.Printf("Timer %s was started twice without being ended in between\n")
					continue TimerLoop
				}
				deltas[i] = tsummary.ends[i] - tsummary.starts[i]
			}
			deltamap[tname] = deltas
		}
		
	return deltamap
}
