package timers

import "os"
import "runtime"
import "testing"

func expFibonacci(n uint64) uint64 {
	if n < 2 {
		return n
	}
	return expFibonacci(n - 2) + expFibonacci(n - 1)
}

func wrapChannel(fn func (uint64) uint64, arg uint64, ret chan uint64) {
	ret <- fn(arg)
}

func expParFibonacci(n uint64) uint64 {
	if n < 2 {
		return n
	}
	var subpr chan uint64 = make(chan uint64, 2)
	go wrapChannel(expFibonacci, n - 2, subpr)
	go wrapChannel(expFibonacci, n - 1, subpr)
	return <-subpr + <-subpr
}

func fastFibonacci(n uint64) uint64 {
	var curr uint64 = 0
	var prev uint64 = 1
	for ; n > 0; n-- {
		curr, prev = curr + prev, curr
	}
	return curr
}

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	os.Exit(m.Run())
}

func TestHashTableTimers1(t *testing.T) {
	StartTimer("total")
	StartTimer("t1")
	var expF40 uint64 = expFibonacci(40)
	var expDeltaT int64 = PollTimer("t1")
	ResetTimer("t1")
	var expParF40 uint64 = expParFibonacci(40)
	EndTimer("t1")
	var expParDeltaT int64 = GetTimerDelta("t1")
	StartTimer("t2")
	var fastF40reset uint64 = fastFibonacci(40)
	var fastDeltaTreset int64 = ResetTimer("t2")
	DeleteTimer("t2")
	StartTimer("t3")
	var fastF40poll uint64 = fastFibonacci(40)
	var fastDeltaTpoll int64 = PollTimer("t3")
	EndTimer("total")
	t.Logf("Computed expF40=%v in %v ns\n", expF40, expDeltaT)
	t.Logf("Computed expParF40=%v in %v ns\n", expParF40, expParDeltaT)
	t.Logf("Computed fastF40reset=%v in %v ns\n", fastF40reset, fastDeltaTreset)
	t.Logf("Computed fastF40poll=%v in %v ns\n", fastF40poll, fastDeltaTpoll)
	t.Logf("Total time taken: %v", GetTimerDelta("total"))
	DeleteTimer("t3") // Free memory
	DeleteTimer("t1")
}

func TestHashTableTimers2(t *testing.T) {
	StartTimer("t1")
	StartTimer("t2")
	DeleteTimer("t1")
	StartTimer("t3")
	DeleteTimer("t2")
	StartTimer("t1")
	DeleteTimer("t1")
	StartTimer("t4")
	StartTimer("t1")
	DeleteTimer("t4")
	DeleteTimer("t1")
	DeleteTimer("t3")
}

func TestHashTableTimers3(t *testing.T) {
	var finished bool = false
	defer func () {
			r := recover()
			if r == nil || !finished {
				t.Fail()
			}
			DeleteTimer("t1")
		}()
	StartTimer("t1")
	DeleteTimer("t1")
	StartTimer("t1")
	finished = true
	StartTimer("t1")
}

func TestHashTableTimers4(t *testing.T) {
	var finished bool = false
	defer func () {
			r := recover()
			if r == nil || !finished {
				t.Fail()
			}
			DeleteTimer("t2")
			DeleteTimer("t4")
		}()
	StartTimer("t1")
	StartTimer("t2")
	StartTimer("t3")
	DeleteTimer("t1")
	PollTimer("t3")
	PollTimer("t2")
	DeleteTimer("t3")
	StartTimer("t4")
	finished = true
	PollTimer("t3")
}

func TestHashTableTimers5(t *testing.T) {
	var finished bool = false
	defer func () {
			r := recover()
			if r == nil || !finished {
				t.Fail()
			}
			DeleteTimer("t3")
			DeleteTimer("t4")
		}()
	StartTimer("t1")
	StartTimer("t3")
	StartTimer("t2")
	DeleteTimer("t1")
	PollTimer("t3")
	ResetTimer("t2")
	StartTimer("t4")
	DeleteTimer("t2")
	finished = true
	ResetTimer("t1")
}

