package core

import (
	"context"
	"fmt"
	"time"

	"github.com/settlectl/settle-core/common"
	pkgmanager "github.com/settlectl/settle-core/drivers/pkg"
	"github.com/settlectl/settle-core/inventory"
)

type ResourceID string
type EdgeType string
type Layer int
type StateStatus string
type ActionType string

const (
	ActionCreate ActionType = "create"
	ActionUpdate ActionType = "update"
	ActionDelete ActionType = "delete"
	ActionNoOp   ActionType = "no_op"
)

const (
	EdgeDependsOn  EdgeType = "depends_on"
	EdgeConfigures EdgeType = "configures"
	EdgeMonitors   EdgeType = "monitors"
	EdgeTriggers   EdgeType = "triggers"
)

const (
	LayerFoundation     Layer = iota //hardware, OS, etc.
	LayerPlatform                    //packageManagers, base services
	LayerInfrastructure              //databases, message queues stronrag
	LayerApplication                 //services and apps
	LayerConfiguration               //service configs, users, permissions
	LayerRuntime                     //processes, connections, healthchecks, etc.
)

const (
	StatePending StateStatus = "pending"
	StateApplied StateStatus = "applied"
	StateFailed  StateStatus = "failed"
	StateDrifted StateStatus = "drifted"
	StateSkipped StateStatus = "skipped"
	StateUnknown StateStatus = "unknown"
)

type Change struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

type Dependency struct {
	Target   ResourceID `json:"target"`
	EdgeType EdgeType   `json:"edge_type"`
	Required bool       `json:"required"`
}

type ResourceState struct {
	Status      StateStatus            `json:"status"`
	LastApplied time.Time              `json:"last_applied"`
	Checksum    string                 `json:"checksum"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type Action struct {
	ResourceID ResourceID             `json:"resource_id"`
	Type       ActionType             `json:"type"`
	Changes    []Change               `json:"changes"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type Resource interface {
	GetID() ResourceID
	GetType() string
	GetLayer() Layer

	GetDependencies() []Dependency
	AddDependency(dep Dependency) error
	GetState() *ResourceState
	SetState(state *ResourceState)
	GetConfig() map[string]interface{}
	SetConfig(config map[string]interface{})

	Validate() error

	Plan(ctx *inventory.Context) (*Action, error)
	Apply(ctx *inventory.Context) error
	Destroy(ctx *inventory.Context) error
}

type BaseResource struct {
	ID           ResourceID             `json:"id"`
	Type         string                 `json:"type"`
	Layer        Layer                  `json:"layer"`
	Dependencies []Dependency           `json:"dependencies"`
	State        ResourceState          `json:"state"`
	Config       map[string]interface{} `json:"config"`
}

func (r *BaseResource) GetID() ResourceID                       { return r.ID }
func (r *BaseResource) GetType() string                         { return r.Type }
func (r *BaseResource) GetLayer() Layer                         { return r.Layer }
func (r *BaseResource) GetDependencies() []Dependency           { return r.Dependencies }
func (r *BaseResource) GetState() *ResourceState                { return &r.State }
func (r *BaseResource) SetState(state *ResourceState)           { r.State = *state }
func (r *BaseResource) GetConfig() map[string]interface{}       { return r.Config }
func (r *BaseResource) SetConfig(config map[string]interface{}) { r.Config = config }

func (r *BaseResource) AddDependency(dep Dependency) error {
	r.Dependencies = append(r.Dependencies, dep)
	return nil
}

func (r *BaseResource) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("resource ID is required")
	}
	if r.Type == "" {
		return fmt.Errorf("resource type is required")
	}
	return nil
}

func (l Layer) String() string {
	layers := []string{
		"foundation",
		"platform",
		"infrastructure",
		"application",
		"configuration",
		"runtime",
	}
	if int(l) >= len(layers) {
		return "unknown"
	}
	return layers[l]
}

func ValidateLayerDependency(from, to Layer) error {
	if from < to {
		return fmt.Errorf("resource in layer %s cannot depend on layer %s", from.String(), to.String())
	}
	return nil
}

func (r *BaseResource) Plan(ctx *inventory.Context) (*Action, error) {
	return &Action{
		ResourceID: r.ID,
		Type:       ActionNoOp,
		Changes:    []Change{},
		Metadata:   make(map[string]interface{}),
	}, nil
}

func (r *BaseResource) Apply(ctx *inventory.Context) error {
	return fmt.Errorf("Apply not implemented for resource type %s", r.Type)
}

func (r *BaseResource) Destroy(ctx *inventory.Context) error {
	return fmt.Errorf("Destroy not implemented for resource type %s", r.Type)
}

// HostResource represents a host resource
type HostResource struct {
	BaseResource
	Host common.Host
}

