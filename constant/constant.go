package constant

var LanguageName = []string{"C", "C++", "Pascal", "Java", "Ruby", "Bash", "Python", "PHP", "Perl", "C#", "Obj-C", "FreeBasic", "Scheme", "Clang", "Clang++", "Lua", "JavaScript", "Go", "Other Language"}
var LanguageExt = []string{"c", "cc", "pas", "java", "rb", "sh", "py", "php", "pl", "cs", "m", "bas", "scm", "c", "cc", "lua", "js", "go"}
var JudgeResult = []string{
	"等待",
	"等待重判",
	"编译中",
	"运行并评判",
	"正确",
	"格式错误",
	"答案错误",
	"时间超限",
	"内存超限",
	"输出超限",
	"运行错误",
	"编译错误",
	"编译成功",
	"测试运行",
}

const WAIT_TO_JUDGE = 0
const WAIT_TO_REJUDGE = 1
const COMPILING = 2
const RUNNING = 3
const ACCEPT = 4
const LINT_ERROR = 5
const WRONG_ANSWER = 6
const TIME_LIMIT_EXCEEDED = 7
const MEMORY_LIMIT_EXCEEDED = 8
const OUTPUT_LIMIT_EXCEEDED = 9
const RUNTIME_ERROR = 10
const COMPILE_ERROR = 11
const COMPILE_SUCCESSS = 12
const TEST_RUN = 13

const PROBLEM_NORMAL = 0
const PROBLEM_AC = 1
const PROBLEM_WA = 2

const CONTEST_NOT_START = 1
const CONTEST_RUNNING = 2
const CONTEST_FINISH = 3
