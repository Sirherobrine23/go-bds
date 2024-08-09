package exec

import "testing"

func TestOs(t *testing.T) {
	var osExec = new(Os)
	if err := osExec.Start(ProcExec{ Arguments: []string{"echo", "hello world"} }); err != nil {
		t.Fatal(err)
		return
	}

	if err := osExec.Wait(); err != nil {
		t.Fatal(err)
		return
	}
}