func TestHashTableTimers6(t *testing.T) {
	var finished bool = false
	defer func () {
			r := recover()
			if r == nil || !finished {
				t.Fail()
			}
			DeleteTimer("t1")
		}()
	StartTimer("t1")
	DeleteTimer("t1")
	StartTimer("t1")
	StartTimer("t2")
	EndTimer("t1")
	DeleteTimer("t2")
	finished = true
	EndTimer("t1")
}

func TestFileTimers1(t *testing.T) {
	SetFileTimerCollection("/home/sam/timers")
	var exp chan bool = make(chan bool)
	var expPar chan bool = make(chan bool)
	var fast chan bool = make(chan bool)
	var expVal uint64
	var expParVal uint64
	var fastVal uint64
	StartFileTimer("timer1")
	StartFileTimer("timer2")
	StartFileTimer("timer3")
	go func () {
			expVal = expFibonacci(42)
			EndFileTimer("timer1")
			exp <- true
		}()
	go func () {
			expParVal = expParFibonacci(42)
			EndFileTimer("timer2")
			expPar <- true
		}()
	go func () {
			fastVal = fastFibonacci(42)
			EndFileTimer("timer3")
			fast <- true
		}()
		
	for i := 0; i < 3; i++ {
		select {
			case <-exp:
			case <-expPar:
			case <-fast:
		}
	}
	var expTime int64 = GetFileTimerDelta("timer1")
	var expParTime int64 = GetFileTimerDelta("timer2")
	var fastTime int64 = GetFileTimerDelta("timer3")
	DeleteFileTimer("timer1")
	DeleteFileTimer("timer2")
	DeleteFileTimer("timer3")
	t.Logf("Computed expF42=%v in %v ns\n", expVal, expTime)
	t.Logf("Computed expParF42=%v in %v ns\n", expParVal, expParTime)
	t.Logf("Computed fastF42=%v in %v ns\n", fastVal, fastTime)
}

// These test are similar to those for the hash table timers

func TestFileTimers2(t *testing.T) {
	StartFileTimer("t1")
	StartFileTimer("t2")
	DeleteFileTimer("t1")
	StartFileTimer("t3")
	DeleteFileTimer("t2")
	StartFileTimer("t1")
	DeleteFileTimer("t1")
	StartFileTimer("t4")
	StartFileTimer("t1")
	DeleteFileTimer("t4")
	DeleteFileTimer("t1")
	DeleteFileTimer("t3")
}

func TestFileTimers3(t *testing.T) {
	StartFileTimer("t1")
	DeleteFileTimer("t1")
	StartFileTimer("t1")
	StartFileTimer("t1")
}

func TestFileTimers4(t *testing.T) {
	var finished bool = false
	defer func () {
			r := recover()
			if r == nil || !finished {
				t.Fail()
			}
			DeleteFileTimer("t2")
			DeleteFileTimer("t4")
		}()
	StartFileTimer("t1")
	StartFileTimer("t2")
	StartFileTimer("t3")
	DeleteFileTimer("t1")
	PollFileTimer("t3")
	PollFileTimer("t2")
	DeleteFileTimer("t3")
	StartFileTimer("t4")
	finished = true
	PollFileTimer("t3")
}

func TestFileTimers5(t *testing.T) {
	StartFileTimer("t1")
	StartFileTimer("t3")
	StartFileTimer("t2")
	DeleteFileTimer("t1")
	PollFileTimer("t3")
	StartFileTimer("t2")
	StartFileTimer("t4")
	DeleteFileTimer("t2")
	StartFileTimer("t1")
	DeleteFileTimer("t3")
}

func TestFileTimers6(t *testing.T) {
	var finished bool = false
	defer func () {
			r := recover()
			if r == nil || !finished {
				t.Fail()
			}
		}()
	StartFileTimer("t1")
	StartFileTimer("t2")
	DeleteFileTimer("t2")
	var delta int64 = GetFileTimerDelta("t1")
	if delta != -2 {
		t.Logf("Part 1: delta is %v", delta)
		t.Fail()
		DeleteFileTimer("t1")
		return
	}
	EndFileTimer("t1")
	EndFileTimer("t3")
	delta = GetFileTimerDelta("t3")
	if delta != -1 {
		t.Logf("Part 2: delta is %v", delta)
		t.Fail()
		DeleteFileTimer("t1")
		DeleteFileTimerIfExists("t3")
	}
	DeleteFileTimerIfExists("t3")
	delta = GetFileTimerDelta("t1")
	if delta < 0 {
		t.Logf("Part 3: delta is %v", delta)
		t.Fail()
		DeleteFileTimer("t1")
		return
	}
	DeleteFileTimer("t1")
	StartFileTimer("t1")
	StartFileTimer("t2")
	DeleteFileTimer("t1")
	DeleteFileTimer("t2")
	finished = true
	DeleteFileTimer("t1")
}

