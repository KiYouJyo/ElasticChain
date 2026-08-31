package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/KiYouJyo/ElasticChain/internal/elastic"
)

func newDemoScalingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "demo-scaling",
		Short: "Run the deterministic execution-domain split/merge demonstration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return demoScaling(cmd)
		},
	}
}

func demoScaling(cmd *cobra.Command) error {
	policy := elastic.DefaultScalingPolicy()
	policy.ScaleOutConsecutiveEpochs = 2
	policy.ScaleInConsecutiveEpochs = 2
	policy.MaxDomains = 4

	planner, err := elastic.NewPlanner(policy)
	if err != nil {
		return err
	}
	topology := elastic.NewTopology()

	fmt.Fprintln(cmd.OutOrStdout(), "epoch 0 active domains:", topology.ActiveDomains())

	for epoch := uint64(1); epoch <= 2; epoch++ {
		actions, err := planner.Plan(topology, []elastic.DomainMetrics{{
			DomainID: 0, Epoch: epoch, GasUtilizationBps: 9_000, DAUtilizationBps: 2_000,
		}})
		if err != nil {
			return err
		}
		if len(actions) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "epoch %d actions: %v\n", epoch, actions)
			if err := planner.Apply(topology, actions); err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "after scale-out active domains:", topology.ActiveDomains())
	active := topology.ActiveDomains()
	for epoch := uint64(3); epoch <= 4; epoch++ {
		metrics := make([]elastic.DomainMetrics, 0, len(active))
		for _, id := range active {
			metrics = append(metrics, elastic.DomainMetrics{
				DomainID: id, Epoch: epoch, GasUtilizationBps: 500, DAUtilizationBps: 500,
			})
		}
		actions, err := planner.Plan(topology, metrics)
		if err != nil {
			return err
		}
		if len(actions) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "epoch %d actions: %v\n", epoch, actions)
			if err := planner.Apply(topology, actions); err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "after scale-in active domains:", topology.ActiveDomains())
	return nil
}
