package qwen

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Executor qwen-code CLI 执行器
type Executor struct {
	workDir    string
	timeout    time.Duration
	maxOutput  int
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	WorkDir   string        // 工作目录
	Timeout   time.Duration // 命令超时时间
	MaxOutput int           // 最大输出行数
}

// DefaultConfig 默认配置
func DefaultConfig() ExecutorConfig {
	return ExecutorConfig{
		WorkDir:   "/tmp/qwen-workspace",
		Timeout:   5 * time.Minute,
		MaxOutput: 500,
	}
}

// NewExecutor 创建执行器
func NewExecutor(config ExecutorConfig) *Executor {
	return &Executor{
		workDir:   config.WorkDir,
		timeout:   config.Timeout,
		maxOutput: config.MaxOutput,
	}
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success   bool
	Output    string
	Error     string
	ExitCode  int
	Duration  time.Duration
}

// Execute 执行 qwen-code 命令
// prompt: 用户输入的提示词/命令
func (e *Executor) Execute(ctx context.Context, prompt string) (*ExecutionResult, error) {
	startTime := time.Now()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 构建命令：qwen --headless -p "<prompt>"
	// 使用 headless 模式进行自动化执行
	cmd := exec.CommandContext(ctx, "qwen", "--headless", "-p", prompt)
	cmd.Dir = e.workDir

	// 创建管道获取输出
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe failed: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe failed: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start command failed: %w", err)
	}

	// 并发读取输出
	var outputBuilder strings.Builder
	var errorBuilder strings.Builder
	var wg sync.WaitGroup

	wg.Add(2)
	go e.readOutput(stdoutPipe, &outputBuilder, &wg)
	go e.readOutput(stderrPipe, &errorBuilder, &wg)

	// 等待命令完成
	err = cmd.Wait()
	wg.Wait()

	duration := time.Since(startTime)

	// 限制输出长度
	output := truncateOutput(outputBuilder.String(), e.maxOutput)
	errorOutput := truncateOutput(errorBuilder.String(), e.maxOutput)

	result := &ExecutionResult{
		Success:  err == nil,
		Output:   output,
		Error:    errorOutput,
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	log.Printf("Qwen-code execution completed: success=%v, duration=%v", 
		result.Success, duration)

	return result, nil
}

// readOutput 读取输出流
func (e *Executor) readOutput(pipe io.ReadCloser, builder *strings.Builder, wg *sync.WaitGroup) {
	defer wg.Done()
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	lineCount := 0
	for scanner.Scan() {
		if lineCount >= e.maxOutput {
			break
		}
		builder.WriteString(scanner.Text())
		builder.WriteString("\n")
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Read output error: %v", err)
	}
}

// truncateOutput 截断输出
func truncateOutput(output string, maxLines int) string {
	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}

	truncated := strings.Join(lines[:maxLines], "\n")
	return fmt.Sprintf("%s\n... [truncated, total %d lines]", truncated, len(lines))
}

// ExecuteWithStream 执行并返回流式输出
// 用于需要实时反馈的场景
func (e *Executor) ExecuteWithStream(ctx context.Context, prompt string, outputChan chan<- string) (*ExecutionResult, error) {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "qwen-code", "--headless", "-p", prompt)
	cmd.Dir = e.workDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe failed: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start command failed: %w", err)
	}

	var wg sync.WaitGroup
	var outputBuilder strings.Builder
	var errorBuilder strings.Builder

	wg.Add(2)
	go e.streamOutput(stdoutPipe, &outputBuilder, outputChan, &wg)
	go e.streamOutput(stderrPipe, &errorBuilder, outputChan, &wg)

	err = cmd.Wait()
	wg.Wait()

	duration := time.Since(startTime)

	result := &ExecutionResult{
		Success:  err == nil,
		Output:   outputBuilder.String(),
		Error:    errorBuilder.String(),
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Success = false
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	return result, nil
}

// streamOutput 流式输出
func (e *Executor) streamOutput(pipe io.ReadCloser, builder *strings.Builder, outputChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	lineCount := 0
	for scanner.Scan() {
		if lineCount >= e.maxOutput {
			break
		}
		line := scanner.Text()
		builder.WriteString(line + "\n")
		
		// 发送到输出通道
		select {
		case outputChan <- line:
		default:
			// 通道已满，跳过
		}
		lineCount++
	}
}

// GetVersion 获取 qwen 版本
func (e *Executor) GetVersion() (string, error) {
	cmd := exec.Command("qwen", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get version failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CheckInstalled 检查 qwen 是否已安装
func CheckInstalled() bool {
	_, err := exec.LookPath("qwen")
	return err == nil
}
