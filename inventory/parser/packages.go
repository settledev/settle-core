package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/settlectl/settle-core/common"
)

func ParsePackages(path string) ([]common.Package, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var packages []common.Package
	var pkg common.Package

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, common.MaxFileSize)
	scanner.Buffer(buf, common.MaxFileSize)

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		if lineCount > common.MaxFileSize/100 {
			return nil, fmt.Errorf("too many lines in file")
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "package ") {
			if pkg.Name != "" {
				packages = append(packages, pkg)
				pkg = common.Package{}
			}

			if len(packages) >= common.MaxHosts {
				return nil, fmt.Errorf("too many packages (max: %d)", common.MaxHosts)
			}

			line = strings.TrimSuffix(line, "{")
			line = strings.TrimSpace(line)
			parts := strings.Split(line, "\"")
			if len(parts) >= 2 {
				pkgName := parts[1]
				if len(pkgName) > common.MaxNameLength {
					return nil, fmt.Errorf("package name too long: %s", pkgName)
				}
				pkg.Name = pkgName
			}
		} else if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)

			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			switch key {
			case "version":
				//TODO: validate version format
				pkg.Version = val
			case "manager":
				//TODO: validate package managers
				pkg.Manager = val
			}
		}
	}
	if pkg.Name != "" {
		packages = append(packages, pkg)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return packages, nil
}
