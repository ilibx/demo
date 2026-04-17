package main

import (
	"fmt"
	"time"

	"taskmanager"
)

func main() {
	// 创建任务管理器
	tm := taskmanager.NewTaskManager()

	// 示例 1: Bash 脚本任务 - 每 5 秒执行一次
	bashTask := &taskmanager.Task{
		ID:       "bash-task-1",
		Name:     "Bash Hello World",
		Type:     taskmanager.TaskTypeBash,
		Script:   "echo 'Hello from Bash!' && date",
		Interval: 5 * time.Second,
	}

	if err := tm.AddTask(bashTask); err != nil {
		fmt.Printf("添加 Bash 任务失败：%v\n", err)
		return
	}

	// 示例 2: JavaScript 脚本任务 - 每 10 秒执行一次
	jsTask := &taskmanager.Task{
		ID:       "js-task-1",
		Name:     "JS Hello World",
		Type:     taskmanager.TaskTypeJS,
		Script: `
			console.log("Hello from JavaScript!");
			var now = new Date();
			console.log("Current time:", now.toLocaleString());
			var sum = 0;
			for (var i = 0; i < 10; i++) {
				sum += i;
			}
			console.log("Sum of 0-9:", sum);
		`,
		Interval: 10 * time.Second,
	}

	if err := tm.AddTask(jsTask); err != nil {
		fmt.Printf("添加 JS 任务失败：%v\n", err)
		return
	}

	// 示例 3: Python 脚本任务 - 每 15 秒执行一次
	pythonTask := &taskmanager.Task{
		ID:       "python-task-1",
		Name:     "Python Hello World",
		Type:     taskmanager.TaskTypePython,
		Script: `
import datetime
print("Hello from Python!")
print(f"Current time: {datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
numbers = list(range(10))
print(f"Sum of 0-9: {sum(numbers)}")
`,
		Interval: 15 * time.Second,
	}

	if err := tm.AddTask(pythonTask); err != nil {
		fmt.Printf("添加 Python 任务失败：%v\n", err)
		return
	}

	// 启动所有任务
	fmt.Println("启动任务...")
	tm.StartTask("bash-task-1")
	tm.StartTask("js-task-1")
	tm.StartTask("python-task-1")

	// 监听任务结果
	go func() {
		for result := range tm.GetResultChannel() {
			fmt.Printf("\n=== 任务执行结果 ===\n")
			fmt.Printf("任务 ID: %s\n", result.TaskID)
			fmt.Printf("成功：%v\n", result.Success)
			fmt.Printf("开始时间：%s\n", result.StartTime.Format(time.RFC3339))
			fmt.Printf("结束时间：%s\n", result.EndTime.Format(time.RFC3339))
			fmt.Printf("输出:\n%s\n", result.Output)
			if result.Error != nil {
				fmt.Printf("错误：%v\n", result.Error)
			}
			fmt.Printf("==================\n\n")
		}
	}()

	// 运行 60 秒后停止
	time.Sleep(60 * time.Second)

	fmt.Println("停止所有任务...")
	tm.StopTask("bash-task-1")
	tm.StopTask("js-task-1")
	tm.StopTask("python-task-1")

	// 等待一下让结果输出完成
	time.Sleep(2 * time.Second)

	// 关闭任务管理器
	tm.Shutdown()

	fmt.Println("任务管理器已关闭")
}
