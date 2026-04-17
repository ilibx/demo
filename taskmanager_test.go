package taskmanager

import (
	"testing"
	"time"
)

func TestNewTaskManager(t *testing.T) {
	tm := NewTaskManager()
	if tm == nil {
		t.Fatal("NewTaskManager returned nil")
	}
	if tm.tasks == nil {
		t.Fatal("tasks map not initialized")
	}
	if tm.resultChan == nil {
		t.Fatal("resultChan not initialized")
	}
}

func TestAddTask(t *testing.T) {
	tm := NewTaskManager()

	task := &Task{
		ID:       "test-1",
		Name:     "Test Task",
		Type:     TaskTypeBash,
		Script:   "echo hello",
		Interval: 5 * time.Second,
	}

	err := tm.AddTask(task)
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}

	if len(tm.tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tm.tasks))
	}

	if task.Status != StatusPending {
		t.Fatalf("Expected status pending, got %s", task.Status)
	}

	// Test duplicate task
	err = tm.AddTask(task)
	if err == nil {
		t.Fatal("Expected error for duplicate task, got nil")
	}
}

func TestGetTask(t *testing.T) {
	tm := NewTaskManager()

	task := &Task{
		ID:       "test-2",
		Name:     "Test Task 2",
		Type:     TaskTypeJS,
		Script:   "console.log('hello')",
		Interval: 10 * time.Second,
	}

	tm.AddTask(task)

	retrieved, err := tm.GetTask("test-2")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if retrieved.Name != "Test Task 2" {
		t.Fatalf("Expected name 'Test Task 2', got %s", retrieved.Name)
	}

	// Test non-existent task
	_, err = tm.GetTask("non-existent")
	if err == nil {
		t.Fatal("Expected error for non-existent task, got nil")
	}
}

func TestListTasks(t *testing.T) {
	tm := NewTaskManager()

	tasks := []*Task{
		{ID: "list-1", Name: "Task 1", Type: TaskTypeBash, Script: "echo 1", Interval: 5 * time.Second},
		{ID: "list-2", Name: "Task 2", Type: TaskTypeJS, Script: "console.log(1)", Interval: 10 * time.Second},
		{ID: "list-3", Name: "Task 3", Type: TaskTypePython, Script: "print(1)", Interval: 15 * time.Second},
	}

	for _, task := range tasks {
		tm.AddTask(task)
	}

	list := tm.ListTasks()
	if len(list) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(list))
	}
}

func TestRemoveTask(t *testing.T) {
	tm := NewTaskManager()

	task := &Task{
		ID:       "remove-1",
		Name:     "To Remove",
		Type:     TaskTypeBash,
		Script:   "echo hello",
		Interval: 5 * time.Second,
	}

	tm.AddTask(task)

	err := tm.RemoveTask("remove-1")
	if err != nil {
		t.Fatalf("RemoveTask failed: %v", err)
	}

	if len(tm.tasks) != 0 {
		t.Fatalf("Expected 0 tasks after removal, got %d", len(tm.tasks))
	}

	// Test removing non-existent task
	err = tm.RemoveTask("non-existent")
	if err == nil {
		t.Fatal("Expected error for non-existent task, got nil")
	}
}

func TestExecuteBash(t *testing.T) {
	output, err := executeBash("echo 'Hello from Bash'")
	if err != nil {
		t.Fatalf("executeBash failed: %v", err)
	}

	expected := "Hello from Bash\n"
	if output != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, output)
	}
}

func TestExecuteJS(t *testing.T) {
	script := `
		console.log("Hello from JS");
		var x = 10;
		var y = 20;
		console.log("Sum:", x + y);
	`

	output, err := executeJS(script)
	if err != nil {
		t.Fatalf("executeJS failed: %v", err)
	}

	if output == "" {
		t.Fatal("Expected non-empty output from JS execution")
	}
}

func TestExecuteJSWithPrint(t *testing.T) {
	script := `
		print("Using print function");
		var result = 5 * 5;
		print("Result:", result);
	`

	output, err := executeJS(script)
	if err != nil {
		t.Fatalf("executeJS failed: %v", err)
	}

	if output == "" {
		t.Fatal("Expected non-empty output from JS execution")
	}
}

func TestExecuteJSError(t *testing.T) {
	script := `
		var x = undefinedVariable;
	`

	_, err := executeJS(script)
	if err == nil {
		t.Fatal("Expected error for invalid JS, got nil")
	}
}

func TestParseCron(t *testing.T) {
	tests := []struct {
		expr string
		want bool // true if we expect success
	}{
		{"*/5 * * * *", true},
		{"* * * * *", true},
		{"0 * * * *", true},
		{"invalid", false},
		{"* * *", false},
	}

	for _, tt := range tests {
		next, err := parseCron(tt.expr)
		if tt.want && err != nil {
			t.Errorf("parseCron(%q) returned error: %v", tt.expr, err)
		}
		if !tt.want && err == nil {
			t.Errorf("parseCron(%q) expected error, got nil (next=%v)", tt.expr, next)
		}
	}
}

func TestStartAndStopTask(t *testing.T) {
	tm := NewTaskManager()

	task := &Task{
		ID:       "startstop-1",
		Name:     "Start Stop Test",
		Type:     TaskTypeBash,
		Script:   "echo test",
		Interval: 1 * time.Second,
	}

	tm.AddTask(task)

	err := tm.StartTask("startstop-1")
	if err != nil {
		t.Fatalf("StartTask failed: %v", err)
	}

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	err = tm.StopTask("startstop-1")
	if err != nil {
		t.Fatalf("StopTask failed: %v", err)
	}

	if task.Status != StatusStopped {
		t.Fatalf("Expected status stopped, got %s", task.Status)
	}
}

func TestTaskResultChannel(t *testing.T) {
	tm := NewTaskManager()

	task := &Task{
		ID:       "result-1",
		Name:     "Result Test",
		Type:     TaskTypeBash,
		Script:   "echo 'test output'",
		Interval: 1 * time.Second,
	}

	tm.AddTask(task)
	tm.StartTask("result-1")

	// Wait for task to execute and produce result
	select {
	case result := <-tm.GetResultChannel():
		if result.TaskID != "result-1" {
			t.Fatalf("Expected task ID 'result-1', got %s", result.TaskID)
		}
		if !result.Success {
			t.Fatal("Expected successful execution")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for task result")
	}

	tm.StopTask("result-1")
	tm.Shutdown()
}
