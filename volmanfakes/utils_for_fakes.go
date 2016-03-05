package volmanfakes

import "io"

func CmdStdoutPipeReturnsInOrder(fakeCmd *FakeCmd, inOrderResponses []io.ReadCloser) {
	calls := 0
	fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
		defer func() { calls++ }()
		return inOrderResponses[calls], nil
	}
}
