// Copyright 2015 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.
//
// Author: Marc Berhault (marc@cockroachlabs.com)

package cli

import (
	"github.com/cockroachdb/cockroach-prod/docker"
	"github.com/cockroachdb/cockroach/util/log"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize a cockroach cluster",
	Long: `
Initialize a cockroach cluster. This initializes and starts the first node.
`,
	Run: runInit,
}

func runInit(cmd *cobra.Command, args []string) {
	driver, err := NewDriver(Context)
	if err != nil {
		log.Errorf("could not create driver: %v", err)
		return
	}

	nodes, err := docker.ListCockroachNodes()
	if err != nil {
		log.Errorf("failed to get list of existing cockroach nodes: %v", err)
		return
	}
	if len(nodes) != 0 {
		log.Errorf("init called but docker-machine has %d existing cockroach nodes: %v", len(nodes), nodes)
		return
	}

	nodeName := docker.MakeNodeName(0)

	// Create first node.
	if err := docker.CreateMachine(driver, nodeName); err != nil {
		log.Errorf("could not create machine %s: %v", nodeName, err)
		return
	}

	// Run driver steps after first-node creation.
	if err := driver.AfterFirstNode(); err != nil {
		log.Errorf("could not run AfterFirstNode steps for: %v", err)
		return
	}

	// Lookup node info.
	nodeConfig, err := driver.GetNodeConfig(nodeName)
	if err != nil {
		log.Errorf("could not get node config for %s: %v", nodeName, err)
		return
	}

	// Initialize cockroach node.
	if err := docker.RunDockerInit(driver, nodeName, nodeConfig); err != nil {
		log.Errorf("could not initialize first cockroach node %s: %v", nodeName, err)
		return
	}

	// Do "start node" logic.
	if err := driver.StartNode(nodeName, nodeConfig); err != nil {
		log.Errorf("could not run StartNode steps for %s: %v", nodeName, err)
		return
	}

	// Start the cockroach node.
	if err := docker.RunDockerStart(driver, nodeName, nodeConfig); err != nil {
		log.Errorf("could not initialize first cockroach node %s: %v", nodeName, err)
	}
}
