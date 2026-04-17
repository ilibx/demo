package taskmanager

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// parseCron 解析简单的 cron 表达式
// 支持格式：分 时 日 月 周
// 例如：*/5 * * * * 表示每 5 分钟
func parseCron(expr string) (time.Time, error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return time.Time{}, fmt.Errorf("invalid cron expression: expected 5 fields")
	}

	minute := parts[0]
	hour := parts[1]
	day := parts[2]
	month := parts[3]
	weekday := parts[4]

	now := time.Now()
	next := now

	// 简单的解析逻辑，生产环境建议使用更完善的 cron 库
	// 这里只处理 */n 和 * 的情况
	if minute == "*" || strings.HasPrefix(minute, "*/") {
		interval := 1
		if strings.HasPrefix(minute, "*/") {
			fmt.Sscanf(minute[2:], "%d", &interval)
		}
		// 计算下一个分钟
		next = next.Add(time.Duration(interval) * time.Minute)
		next = next.Truncate(time.Minute)
	}

	if hour != "*" && !strings.HasPrefix(hour, "*/") {
		var h int
		fmt.Sscanf(hour, "%d", &h)
		next = time.Date(next.Year(), next.Month(), next.Day(), h, next.Minute(), next.Second(), 0, next.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
	}

	if day != "*" && !strings.HasPrefix(day, "*/") {
		var d int
		fmt.Sscanf(day, "%d", &d)
		next = time.Date(next.Year(), next.Month(), d, next.Hour(), next.Minute(), next.Second(), 0, next.Location())
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
	}

	if month != "*" && !strings.HasPrefix(month, "*/") {
		var m int
		fmt.Sscanf(month, "%d", &m)
		next = time.Date(next.Year(), time.Month(m), next.Day(), next.Hour(), next.Minute(), next.Second(), 0, next.Location())
		if next.Before(now) {
			next = next.AddDate(1, 0, 0)
		}
	}

	if weekday != "*" && !strings.HasPrefix(weekday, "*/") {
		var w int
		fmt.Sscanf(weekday, "%d", &w)
		for next.Weekday() != time.Weekday(w) {
			next = next.AddDate(0, 0, 1)
		}
	}

	return next, nil
}

// createTempFile 创建临时文件
func createTempFile(content, ext string) (string, error) {
	tmpFile, err := os.CreateTemp("", "task_*"+ext)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(content)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// deleteTempFile 删除临时文件
func deleteTempFile(path string) {
	os.Remove(path)
}
