package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ANSI 颜色常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorDim    = "\033[2m"
)

const cacheFileName = ".git-commit.json"

// 执行 git 命令并返回结果
func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// 验证工单号格式 (必须 bcds-<数字> 或 bcds-<数字>-xxx)
func validateWorkItem(workItem string) bool {
	re := regexp.MustCompile(`^bcds-\d+(-[a-z0-9]+)*$`)
	return re.MatchString(workItem)
}

// CommitType 提交类型结构
type CommitType struct {
	code        string
	description string
}

// LastInputs 用于记忆上次输入
type LastInputs struct {
	WorkItem        string `json:"workItem"`
	CommitTypeIndex int    `json:"commitTypeIndex"`
	Scope           string `json:"scope"`
	Description     string `json:"description"`
	IssueRef        string `json:"issueRef"`
}

// 获取提交类型列表（有序）
func getCommitTypes() []CommitType {
	return []CommitType{
		{"feat", "新功能 (A new feature)"},
		{"fix", "Bug修复 (A bug fix)"},
		{"docs", "文档更新 (Documentation only changes)"},
		{"style", "代码格式 (Changes that do not affect code meaning)"},
		{"refactor", "代码重构 (Neither fixes bug nor adds feature)"},
		{"perf", "性能优化 (A code change that improves performance)"},
		{"test", "测试相关 (Adding or correcting tests)"},
		{"build", "构建相关 (Affect build system or dependencies)"},
		{"ci", "CI配置 (Changes to CI configuration files)"},
		{"chore", "其他杂项 (Other changes)"},
	}
}

// 验证简短描述：≤50 个字符，首字母小写，结尾不加句号
func validateShortDescription(desc string) bool {
	desc = strings.TrimSpace(desc)
	if len(desc) == 0 || len(desc) > 50 {
		return false
	}

	runes := []rune(desc)
	if !unicode.IsLower(runes[0]) {
		return false
	}
	if strings.HasSuffix(desc, ".") {
		return false
	}
	return true
}

// 读取确认输入
func readYesNo(prompt string) bool {
	for {
		answer := strings.ToLower(readLine(prompt + " (y/N): "))
		if answer == "y" || answer == "yes" {
			return true
		}
		if answer == "" || answer == "n" || answer == "no" {
			return false
		}
		fmt.Println(ColorRed + "请输入 y 或 n。" + ColorReset)
	}
}