func (r *HostResource) Apply(ctx *inventory.Context) error {
	// For host resources, we mainly validate connectivity
	// The actual host management would be done by other resources that depend on hosts

	ctx.Logger.Info(fmt.Sprintf("Validating host connectivity: %s", r.Host.Name))

	// Test SSH connectivity
	if ctx.SSHClient == nil {
		// Create SSH client for this host
		sshClient, err := ctx.CreateSSHClient(&r.Host)
		if err != nil {
			return fmt.Errorf("failed to create SSH client for host %s: %w", r.Host.Name, err)
		}
		ctx.SSHClient = sshClient
	}


	if err := ctx.SSHClient.TestConnection(); err != nil {
		return fmt.Errorf("host %s is not reachable: %w", r.Host.Name, err)
	}

	ctx.Logger.Info(fmt.Sprintf("Host %s is reachable", r.Host.Name))
	return nil
}

func (r *HostResource) Destroy(ctx *inventory.Context) error {

	ctx.Logger.Info(fmt.Sprintf("Cleaning up host: %s", r.Host.Name))


	if ctx.SSHClient != nil {
		ctx.SSHClient.Close()
	}

	return nil
}

// PackageResource represents a package resource
type PackageResource struct {
	BaseResource
	Package common.Package
}

func (r *PackageResource) Apply(ctx *inventory.Context) error {
	ctx.Logger.Info(fmt.Sprintf("Installing package: %s (manager: %s)", r.Package.Name, r.Package.Manager))

	// Get the appropriate package manager
	var manager pkgmanager.PackageManager
	var err error

	switch r.Package.Manager {
	case "apt":
		manager, err = pkgmanager.NewAptManager(ctx)
		if err != nil {
			return fmt.Errorf("failed to create apt manager: %w", err)
		}
	default:
		return fmt.Errorf("unsupported package manager: %s", r.Package.Manager)
	}

	// Check if package already exists
	exists, err := manager.DoesExist(context.Background(), ctx, []common.Package{r.Package})
	if err != nil {
		return fmt.Errorf("failed to check if package exists: %w", err)
	}

	if exists {
		ctx.Logger.Info(fmt.Sprintf("Package %s already installed", r.Package.Name))
		return nil
	}

	// Install the package
	err = manager.Install(context.Background(), ctx, []common.Package{r.Package})
	if err != nil {
		return fmt.Errorf("failed to install package %s: %w", r.Package.Name, err)
	}

	ctx.Logger.Info(fmt.Sprintf("Successfully installed package: %s", r.Package.Name))
	return nil
}

func (r *PackageResource) Destroy(ctx *inventory.Context) error {
	ctx.Logger.Info(fmt.Sprintf("Removing package: %s (manager: %s)", r.Package.Name, r.Package.Manager))

	// Get the appropriate package manager
	var manager pkgmanager.PackageManager
	var err error

	switch r.Package.Manager {
	case "apt":
		manager, err = pkgmanager.NewAptManager(ctx)
		if err != nil {
			return fmt.Errorf("failed to create apt manager: %w", err)
		}
	default:
		return fmt.Errorf("unsupported package manager: %s", r.Package.Manager)
	}

	// Remove the package
	err = manager.Remove(context.Background(), ctx, []common.Package{r.Package})
	if err != nil {
		return fmt.Errorf("failed to remove package %s: %w", r.Package.Name, err)
	}

	ctx.Logger.Info(fmt.Sprintf("Successfully removed package: %s", r.Package.Name))
	return nil
}

// ServiceResource represents a service resource
type ServiceResource struct {
	BaseResource
	Service struct {
		Name    string `json:"name"`
		State   string `json:"state"`   // running, stopped, enabled, disabled
		Manager string `json:"manager"` // systemd, rc, launchd
	}
}

func (r *ServiceResource) Apply(ctx *inventory.Context) error {
	ctx.Logger.Info(fmt.Sprintf("Managing service: %s (state: %s)", r.Service.Name, r.Service.State))

	// TODO: Implement service management
	// This would use the service drivers in drivers/svc/

	return fmt.Errorf("service management not yet implemented")
}

func (r *ServiceResource) Destroy(ctx *inventory.Context) error {
	ctx.Logger.Info(fmt.Sprintf("Stopping service: %s", r.Service.Name))

	// TODO: Implement service destruction

	return fmt.Errorf("service destruction not yet implemented")
}

// FileResource represents a file resource
type FileResource struct {
	BaseResource
	File struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Mode    int    `json:"mode"`
		Owner   string `json:"owner"`
		Group   string `json:"group"`
	}
}

func (r *FileResource) Apply(ctx *inventory.Context) error {
	ctx.Logger.Info(fmt.Sprintf("Creating/updating file: %s", r.File.Path))

	// TODO: Implement file management
	// This would use file system drivers

	return fmt.Errorf("file management not yet implemented")
}

func (r *FileResource) Destroy(ctx *inventory.Context) error {
	ctx.Logger.Info(fmt.Sprintf("Removing file: %s", r.File.Path))

	// TODO: Implement file removal

	return fmt.Errorf("file removal not yet implemented")
}
