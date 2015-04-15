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
	var expF40 uint64 = expFibonacci(44)
	var expDeltaT int64 = PollTimer("t1")
	ResetTimer("t1")
	var expParF40 uint64 = expParFibonacci(44)
	EndTimer("t1")
	var expParDeltaT int64 = GetTimerDelta("t1")
	StartTimer("t2")
	var fastF40reset uint64 = fastFibonacci(44)
	var fastDeltaTreset int64 = ResetTimer("t2")
	DeleteTimer("t2")
	StartTimer("t3")
	var fastF40poll uint64 = fastFibonacci(44)
	var fastDeltaTpoll int64 = PollTimer("t3")
	EndTimer("total")
	t.Logf("Computed expF44=%v in %v ns\n", expF40, expDeltaT)
	t.Logf("Computed expParF44=%v in %v ns\n", expParF40, expParDeltaT)
	t.Logf("Computed fastF44reset=%v in %v ns\n", fastF40reset, fastDeltaTreset)
	t.Logf("Computed fastF44poll=%v in %v ns\n", fastF40poll, fastDeltaTpoll)
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
			expVal = expFibonacci(45)
			EndFileTimer("timer1")
			exp <- true
		}()
	go func () {
			expParVal = expParFibonacci(45)
			EndFileTimer("timer2")
			expPar <- true
		}()
	go func () {
			fastVal = fastFibonacci(45)
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
	t.Logf("Computed expF44=%v in %v ns\n", expVal, expTime)
	t.Logf("Computed expParF44=%v in %v ns\n", expParVal, expParTime)
	t.Logf("Computed fastF44=%v in %v ns\n", fastVal, fastTime)
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

/* I'm trying to see how much faster it is to poll a timer than to stop it. */
func BenchmarkHashTableTimersPoll(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ { // a roundabout way of doing it, but I want to make sure it's comparable to BenchmarkHashTableTimersDelete
		StartTimer("t1")
		b.StartTimer()
		PollTimer("t1")
		b.StopTimer()
		DeleteTimer("t1")
	}
}

func BenchmarkHashTableTimersDelete(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		StartTimer("t1")
		b.StartTimer()
		DeleteTimer("t1")
		b.StopTimer()
	}
}

func BenchmarkHashTableTimersEnd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StartTimer("t1")
		b.StartTimer()
		EndTimer("t1")
		b.StopTimer()
		DeleteTimer("t1")
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
		StartFileTimer("timerB")
		b.StartTimer()
		EndFileTimer("timerB")
		b.StopTimer()
		DeleteFileTimer("timerB")
	}
}
