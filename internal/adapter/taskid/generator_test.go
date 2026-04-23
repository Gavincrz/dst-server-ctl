package taskid

import "testing"

func TestGeneratorCreatesNonEmptyDistinctTaskIDs(t *testing.T) {
	generator := Generator{}

	first, err := generator.NewTaskID()
	if err != nil {
		t.Fatalf("NewTaskID() error = %v", err)
	}
	second, err := generator.NewTaskID()
	if err != nil {
		t.Fatalf("second NewTaskID() error = %v", err)
	}

	if first == "" {
		t.Fatal("first task ID is empty")
	}
	if second == "" {
		t.Fatal("second task ID is empty")
	}
	if first == second {
		t.Fatalf("task IDs should be distinct, both were %q", first)
	}
}