func TestLogTimers1(t *testing.T) {
	SetLogFile("/home/sam/timers/logtimer1")
	StartLogTimer("fastfib")
	var f41f uint64 = fastFibonacci(41)
	EndLogTimer("fastfib")
	StartLogTimer("slowfib")
	var f41s uint64 = expFibonacci(41)
	EndLogTimer("slowfib")
	CloseLogFile()
	var timers map[string]*TimerSummary = ParseFileToMap([]string{"/home/sam/timers/logtimer1"})
	var deltas map[string][]int64 = ParseMapToDeltas(timers)
	t.Logf("Fast fib 41 is %v: computed in %v ns", f41f, deltas["fastfib"][0])
	t.Logf("Slow fib 41 is %v: computed in %v ns", f41s, deltas["slowfib"][0])
}

func checkLogBuffer(t *testing.T, timers map[string]*TimerSummary) {
	t1data, ok1 := timers["t1"]
	t2data, ok2 := timers["t2"]
	t3data, ok3 := timers["t3"]
	t4data, ok4 := timers["t4"]
	t5data, ok5 := timers["t5"]
	t6data, ok6 := timers["t6"]
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
		t.Log("Some timers are missing")
		t.Fail()
		return
	}
	t1s := len(t1data.starts)
	t2s := len(t2data.starts)
	t3s := len(t3data.starts)
	t4s := len(t4data.starts)
	t5s := len(t5data.starts)
	t6s := len(t6data.starts)
	t1e := len(t1data.ends)
	t2e := len(t2data.ends)
	t3e := len(t3data.ends)
	t4e := len(t4data.ends)
	t5e := len(t5data.ends)
	t6e := len(t6data.ends)
	if t1s != 2 || t2s != 2 || t3s != 2 || t4s != 0 || t5s != 1 || t6s != 2 {
		t.Log("Bad start counts")
		t.Fail()
		return
	}
	if t1e != 2 || t2e != 2 || t3e != 2 || t4e != 1 || t5e != 0 || t6e != 2 {
		t.Log("Bad end counts")
		t.Fail()
		return
	}
	var deltas map[string][]int64 = ParseMapToDeltas(timers)
	_, ex1 := deltas["t1"]
	_, ex2 := deltas["t2"]
	t3deltas, ex3 := deltas["t3"]
	_, ex4 := deltas["t4"]
	_, ex5 := deltas["t5"]
	_, ex6 := deltas["t6"]
	if ex1 || ex2 || !ex3 || ex4 || ex5 || ex6 || len(t3deltas) != 2 {
		t.Log("Map parsed into deltas incorrectly")
		t.Fail()
		return
	}
}

func TestLogTimers2(t *testing.T) {
	SetLogFile("/home/sam/timers/logtimer2")
	StartLogTimer("t1")
	StartLogTimer("t2")
	EndLogTimer("t1")
	EndLogTimer("t2")
	EndLogTimer("t1")
	StartLogTimer("t3")
	StartLogTimer("t6")
	StartLogTimer("t1")
	CloseLogFile()
	SetLogFile("/home/sam/timers/logtimer3")
	StartLogTimer("t6")
	StartLogTimer("t5")
	EndLogTimer("t2")
	EndLogTimer("t6")
	EndLogTimer("t3")
	StartLogTimer("t2")
	EndLogTimer("t6")
	StartLogTimer("t3")
	EndLogTimer("t4")
	EndLogTimer("t3")
	CloseLogFile()
	SetLogFile("/home/sam/timers/logtimer4")
	CloseLogFile()
	var timers map[string]*TimerSummary = ParseFileToMap([]string{"/home/sam/timers/logtimer2", "/home/sam/timers/logtimer3", "/home/sam/timers/logtimer4"})
	checkLogBuffer(t, timers)
}

