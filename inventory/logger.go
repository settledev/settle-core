package inventory

import (
	"log"
	"os"
	"strings"
	"time"
)

type Logger struct {
	*log.Logger
	indentLevel int
	hostName    string
}

func NewLogger() *Logger {
	return &Logger{
		Logger:      log.New(os.Stdout, "", 0),
		indentLevel: 0,
	}
}

func (l *Logger) SetHost(hostName string) {
	l.hostName = hostName
}

func (l *Logger) indent() string {
	return strings.Repeat("  ", l.indentLevel)
}

func (l *Logger) Task(taskName string) {
	l.Printf("\n%s", strings.Repeat("=", 80))
	l.Printf("TASK [%s]", taskName)
	l.Printf("%s", strings.Repeat("=", 80))
}

func (l *Logger) HostTask(hostName, taskName string) {
	l.Printf("\n%s", strings.Repeat("-", 80))
	l.Printf("TASK [%s] ***********************************************************", taskName)
	l.Printf("host: %s", hostName)
	l.Printf("%s", strings.Repeat("-", 80))
}

func (l *Logger) HostSection(hostName string) {
	l.Printf("\n%s", strings.Repeat("=", 80))
	l.Printf("HOST: %s", hostName)
	l.Printf("%s", strings.Repeat("=", 80))
}

func (l *Logger) Info(message string) {
	l.Printf("%s[INFO] %s", l.indent(), message)
}

func (l *Logger) Success(message string) {
	l.Printf("%s[SUCCESS] %s", l.indent(), message)
}

func (l *Logger) Error(message string) {
	l.Printf("%s[ERROR] %s", l.indent(), message)
}

func (l *Logger) Warning(message string) {
	l.Printf("%s[WARNING] %s", l.indent(), message)
}

func (l *Logger) Debug(message string) {
	l.Printf("%s[DEBUG] %s", l.indent(), message)
}

func (l *Logger) Command(command string) {
	l.Printf("%s$ %s", l.indent(), command)
}

func (l *Logger) CommandOutput(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			l.Printf("%s%s", l.indent(), line)
		}
	}
}

func (l *Logger) PackageInstall(pkgName, manager string) {
	l.Printf("%s📦 Installing %s via %s...", l.indent(), pkgName, manager)
}

func (l *Logger) PackageRemove(pkgName, manager string) {
	l.Printf("%s📦 Removing %s via %s...", l.indent(), pkgName, manager)
}

func (l *Logger) PackageExists(pkgName, manager string) {
	l.Printf("%s📦 %s via %s already exists", l.indent(), pkgName, manager)
}

func (l *Logger) PackageSuccess(pkgName string, duration time.Duration) {
	l.Printf("%s✅ Successfully installed %s in %v", l.indent(), pkgName, duration)
}

func (l *Logger) PackageError(pkgName string, err error) {
	l.Printf("%s❌ Failed to install %s: %v", l.indent(), pkgName, err)
}

func (l *Logger) SSHConnection(host, user, port string) {
	l.Printf("%s🔌 Connecting to %s@%s:%s...", l.indent(), user, host, port)
}

func (l *Logger) SSHSuccess() {
	l.Printf("%s✅ SSH connection established", l.indent())
}

func (l *Logger) SSHError(err error) {
	l.Printf("%s❌ SSH connection failed: %v", l.indent(), err)
}

func (l *Logger) Summary(success, failed int) {
	l.Printf("\n%s", strings.Repeat("-", 80))
	l.Printf("SUMMARY")
	l.Printf("%s", strings.Repeat("-", 80))
	l.Printf("  Successful: %d", success)
	l.Printf("  Failed: %d", failed)
	l.Printf("%s", strings.Repeat("-", 80))
}

func (l *Logger) Indent() {
	l.indentLevel++
}

func (l *Logger) Unindent() {
	if l.indentLevel > 0 {
		l.indentLevel--
	}
}
