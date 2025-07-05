package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"github.com/settlectl/settle-core/common"
)






func validateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(hostname) > common.MaxNameLength {
		return fmt.Errorf("hostname too long: %d characters", len(hostname))
	}
	

	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$|^(\d{1,3}\.){3}\d{1,3}$`)
	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format: %s", hostname)
	}
	return nil
}


func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be 1-65535)", port)
	}
	return nil
}


func sanitizePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}
	

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	

	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("path contains directory traversal: %s", path)
	}
	
	return absPath, nil
}


func validateFileSize(file *os.File) error {
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}
	
	if stat.Size() > common.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", stat.Size(), common.MaxFileSize)
	}
	
	return nil
}

func ParseHosts(path string) ([]common.Host, error) {

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	

	if strings.Contains(path, "..") {
		return nil, fmt.Errorf("path contains directory traversal: %s", path)
	}
	
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()


	if err := validateFileSize(file); err != nil {
		return nil, err
	}

	var hosts []common.Host
	var current common.Host
	scanner := bufio.NewScanner(file)
	

	buf := make([]byte, 0, common.MaxLineLength)
	scanner.Buffer(buf, common.MaxLineLength)

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

		if strings.HasPrefix(line, "host ") {
			if current.Name != "" {
				hosts = append(hosts, current)
				current = common.Host{}
			}
			

			if len(hosts) >= common.MaxHosts {
				return nil, fmt.Errorf("too many hosts (max: %d)", common.MaxHosts)
			}
			
			line = strings.TrimSuffix(line, "{")
			line = strings.TrimSpace(line)
			parts := strings.Split(line, "\"")
			if len(parts) >= 2 {
				hostName := parts[1]
				if len(hostName) > common.MaxNameLength {
					return nil, fmt.Errorf("host name too long: %s", hostName)
				}
				current.Name = hostName
			}
		} else if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue 
			}
			
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			
			switch key {
			case "hostname":
				if err := validateHostname(val); err != nil {
					return nil, fmt.Errorf("invalid hostname in host %s: %w", current.Name, err)
				}
				current.Hostname = val
			case "user":
				if len(val) > common.MaxNameLength {
					return nil, fmt.Errorf("username too long in host %s", current.Name)
				}
				current.User = val
			case "port":
				port, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid port in host %s: %w", current.Name, err)
				}
				if err := validatePort(port); err != nil {
					return nil, fmt.Errorf("invalid port in host %s: %w", current.Name, err)
				}
				current.Port = port
			case "key_file", "keyfile":
				sanitizedPath, err := sanitizePath(val)
				if err != nil {
					return nil, fmt.Errorf("invalid key_file in host %s: %w", current.Name, err)
				}
				current.Keyfile = sanitizedPath
			case "group":
				if len(val) > common.MaxNameLength {
					return nil, fmt.Errorf("group name too long in host %s", current.Name)
				}
				current.Group = val
			}
		}
	}
	
	if current.Name != "" {
		hosts = append(hosts, current)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	
	return hosts, nil
}
