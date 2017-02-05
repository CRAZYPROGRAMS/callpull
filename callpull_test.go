package callpull

import (
	"testing"
	"time"
)

func tListenTimeout(t *testing.T, tInterval time.Duration) {
	pull := NewCallPull()
	start := time.Now()
	Param, Result, err := pull.Listen("Test", tInterval)
	if err != ErrorTimeout || Param != nil || Result != nil {
		t.Error("Unknown error")
	}
	timeout := time.Now().Sub(start)
	delta := timeout - tInterval
	if delta < 0 {
		t.Error("Сanceled before timeout")
	}
	if delta > tInterval+tInterval/10 {
		t.Error("Too long")
	}
}
func TestListenTimeout(t *testing.T) {
	tListenTimeout(t, time.Millisecond*100)
	tListenTimeout(t, time.Millisecond*200)
	tListenTimeout(t, time.Millisecond*300)
	tListenTimeout(t, time.Millisecond*400)
	tListenTimeout(t, time.Millisecond*500)
}
func tCallTimeout(t *testing.T, tInterval time.Duration) {
	pull := NewCallPull()
	start := time.Now()
	Result, err := pull.Call("Test", map[string]interface{}{"0": "Param1", "1": "Param2"}, tInterval)
	if err != ErrorTimeout || Result != nil {
		t.Error("Unknown error")
	}
	timeout := time.Now().Sub(start)
	delta := timeout - tInterval
	if delta < 0 {
		t.Error("Сanceled before timeout")
	}
	if delta > tInterval+tInterval/10 {
		t.Error("Too long")
	}
}
func TestCallTimeout(t *testing.T) {
	tCallTimeout(t, time.Millisecond*100)
	tCallTimeout(t, time.Millisecond*200)
	tCallTimeout(t, time.Millisecond*300)
	tCallTimeout(t, time.Millisecond*400)
	tCallTimeout(t, time.Millisecond*500)
}
func tPull(t *testing.T, tIntervalCall time.Duration, tIntervalWork time.Duration, Name string, Param map[string]interface{}) (interface{}, bool, error) {
	var brake bool = false
	done := make(chan bool)
	pull := NewCallPull()
	worker := func(Name string) {
		process := func(Param map[string]interface{}, Result chan interface{}) {
			defer func() {
				if r := recover(); r != nil {
					brake = true
				}
				done <- true
			}()
			P1 := Param["0"].(string)
			P2 := Param["1"].(string)
			time.Sleep(tIntervalWork)
			Result <- "result" + Name + P1 + P2
		}
		for {
			Param, Result, err := pull.Listen(Name, time.Minute)
			if err == nil {
				process(Param, Result)
			} else {
				t.Error("Timout listen")
			}
		}
	}
	go worker(Name)
	Result, err := pull.Call(Name, Param, tIntervalCall)
	<-done
	return Result, brake, err
}
func TestPull(t *testing.T) {
	Result, Brake, err := tPull(t, time.Millisecond*200, time.Millisecond*100, "Test1", []interface{}{"Param1", "Param2"})
	if Brake {
		t.Error("Brake")
	}
	if err != nil {
		t.Error(err)
	}
	val, ok := Result.(string)
	if !ok || val != "resultTest1Param1Param2" {
		t.Error("Invalid result")
	}
}
func TestPullBrake(t *testing.T) {
	Result, Brake, err := tPull(t, time.Millisecond*100, time.Millisecond*200, "Test1", []interface{}{"Param1", "Param2"})
	if !Brake {
		t.Error("Brake not made")
	}
	if err == nil {
		t.Error("No error")
	}
	val, ok := Result.(string)
	if ok && val != "resultTest1Param1Param2" {
		t.Error("not invalid result")
	}
}
