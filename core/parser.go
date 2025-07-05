package core

import (
	"fmt"

	"github.com/settlectl/settle-core/common"
)

// ResourceParser converts parsed data into Resource objects
type ResourceParser struct {
	hosts    []common.Host
	packages []common.Package
}

func NewResourceParser() *ResourceParser {
	return &ResourceParser{}
}

// SetHosts sets the hosts data for the parser (for context, not as resources)
func (rp *ResourceParser) SetHosts(hosts []common.Host) {
	rp.hosts = hosts
}

// SetPackages sets the packages data for the parser
func (rp *ResourceParser) SetPackages(packages []common.Package) {
	rp.packages = packages
}

// GetHosts returns the hosts (for context)
func (rp *ResourceParser) GetHosts() []common.Host {
	return rp.hosts
}

// CreatePackageResources converts Package objects to PackageResource objects
func (rp *ResourceParser) CreatePackageResources() ([]Resource, error) {
	var resources []Resource

	for _, pkg := range rp.packages {
		// Create a unique resource ID for the package
		resourceID := ResourceID(fmt.Sprintf("package:%s:%s", pkg.Manager, pkg.Name))

		// Create the package resource
		pkgResource := &PackageResource{
			BaseResource: BaseResource{
				ID:    resourceID,
				Type:  "package",
				Layer: LayerPlatform, // Packages are at the platform layer
				State: ResourceState{
					Status: StatePending,
				},
				Config: map[string]interface{}{
					"name":    pkg.Name,
					"version": pkg.Version,
					"manager": pkg.Manager,
				},
			},
			Package: pkg,
		}

		resources = append(resources, pkgResource)
	}

	return resources, nil
}

// ParseResources creates all resources from the stored data (excluding hosts)
func (rp *ResourceParser) ParseResources() ([]Resource, error) {
	var resources []Resource

	// Only create actual resources, not hosts
	// Hosts are targets, not resources to be created

	// Create package resources
	pkgResources, err := rp.CreatePackageResources()
	if err != nil {
		return nil, fmt.Errorf("failed to create package resources: %w", err)
	}
	resources = append(resources, pkgResources...)

	// TODO: Add other resource types (services, files, etc.)

	return resources, nil
}

// CreateResourceFromPackage creates a single PackageResource from a Package
func (rp *ResourceParser) CreateResourceFromPackage(pkg common.Package) Resource {
	resourceID := ResourceID(fmt.Sprintf("package:%s:%s", pkg.Manager, pkg.Name))

	return &PackageResource{
		BaseResource: BaseResource{
			ID:    resourceID,
			Type:  "package",
			Layer: LayerPlatform,
			State: ResourceState{
				Status: StatePending,
			},
			Config: map[string]interface{}{
				"name":    pkg.Name,
				"version": pkg.Version,
				"manager": pkg.Manager,
			},
		},
		Package: pkg,
	}
}

// ValidateResources validates all created resources
func (rp *ResourceParser) ValidateResources(resources []Resource) error {
	for _, resource := range resources {
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("resource %s validation failed: %w", resource.GetID(), err)
		}
	}
	return nil
}
