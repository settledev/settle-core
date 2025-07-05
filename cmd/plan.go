package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/core"
	"github.com/settlectl/settle-core/inventory"
	"github.com/settlectl/settle-core/inventory/parser"
	"github.com/spf13/cobra"
)

var (
	planOutput string
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "show what would be executed",
	Run: func(cmd *cobra.Command, args []string) {
		logger := inventory.NewLogger()
		logger.Info("Creating execution plan")

		hosts, err := parser.ParseHosts("hosts.stl")
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing hosts file: %v", err))
			return
		}
		logger.Info(fmt.Sprintf("Found %d hosts", len(hosts)))

		resourceFiles, err := findResourceFiles()
		if err != nil {
			logger.Error(fmt.Sprintf("Error finding resource files: %v", err))
			return
		}

		resourceParser := core.NewResourceParser()
		resourceParser.SetHosts(hosts)

		var allPackages []common.Package
		for _, file := range resourceFiles {
			packages, err := parser.ParsePackages(file)
			if err != nil {
				logger.Error(fmt.Sprintf("Error parsing packages from %s: %v", file, err))
				continue
			}
			allPackages = append(allPackages, packages...)
		}
		resourceParser.SetPackages(allPackages)

		resources, err := resourceParser.ParseResources()
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating resources: %v", err))
			return
		}
		logger.Info(fmt.Sprintf("Created %d resources", len(resources)))

		graph := core.NewGraph()
		for _, resource := range resources {
			if err := graph.AddResource(resource); err != nil {
				logger.Error(fmt.Sprintf("Error adding resource %s to graph: %v", resource.GetID(), err))
				continue
			}
		}

		if err := graph.ValidateDependencies(); err != nil {
			logger.Error(fmt.Sprintf("Graph validation failed: %v", err))
			return
		}

		stateManager := core.NewStateManager(".settle/state.json", graph)
		if err := stateManager.LoadState(); err != nil {
			logger.Error(fmt.Sprintf("Error loading state: %v", err))
			return
		}

		planner := core.NewPlanner(graph, stateManager, logger)
		plan, err := planner.Plan()
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating plan: %v", err))
			return
		}

		logger.Info("=== EXECUTION PLAN ===")
		logger.Info(fmt.Sprintf("Plan created at: %s", plan.CreatedAt.Format("2006-01-02 15:04:05")))
		logger.Info("")

		logger.Info("Summary:")
		logger.Info(fmt.Sprintf("  Create: %d resources", plan.GetActionCount(core.ActionCreate)))
		logger.Info(fmt.Sprintf("  Update: %d resources", plan.GetActionCount(core.ActionUpdate)))
		logger.Info(fmt.Sprintf("  Delete: %d resources", plan.GetActionCount(core.ActionDelete)))
		logger.Info(fmt.Sprintf("  No-op: %d resources", plan.GetActionCount(core.ActionNoOp)))
		logger.Info("")

		if len(plan.Actions) > 0 {
			logger.Info("Detailed Actions:")
			for i, action := range plan.Actions {
				logger.Info(fmt.Sprintf("  %d. %s (%s)", i+1, action.ResourceID, action.Type))
				if reason, ok := action.Metadata["reason"]; ok {
					logger.Info(fmt.Sprintf("      Reason: %s", reason))
				}

				resource, exists := graph.GetResource(action.ResourceID)
				if exists {
					config := resource.GetConfig()
					logger.Info(fmt.Sprintf("      Type: %s", resource.GetType()))
					logger.Info(fmt.Sprintf("      Layer: %s", resource.GetLayer().String()))

					if len(config) > 0 {
						logger.Info("      Configuration:")
						for key, value := range config {
							logger.Info(fmt.Sprintf("        %s: %v", key, value))
						}
					}
				}
				logger.Info("")
			}
		} else {
			logger.Info("No changes needed. All resources are up to date.")
		}

		logger.Info("")
		logger.Info("To apply this plan, run: settlectl create")

		if planOutput != "" {
			if err := savePlanToFile(plan, planOutput); err != nil {
				logger.Error(fmt.Sprintf("Error saving plan to file: %v", err))
				return
			}
			logger.Info(fmt.Sprintf("Plan saved to: %s", planOutput))
		}
	},
}

func savePlanToFile(plan *core.Plan, filename string) error {
	planOutput := struct {
		CreatedAt string                 `json:"created_at"`
		Summary   map[string]int         `json:"summary"`
		Actions   []*core.Action         `json:"actions"`
		Resources map[string]interface{} `json:"resources"`
	}{
		CreatedAt: plan.CreatedAt.Format("2006-01-02 15:04:05"),
		Summary: map[string]int{
			"create": plan.GetActionCount(core.ActionCreate),
			"update": plan.GetActionCount(core.ActionUpdate),
			"delete": plan.GetActionCount(core.ActionDelete),
			"no_op":  plan.GetActionCount(core.ActionNoOp),
		},
		Actions:   plan.Actions,
		Resources: make(map[string]interface{}),
	}

	for _, action := range plan.Actions {
		resource, exists := plan.Graph.GetResource(action.ResourceID)
		if exists {
			planOutput.Resources[string(action.ResourceID)] = map[string]interface{}{
				"type":   resource.GetType(),
				"layer":  resource.GetLayer().String(),
				"config": resource.GetConfig(),
				"action": action.Type,
				"reason": action.Metadata["reason"],
			}
		}
	}

	data, err := json.MarshalIndent(planOutput, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	return nil
}

func init() {
	planCmd.Flags().StringVarP(&planOutput, "output", "o", "", "Output plan to file")
	rootCmd.AddCommand(planCmd)
}
