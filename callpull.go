package callpull

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var ErrorTimeout = errors.New("call timeout")

type callName struct {
	lock   sync.Mutex
	Update chan *callName
	Name   string
	Calls  list.List
}

type callItem struct {
	Result chan interface{}
	Param  map[string]interface{}
}

func newCallName(Name string) *callName {
	cn := &callName{
		Name:   Name,
		lock:   sync.Mutex{},
		Calls:  list.List{},
		Update: make(chan *callName, 100),
	}
	return cn
}

func (cn *callName) add(Name string, Item *callItem) {
	cn.lock.Lock()
	cn.Calls.PushBack(Item)
	cn.lock.Unlock()
	go func() {
		cn.Update <- cn
	}()
}
func (cn *callName) remove(Item *callItem) bool {
	cn.lock.Lock()
	defer cn.lock.Unlock()
	for e := cn.Calls.Front(); e != nil; e = e.Next() {
		if e.Value.(*callItem) == Item {
			cn.Calls.Remove(e)
			return true
		}
	}
	return false
}
func (cn *callName) get() *callItem {
	cn.lock.Lock()
	defer cn.lock.Unlock()
	el := cn.Calls.Front()
	if el == nil {
		return nil
	}
	cn.Calls.Remove(el)
	return el.Value.(*callItem)
}

// CallPull - Manager asynchronous execution calls.
type CallPull struct {
	lock      sync.Mutex
	callnames map[string]*callName
}

// NewCallPull -  Create a new manager asynchronous execution calls.
func NewCallPull() *CallPull {
	return &CallPull{callnames: make(map[string]*callName)}
}

func (cp *CallPull) getCallName(Name string) *callName {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	n, ok := cp.callnames[Name]
	if !ok {
		n = newCallName(Name)
		cp.callnames[Name] = n
	}
	return n
}

// Listen - Wait for the request to perform call
func (cp *CallPull) Listen(Name string, timeOutWait time.Duration) (Param map[string]interface{}, Result chan interface{}, err error) {
	update := cp.getCallName(Name).Update
	timeoutChan := time.After(timeOutWait)
	for {
		select {
		case cn := <-update:
			ci := cn.get()
			if ci != nil {
				return ci.Param, ci.Result, nil
			}
		case <-timeoutChan:
			return nil, nil, ErrorTimeout
		}
	}
}

// Call - Execution of call
func (cp *CallPull) Call(Name string, Param map[string]interface{}, timeOutWait time.Duration) (interface{}, error) {
	cn := cp.getCallName(Name)
	ci := &callItem{
		Param:  Param,
		Result: make(chan interface{}, 1),
	}
	cn.add(Name, ci)
	select {
	case result := <-ci.Result:
		close(ci.Result)
		return result, nil
	case <-time.After(timeOutWait):
		// Close the chan to inform the Executive that the time is up and the result is no longer needed
		close(ci.Result)
		// if the result had inserted in the last momet then refund it
		if len(ci.Result) > 0 {
			return <-ci.Result, nil
		}
		return nil, ErrorTimeout
	}
}