// 从控制台读取一行输入
func readLine(prompt string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

// 从控制台读取多行输入，空行结束
func readMultiline(prompt string) string {
	fmt.Println(prompt + " (直接回车结束):")
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for {
		scanner.Scan()
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// 带默认值的单行输入（默认值使用虚影展示）
func readLineWithDefault(prompt string, defaultVal string) string {
	display := prompt
	if defaultVal != "" {
		display = fmt.Sprintf("%s%s%s%s ", prompt, ColorDim, defaultVal, ColorReset)
	}
	input := readLine(display)
	if strings.TrimSpace(input) == "" {
		return defaultVal
	}
	return input
}

// 获取当前分支名
func getCurrentBranch() string {
	out, err := runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func extractWorkItemFromBranch(branch string) string {
	lower := strings.ToLower(branch)
	re := regexp.MustCompile(`bcds-\d+(?:-[a-z0-9]+)*`)
	match := re.FindString(lower)
	if match != "" {
		return match
	}
	return lower
}

// 获取缓存文件路径
func getCacheFilePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return cacheFileName
	}
	guidePath := filepath.Join(cwd, ".git-guide")
	_ = os.MkdirAll(guidePath, 0755)
	return filepath.Join(cwd, ".git-guide", cacheFileName)
}

// 读取上次输入
func loadLastInputs() LastInputs {
	data, err := os.ReadFile(getCacheFilePath())
	if err != nil {
		return LastInputs{}
	}
	var last LastInputs
	if err := json.Unmarshal(data, &last); err != nil {
		return LastInputs{}
	}
	return last
}

// 保存本次输入
func saveLastInputs(inputs LastInputs) {
	data, err := json.MarshalIndent(inputs, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(getCacheFilePath(), data, 0o644)
}

// 程序入口
func main() {
	// 检查是否有暂存的更改
	_, err := runGitCommand("diff", "--cached", "--quiet")
	if err == nil {
		fmt.Println(ColorRed + "错误：没有暂存的更改，请先使用 'git add' 添加文件。" + ColorReset)
		os.Exit(1)
	}

	fmt.Println(ColorCyan + "=== Git 提交助手 (Custom Commit CLI) ===" + ColorReset)

	lastInputs := loadLastInputs()
	branchSuggestion := ""
	if branch := getCurrentBranch(); branch != "" {
		branchSuggestion = extractWorkItemFromBranch(branch)
	}

	// 工单号
	var workItem string
	for {
		defaultWorkItem := lastInputs.WorkItem
		if defaultWorkItem == "" {
			defaultWorkItem = branchSuggestion
		}
		workItem = strings.ToLower(readLineWithDefault(ColorYellow+"1. 请输入工单号 (格式: bcds-<数字> 或 bcds-<数字>-xxx): "+ColorReset, defaultWorkItem))
		if validateWorkItem(workItem) {
			break
		}
		fmt.Println(ColorRed + "工单号格式无效，请重新输入。" + ColorReset)
	}

	// 选择提交类型
	commitTypes := getCommitTypes()
	fmt.Println("\n" + ColorYellow + "2. 请选择提交类型:" + ColorReset)
	for i, ct := range commitTypes {
		fmt.Printf("   %d. %-8s %s\n", i+1, ct.code, ct.description)
	}

	var commitType string
	defaultIndex := lastInputs.CommitTypeIndex
	if defaultIndex < 1 || defaultIndex > len(commitTypes) {
		defaultIndex = 0
	}
	defaultIndexStr := ""
	if defaultIndex > 0 {
		defaultIndexStr = strconv.Itoa(defaultIndex)
	}
	for {
		input := readLineWithDefault(fmt.Sprintf("   请输入编号 (1-%d): ", len(commitTypes)), defaultIndexStr)
		choice, err := strconv.Atoi(input)
		if err == nil && choice >= 1 && choice <= len(commitTypes) {
			commitType = commitTypes[choice-1].code
			defaultIndex = choice
			break
		}
		fmt.Printf(ColorRed+"请输入有效的数字 (1-%d)\n"+ColorReset, len(commitTypes))
	}

	// 输入范围 scope（可选）
	scope := strings.TrimSpace(readLineWithDefault("\n"+ColorYellow+"3. 请输入影响范围（可选，如 api、payment，回车跳过）: "+ColorReset, lastInputs.Scope))

	// 输入并验证简短描述
	fmt.Println("\n" + ColorYellow + "4. 请输入简短描述（≤50字符，首字母小写，无句号）:" + ColorReset)
	var description string
	for {
		description = readLineWithDefault("   描述: ", lastInputs.Description)
		if validateShortDescription(description) {
			break
		}
		fmt.Println(ColorRed + "描述不符合规范，请重试。" + ColorReset)
	}

	// 可选正文
	var body string
	if readYesNo("\n需要添加详细正文吗？") {
		body = readMultiline("请输入正文，每行 ≤72 字符")
	}

	// 破坏性变更说明
	var breakingChange string
	if readYesNo("\n是否存在破坏性变更（Breaking Change）？") {
		breakingChange = readMultiline("描述 Breaking Change，空行结束")
	}

	// 关联 Issue（可选）
	issueRef := strings.TrimSpace(readLineWithDefault("\n需要关联 Issue 吗？直接输入（例如 Closes #123，回车跳过）: ", lastInputs.IssueRef))

	// 构建提交信息
	header := commitType
	if scope != "" {
		header = fmt.Sprintf("%s(%s)", header, scope)
	}
	header = fmt.Sprintf("%s: %s", header, strings.TrimSpace(description))

	var sections []string
	sections = append(sections, header)
	if body != "" {
		sections = append(sections, body)
	}

	var footers []string
	if breakingChange != "" {
		footers = append(footers, fmt.Sprintf("BREAKING CHANGE: %s", breakingChange))
	}
	if issueRef != "" {
		footers = append(footers, issueRef)
	}
	if workItem != "" {
		footers = append(footers, fmt.Sprintf("Refs %s", workItem))
	}
	if len(footers) > 0 {
		sections = append(sections, strings.Join(footers, "\n"))
	}

	commitMessage := strings.Join(sections, "\n\n")

	saveLastInputs(LastInputs{
		WorkItem:        workItem,
		CommitTypeIndex: defaultIndex,
		Scope:           scope,
		Description:     description,
		IssueRef:        issueRef,
	})

	fmt.Println("\n" + ColorGreen + "生成的提交信息:" + ColorReset)
	fmt.Println("--------------------------------------------------")
	fmt.Println(commitMessage)
	fmt.Println("--------------------------------------------------")

	// 确认提交
	confirm := strings.ToLower(readLine("是否确认提交? (Y/N): "))
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println(ColorRed + "提交已取消。" + ColorReset)
		os.Exit(0)
	}

	// 执行提交
	output, err := runGitCommand("commit", "-m", commitMessage)
	if err != nil {
		fmt.Printf(ColorRed+"提交失败: %s\n"+ColorReset, output)
		os.Exit(1)
	}

	// 提交成功
	fmt.Println("\n" + ColorGreen + "=== 提交成功 ===" + ColorReset)
	fmt.Printf("提交内容:\n%s\n", commitMessage)
	fmt.Println("可以使用 'git push' 推送更改。")
}
