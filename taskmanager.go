package taskmanager

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeBash   TaskType = "bash"
	TaskTypeJS     TaskType = "js"
	TaskTypePython TaskType = "python"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusStopped   TaskStatus = "stopped"
)

// Task 任务定义
type Task struct {
	ID          string
	Name        string
	Type        TaskType
	Script      string
	CronExpr    string // 简单的 cron 表达式，如 "*/5 * * * *" 表示每 5 分钟
	Interval    time.Duration
	Status      TaskStatus
	LastRun     time.Time
	NextRun     time.Time
	Description string
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID     string
	Success    bool
	Output     string
	Error      error
	StartTime  time.Time
	EndTime    time.Time
}

// ScheduledTask 定时任务
type ScheduledTask struct {
	Task       *Task
	stopChan   chan struct{}
	isRunning  bool
	ctx        context.Context
	cancel     context.CancelFunc
}

// TaskManager 任务管理器
type TaskManager struct {
	tasks      map[string]*ScheduledTask
	mu         sync.RWMutex
	resultChan chan *TaskResult
	wg         sync.WaitGroup
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks:      make(map[string]*ScheduledTask),
		resultChan: make(chan *TaskResult, 100),
	}
}

// AddTask 添加任务
func (tm *TaskManager) AddTask(task *Task) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	st := &ScheduledTask{
		Task:     task,
		stopChan: make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}

	tm.tasks[task.ID] = st
	task.Status = StatusPending

	// 计算下次运行时间
	if task.Interval > 0 {
		task.NextRun = time.Now().Add(task.Interval)
	} else if task.CronExpr != "" {
		next, err := parseCron(task.CronExpr)
		if err != nil {
			delete(tm.tasks, task.ID)
			return fmt.Errorf("invalid cron expression: %v", err)
		}
		task.NextRun = next
	}

	return nil
}

// StartTask 启动任务
func (tm *TaskManager) StartTask(taskID string) error {
	tm.mu.Lock()
	st, exists := tm.tasks[taskID]
	tm.mu.Unlock()

	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	if st.isRunning {
		return fmt.Errorf("task is already running")
	}

	st.isRunning = true
	tm.wg.Add(1)

	go tm.runScheduler(st)
	return nil
}

// StopTask 停止任务
func (tm *TaskManager) StopTask(taskID string) error {
	tm.mu.Lock()
	st, exists := tm.tasks[taskID]
	tm.mu.Unlock()

	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	if !st.isRunning {
		return fmt.Errorf("task is not running")
	}

	st.cancel()
	close(st.stopChan)
	st.isRunning = false
	st.Task.Status = StatusStopped

	return nil
}

// RemoveTask 移除任务
func (tm *TaskManager) RemoveTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	st, exists := tm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	if st.isRunning {
		st.cancel()
		close(st.stopChan)
	}

	delete(tm.tasks, taskID)
	return nil
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	st, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	return st.Task, nil
}

// ListTasks 列出所有任务
func (tm *TaskManager) ListTasks() []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*Task, 0, len(tm.tasks))
	for _, st := range tm.tasks {
		tasks = append(tasks, st.Task)
	}

	return tasks
}

// GetResultChannel 获取结果通道
func (tm *TaskManager) GetResultChannel() <-chan *TaskResult {
	return tm.resultChan
}

// Shutdown 关闭任务管理器
func (tm *TaskManager) Shutdown() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, st := range tm.tasks {
		if st.isRunning {
			st.cancel()
			close(st.stopChan)
		}
	}

	tm.wg.Wait()
	close(tm.resultChan)
}

// runScheduler 运行任务调度器
func (tm *TaskManager) runScheduler(st *ScheduledTask) {
	defer tm.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-st.ctx.Done():
			return
		case <-ticker.C:
			if time.Now().After(st.Task.NextRun) {
				go tm.executeTask(st)
				
				// 更新下次运行时间
				if st.Task.Interval > 0 {
					st.Task.NextRun = time.Now().Add(st.Task.Interval)
				} else if st.Task.CronExpr != "" {
					next, err := parseCron(st.Task.CronExpr)
					if err == nil {
						st.Task.NextRun = next
					}
				}
				st.Task.LastRun = time.Now()
			}
		}
	}
}

// executeTask 执行任务
func (tm *TaskManager) executeTask(st *ScheduledTask) {
	task := st.Task
	task.Status = StatusRunning

	result := &TaskResult{
		TaskID:    task.ID,
		StartTime: time.Now(),
	}

	var output string
	var err error

	switch task.Type {
	case TaskTypeBash:
		output, err = executeBash(task.Script)
	case TaskTypeJS:
		output, err = executeJS(task.Script)
	case TaskTypePython:
		output, err = executePython(task.Script)
	default:
		err = fmt.Errorf("unsupported task type: %s", task.Type)
	}

	result.EndTime = time.Now()
	result.Output = output
	result.Error = err
	result.Success = err == nil

	if err != nil {
		task.Status = StatusFailed
	} else {
		task.Status = StatusCompleted
	}

	tm.resultChan <- result
}

// executeBash 执行 Bash 脚本
func executeBash(script string) (string, error) {
	cmd := exec.Command("bash", "-c", script)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// executeJS 执行 JavaScript 脚本（使用 goja）
func executeJS(script string) (string, error) {
	vm := goja.New()

	// 捕获 console.log 输出
	var output string
	vm.Set("print", func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			output += arg.String() + "\n"
		}
		return goja.Undefined()
	})

	// 设置 console 对象
	console := vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		for _, arg := range call.Arguments {
			output += arg.String() + "\n"
		}
		return goja.Undefined()
	})
	vm.Set("console", console)

	_, err := vm.RunString(script)
	if err != nil {
		return output, err
	}

	return output, nil
}

// executePython 执行 Python 脚本
func executePython(script string) (string, error) {
	// 创建临时文件
	tmpFile, err := createTempFile(script, ".py")
	if err != nil {
		return "", err
	}
	defer deleteTempFile(tmpFile)

	cmd := exec.Command("python3", tmpFile)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
