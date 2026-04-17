# Golang 定时任务管理器

纯 Golang 实现的定时任务管理系统，支持执行 Bash、JavaScript 和 Python 脚本任务。

## 特性

- ✅ **Bash 脚本支持**: 直接执行系统 Bash 命令
- ✅ **JavaScript 支持**: 使用 [goja](https://github.com/dop251/goja) 引擎执行 JS 脚本
- ✅ **Python 脚本支持**: 通过调用系统 Python 解释器执行 Python 脚本
- ✅ **灵活的调度**: 支持固定间隔和 Cron 表达式
- ✅ **任务管理**: 添加、启动、停止、删除任务
- ✅ **结果追踪**: 获取任务执行结果和状态
- ✅ **并发安全**: 使用 mutex 保证线程安全

## 项目结构

```
taskmanager/
├── taskmanager.go        # 核心任务管理器实现
├── cron.go               # Cron 表达式解析工具
├── taskmanager_test.go   # 单元测试
├── cmd/
│   └── taskmanager-demo/
│       └── main.go       # 示例程序
├── go.mod                # Go 模块定义
└── README.md             # 说明文档
```

## 安装依赖

```bash
go mod tidy
```

## 快速开始

### 1. 创建任务管理器

```go
import "taskmanager"

tm := taskmanager.NewTaskManager()
```

### 2. 添加 Bash 任务

```go
bashTask := &taskmanager.Task{
    ID:       "bash-task-1",
    Name:     "系统监控",
    Type:     taskmanager.TaskTypeBash,
    Script:   "echo '当前时间:' && date && uptime",
    Interval: 30 * time.Second, // 每 30 秒执行一次
}

tm.AddTask(bashTask)
tm.StartTask("bash-task-1")
```

### 3. 添加 JavaScript 任务

```go
jsTask := &taskmanager.Task{
    ID:       "js-task-1",
    Name:     "数据计算",
    Type:     taskmanager.TaskTypeJS,
    Script: `
        console.log("执行数据计算...");
        var data = [1, 2, 3, 4, 5];
        var sum = data.reduce((a, b) => a + b, 0);
        console.log("总和:", sum);
    `,
    Interval: 60 * time.Second, // 每分钟执行一次
}

tm.AddTask(jsTask)
tm.StartTask("js-task-1")
```

### 4. 添加 Python 任务

```go
pythonTask := &taskmanager.Task{
    ID:       "python-task-1",
    Name:     "数据分析",
    Type:     taskmanager.TaskTypePython,
    Script: `
import datetime
import json

print("执行 Python 数据分析...")
now = datetime.datetime.now()
print(f"当前时间：{now}")

# 示例数据处理
data = {"values": [1, 2, 3, 4, 5]}
result = sum(data["values"])
print(f"计算结果：{result}")
`,
    Interval: 120 * time.Second, // 每 2 分钟执行一次
}

tm.AddTask(pythonTask)
tm.StartTask("python-task-1")
```

### 5. 使用 Cron 表达式

```go
task := &taskmanager.Task{
    ID:       "cron-task-1",
    Name:     "每小时任务",
    Type:     taskmanager.TaskTypeBash,
    Script:   "echo '每小时执行一次'",
    CronExpr: "0 * * * *", // 每小时整点执行
}

tm.AddTask(task)
tm.StartTask("cron-task-1")
```

### 6. 监听任务结果

```go
go func() {
    for result := range tm.GetResultChannel() {
        fmt.Printf("任务 %s 执行完成\n", result.TaskID)
        fmt.Printf("成功：%v\n", result.Success)
        fmt.Printf("输出：%s\n", result.Output)
        if result.Error != nil {
            fmt.Printf("错误：%v\n", result.Error)
        }
    }
}()
```

### 7. 任务控制

```go
// 获取任务信息
task, err := tm.GetTask("task-id")

// 列出所有任务
tasks := tm.ListTasks()

// 停止任务
tm.StopTask("task-id")

// 删除任务
tm.RemoveTask("task-id")

// 关闭任务管理器
tm.Shutdown()
```

## API 参考

### Task 结构

```go
type Task struct {
    ID          string        // 任务唯一标识
    Name        string        // 任务名称
    Type        TaskType      // 任务类型 (bash/js/python)
    Script      string        // 要执行的脚本内容
    CronExpr    string        // Cron 表达式 (如 "*/5 * * * *")
    Interval    time.Duration // 执行间隔
    Status      TaskStatus    // 任务状态
    LastRun     time.Time     // 上次运行时间
    NextRun     time.Time     // 下次运行时间
    Description string        // 任务描述
}
```

### TaskType 常量

```go
TaskTypeBash   // Bash 脚本
TaskTypeJS     // JavaScript 脚本
TaskTypePython // Python 脚本
```

### TaskStatus 常量

```go
StatusPending   // 等待中
StatusRunning   // 运行中
StatusCompleted // 已完成
StatusFailed    // 失败
StatusStopped   // 已停止
```

### TaskManager 方法

| 方法 | 说明 |
|------|------|
| `NewTaskManager()` | 创建新的任务管理器 |
| `AddTask(task *Task)` | 添加任务 |
| `StartTask(taskID string)` | 启动任务 |
| `StopTask(taskID string)` | 停止任务 |
| `RemoveTask(taskID string)` | 删除任务 |
| `GetTask(taskID string)` | 获取任务信息 |
| `ListTasks()` | 列出所有任务 |
| `GetResultChannel()` | 获取结果通道 |
| `Shutdown()` | 关闭任务管理器 |

## 运行示例

```bash
# 编译并运行示例程序
go run ./cmd/taskmanager-demo/main.go
```

示例程序会创建三个定时任务（Bash、JS、Python），运行 60 秒后自动停止。

## 运行测试

```bash
go test -v ./...
```

## Cron 表达式格式

支持标准的 5 字段 Cron 格式：

```
分 时 日 月 周
```

示例：
- `*/5 * * * *` - 每 5 分钟
- `0 * * * *` - 每小时整点
- `0 0 * * *` - 每天午夜
- `0 0 * * 1` - 每周一午夜
- `* * * * *` - 每分钟

## 注意事项

1. **Python 依赖**: 执行 Python 脚本需要系统安装 Python 3
2. **Cron 解析**: 当前实现支持基本的 Cron 语法，复杂场景建议使用专业 cron 库
3. **资源管理**: 使用完毕后请调用 `Shutdown()` 释放资源
4. **错误处理**: 建议始终检查任务执行结果中的错误信息

## 许可证

MIT License
