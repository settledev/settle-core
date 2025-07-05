package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/inventory"
	"github.com/settlectl/settle-core/inventory/ssh"
)

type AptManager struct {
	SSHClient *ssh.SSHClient
}	

type InstallResult struct {
	Package     common.Package
	Success     bool
	Error       error
	Output      string
	InstallTime time.Duration
}

func NewAptManager(ctx *inventory.Context) (*AptManager, error) {
	ctx.Logger.SSHConnection(ctx.Host.Hostname, ctx.Host.User, fmt.Sprintf("%d", ctx.Host.Port))

	sshClient, err := ssh.NewSSHClient(ctx.Host)
	if err != nil {
		ctx.Logger.SSHError(err)
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	ctx.Logger.SSHSuccess()
	return &AptManager{
		SSHClient: sshClient,
	}, nil
}

func (m *AptManager) Install(ctx context.Context, runtimeCtx *inventory.Context, packages []common.Package) error {
	runtimeCtx.Logger.Info("Starting package installation...")

	results := make([]InstallResult, 0, len(packages))
	successCount := 0
	failureCount := 0

	for _, pkg := range packages {
		startTime := time.Now()

		var pkgName string
		if pkg.Version != "" && pkg.Version != "latest" {
			pkgName = fmt.Sprintf("%s=%s", pkg.Name, pkg.Version)
		} else {
			pkgName = pkg.Name
		}

		runtimeCtx.Logger.Info(fmt.Sprintf("Installing %s...", pkgName))

		command := fmt.Sprintf("sudo apt-get install -y %s", pkgName)
		runtimeCtx.Logger.Command(command)
		out, err := m.SSHClient.RunCommand(ctx, command)

		result := InstallResult{
			Package:     pkg,
			InstallTime: time.Since(startTime),
		}

		if err != nil {
			result.Success = false
			result.Error = err
			result.Output = out
			failureCount++

			runtimeCtx.Logger.Error(fmt.Sprintf("Failed to install %s: %v", pkgName, err))
			if out != "" {
				runtimeCtx.Logger.CommandOutput(out)
			}

			results = append(results, result)
			continue
		}

		result.Success = true
		result.Output = out
		successCount++

		runtimeCtx.Logger.Success(fmt.Sprintf("Successfully installed %s in %v", pkgName, result.InstallTime))
		if out != "" {
			runtimeCtx.Logger.CommandOutput(out)
		}

		results = append(results, result)
	}

	runtimeCtx.Logger.Info(fmt.Sprintf("Installation complete: %d successful, %d failed", successCount, failureCount))

	if failureCount == len(packages) {
		return fmt.Errorf("all package installations failed on host %s", runtimeCtx.Host.Name)
	}

	if failureCount > 0 {
		runtimeCtx.Logger.Warning("Failed packages:")
		for _, result := range results {
			if !result.Success {
				runtimeCtx.Logger.Error(fmt.Sprintf("  - %s: %v", result.Package.Name, result.Error))
			}
		}
	}

	return nil
}

func (m *AptManager) Remove(ctx context.Context, runtimeCtx *inventory.Context, packages []common.Package) error {
	runtimeCtx.Logger.Info("Starting package removal...")
	results := make([]InstallResult, 0, len(packages))
	successCount := 0
	failureCount := 0

	for _, pkg := range packages {
		startTime := time.Now()
		runtimeCtx.Logger.Info(fmt.Sprintf("Removing %s...", pkg.Name))

		var pkgName string
		if pkg.Version != "" && pkg.Version != "latest" {
			pkgName = fmt.Sprintf("%s=%s", pkg.Name, pkg.Version)
		} else {
			pkgName = pkg.Name
		}

		command := fmt.Sprintf("sudo apt-get remove -y %s", pkgName)
		runtimeCtx.Logger.Command(command)
		out, err := m.SSHClient.RunCommand(ctx, command)

		result := InstallResult{
			Package:     pkg,
			InstallTime: time.Since(startTime),
		}

		if err != nil {
			result.Success = false
			result.Error = err
			result.Output = out
			failureCount++

			runtimeCtx.Logger.Error(fmt.Sprintf("Failed to remove %s: %v", pkgName, err))
			if out != "" {
				runtimeCtx.Logger.CommandOutput(out)
			}

			results = append(results, result)
			continue
		}

		result.Success = true
		result.Output = out
		successCount++

		runtimeCtx.Logger.Success(fmt.Sprintf("Successfully removed %s in %v", pkgName, result.InstallTime))
		if out != "" {
			runtimeCtx.Logger.CommandOutput(out)
		}

		results = append(results, result)
	}

	runtimeCtx.Logger.Info(fmt.Sprintf("Removal complete: %d successful, %d failed", successCount, failureCount))

	if failureCount == len(packages) {
		return fmt.Errorf("all package removals failed on host %s", runtimeCtx.Host.Name)
	}

	if failureCount > 0 {
		runtimeCtx.Logger.Warning("Failed packages:")
		for _, result := range results {
			if !result.Success {
				runtimeCtx.Logger.Error(fmt.Sprintf("  - %s: %v", result.Package.Name, result.Error))
			}
		}
	}

	return nil
}

func (m *AptManager) DoesExist(ctx context.Context, runtimeCtx *inventory.Context, packages []common.Package) (bool, error) {
	runtimeCtx.Logger.Info("Checking if packages exist...")
	results := make([]InstallResult, 0, len(packages))
	successCount := 0
	failureCount := 0

	for _, pkg := range packages {
		startTime := time.Now()
		runtimeCtx.Logger.Info(fmt.Sprintf("Checking if %s exists...", pkg.Name))

		command := fmt.Sprintf("dpkg -l | grep -w %s", pkg.Name)
		runtimeCtx.Logger.Command(command)
		out, err := m.SSHClient.RunCommand(ctx, command)

		result := InstallResult{
			Package:     pkg,
			InstallTime: time.Since(startTime),
		}

		if err != nil {
			result.Success = false
			result.Error = err
			result.Output = out
			failureCount++

			runtimeCtx.Logger.Error(fmt.Sprintf("Failed to check if %s exists: %v", pkg.Name, err))
			if out != "" {
				runtimeCtx.Logger.CommandOutput(out)
			}

			results = append(results, result)
			continue
		}

		result.Success = true
		result.Output = out
		successCount++

		runtimeCtx.Logger.Success(fmt.Sprintf("Package %s exists", pkg.Name))
		if out != "" {
			runtimeCtx.Logger.CommandOutput(out)
		}

		results = append(results, result)
	}

	runtimeCtx.Logger.Info(fmt.Sprintf("Check complete: %d successful, %d failed", successCount, failureCount))

	if failureCount == len(packages) {
		return false, fmt.Errorf("all package checks failed on host %s", runtimeCtx.Host.Name)
	}

	if failureCount > 0 {
		runtimeCtx.Logger.Warning("Failed checks:")
		for _, result := range results {
			if !result.Success {
				runtimeCtx.Logger.Error(fmt.Sprintf("  - %s: %v", result.Package.Name, result.Error))
			}
		}
	}
	return successCount == len(packages), nil
}
