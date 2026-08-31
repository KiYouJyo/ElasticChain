package main

import (
	"fmt"
	"os"

	"github.com/KiYouJyo/ElasticChain/internal/elastic"
)

const version = "0.0.1-dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "version":
		fmt.Println(version)
	case "demo-scaling":
		if err := demoScaling(); err != nil {
			fmt.Fprintln(os.Stderr, "demo failed:", err)
			os.Exit(1)
		}
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Printf(`ElasticChain %s

Research prototype commands:
  elasticchaind version       print prototype version
  elasticchaind demo-scaling  run a deterministic split/merge demonstration
`, version)
}

func demoScaling() error {
	policy := elastic.DefaultScalingPolicy()
	policy.ScaleOutConsecutiveEpochs = 2
	policy.ScaleInConsecutiveEpochs = 2
	policy.MaxDomains = 4

	planner, err := elastic.NewPlanner(policy)
	if err != nil {
		return err
	}
	topology := elastic.NewTopology()

	fmt.Println("epoch 0 active domains:", topology.ActiveDomains())

	for epoch := uint64(1); epoch <= 2; epoch++ {
		actions, err := planner.Plan(topology, []elastic.DomainMetrics{{
			DomainID: 0, Epoch: epoch, GasUtilizationBps: 9_000, DAUtilizationBps: 2_000,
		}})
		if err != nil {
			return err
		}
		if len(actions) > 0 {
			fmt.Printf("epoch %d actions: %v\n", epoch, actions)
			if err := planner.Apply(topology, actions); err != nil {
				return err
			}
		}
	}

	fmt.Println("after scale-out active domains:", topology.ActiveDomains())
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
			fmt.Printf("epoch %d actions: %v\n", epoch, actions)
			if err := planner.Apply(topology, actions); err != nil {
				return err
			}
		}
	}

	fmt.Println("after scale-in active domains:", topology.ActiveDomains())
	return nil
}
