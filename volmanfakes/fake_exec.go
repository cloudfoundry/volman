// This file was generated by counterfeiter
package volmanfakes

import (
	"sync"

	"github.com/cloudfoundry-incubator/volman/system"
)

type FakeExec struct {
	CommandStub        func(name string, arg ...string) system.Cmd
	commandMutex       sync.RWMutex
	commandArgsForCall []struct {
		name string
		arg  []string
	}
	commandReturns struct {
		result1 system.Cmd
	}
}

func (fake *FakeExec) Command(name string, arg ...string) system.Cmd {
	fake.commandMutex.Lock()
	fake.commandArgsForCall = append(fake.commandArgsForCall, struct {
		name string
		arg  []string
	}{name, arg})
	fake.commandMutex.Unlock()
	if fake.CommandStub != nil {
		return fake.CommandStub(name, arg...)
	} else {
		return fake.commandReturns.result1
	}
}

func (fake *FakeExec) CommandCallCount() int {
	fake.commandMutex.RLock()
	defer fake.commandMutex.RUnlock()
	return len(fake.commandArgsForCall)
}

func (fake *FakeExec) CommandArgsForCall(i int) (string, []string) {
	fake.commandMutex.RLock()
	defer fake.commandMutex.RUnlock()
	return fake.commandArgsForCall[i].name, fake.commandArgsForCall[i].arg
}

func (fake *FakeExec) CommandReturns(result1 system.Cmd) {
	fake.CommandStub = nil
	fake.commandReturns = struct {
		result1 system.Cmd
	}{result1}
}

var _ system.Exec = new(FakeExec)