// Nearly identical to the previous test, but tests the buffer
func TestLogTimers3(t *testing.T) {
	StartBufferedLogTimer("t1")
	StartBufferedLogTimer("t2")
	EndBufferedLogTimer("t1")
	EndBufferedLogTimer("t2")
	EndBufferedLogTimer("t1")
	StartBufferedLogTimer("t3")
	StartBufferedLogTimer("t6")
	StartBufferedLogTimer("t1")
	StartBufferedLogTimer("t6")
	StartBufferedLogTimer("t5")
	EndBufferedLogTimer("t2")
	EndBufferedLogTimer("t6")
	EndBufferedLogTimer("t3")
	StartBufferedLogTimer("t2")
	EndBufferedLogTimer("t6")
	StartBufferedLogTimer("t3")
	EndBufferedLogTimer("t4")
	EndBufferedLogTimer("t3")
	var timers map[string]*TimerSummary = GetLogBuffer()
	checkLogBuffer(t, timers)
	var f *os.File
	f, _ = os.Create("/home/sam/timers/bufferedlog")
	WriteLogBuffer(f)
	f.Close()
	timers = ParseFileToMap([]string{"/home/sam/timers/bufferedlog"})
	checkLogBuffer(t, timers)
	ResetLogBuffer()
	f, _ = os.Create("/home/sam/timers/bufferedlog")
	WriteLogBuffer(f)
	f.Close()
	var timers2 map[string]*TimerSummary = ParseFileToMap([]string{"/home/sam/timers/bufferedlog"})
	if len(timers2) != 0 {
		t.Fail()
	}
	SetLogBuffer(timers)
	f, _ = os.Create("/home/sam/timers/bufferedlog2")
	WriteLogBuffer(f)
	f.Close()
	timers2 = ParseFileToMap([]string{"/home/sam/timers/bufferedlog2"})
	checkLogBuffer(t, timers2)
}

/* I'm trying to see how much faster it is to poll a timer than to stop it. */
func BenchmarkHashTableTimersPoll(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping test in short mode.")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ { // a roundabout way of doing it, but I want to make sure it's comparable to BenchmarkHashTableTimersDelete
		StartTimer("t1")
		b.StartTimer()
		PollTimer("t1")
		b.StopTimer()
		DeleteTimer("t1")
	}
}

func BenchmarkHashTableTimersDelete(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping test in short mode.")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StartTimer("t1")
		b.StartTimer()
		DeleteTimer("t1")
		b.StopTimer()
	}
}

func BenchmarkHashTableTimersEnd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StartTimer("t2")
		b.StartTimer()
		EndTimer("t2")
		b.StopTimer()
		DeleteTimer("t2")
	}
}

func BenchmarkHashTableTimersStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		StartTimer("t3")
		b.StopTimer()
		EndTimer("t3")
		DeleteTimer("t3")
	}
}

func BenchmarkFileTimersPoll(b *testing.B) {
	StartFileTimer("t1")
	b.ResetTimer();
	for i := 0; i < b.N; i++ {
		PollFileTimer("t1")
	}
	b.StopTimer()
	DeleteFileTimer("t1")
}

func BenchmarkFileTimersEnd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StartFileTimer("timerA")
		b.StartTimer()
		EndFileTimer("timerA")
		b.StopTimer()
	}
	DeleteFileTimer("timerA")
}

func BenchmarkFileTimersStart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		StartFileTimer("timerB")
		b.StopTimer()
		EndFileTimer("timerB")
	}
	DeleteFileTimer("timerB")
}

func BenchmarkLogTimersEnd(b *testing.B) {
	SetLogFile("/home/sam/timers/logfile1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StartLogTimer("timerC")
		b.StartTimer()
		EndLogTimer("timerC")
		b.StopTimer()
	}
	CloseLogFile()
}

func BenchmarkLogTimersStart(b *testing.B) {
	SetLogFile("/home/sam/timers/logfile2")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		StartLogTimer("timerD")
		b.StopTimer()
		EndLogTimer("timerD")
	}
	CloseLogFile()
}
