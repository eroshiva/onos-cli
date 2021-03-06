// Copyright 2021-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ransim

import (
	"context"
	"fmt"
	"strconv"

	modelapi "github.com/onosproject/onos-api/go/onos/ransim/model"
	"github.com/onosproject/onos-api/go/onos/ransim/types"
	"github.com/onosproject/onos-lib-go/pkg/cli"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func getPlmnIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plmnid",
		Short: "Get the PLMNID",
		RunE:  runGetPlmnIDCommand,
	}
	return cmd
}

func getNodesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "Get all E2 nodes",
		RunE:  runGetNodesCommand,
	}
	cmd.Flags().Bool("no-headers", false, "disables output headers")
	cmd.Flags().BoolP("watch", "w", false, "watch node changes")
	return cmd
}

func createNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <enbid> [field options]",
		Args:  cobra.ExactArgs(1),
		Short: "Create an E2 node",
		RunE:  runCreateNodeCommand,
	}
	cmd.Flags().UintSlice("cells", []uint{}, "cell ECGIs")
	cmd.Flags().StringSlice("service-models", []string{}, "supported service models")
	cmd.Flags().StringSlice("controllers", []string{}, "E2T controller")
	return cmd
}

func getNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <enbid>",
		Args:  cobra.ExactArgs(1),
		Short: "Get an E2 node",
		RunE:  runGetNodeCommand,
	}
	return cmd
}

func updateNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <enbid> [field options]",
		Args:  cobra.ExactArgs(1),
		Short: "Update an E2 node",
		RunE:  runUpdateNodeCommand,
	}
	cmd.Flags().UintSlice("cells", []uint{}, "cell ECGIs")
	cmd.Flags().StringSlice("service-models", []string{}, "supported service models")
	cmd.Flags().StringSlice("controllers", []string{}, "E2T controller")
	return cmd
}

func deleteNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <enbid>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete an E2 node",
		RunE:  runDeleteNodeCommand,
	}
	return cmd
}

func startNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <enbid>",
		Args:  cobra.ExactArgs(1),
		Short: "Start E2 node agent",
		RunE:  runStartNodeCommand,
	}
	return cmd
}

func stopNodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <enbid>",
		Args:  cobra.ExactArgs(1),
		Short: "Stop E2 node agent",
		RunE:  runStopNodeCommand,
	}
	return cmd
}

func getNodeClient(cmd *cobra.Command) (modelapi.NodeModelClient, *grpc.ClientConn, error) {
	conn, err := cli.GetConnection(cmd)
	if err != nil {
		return nil, nil, err
	}
	return modelapi.NewNodeModelClient(conn), conn, nil
}

func runGetPlmnIDCommand(cmd *cobra.Command, args []string) error {
	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()
	resp, err := client.GetPlmnID(context.Background(), &modelapi.PlmnIDRequest{})
	if err != nil {
		return err
	}
	cli.Output("%d\n", resp.PlmnID)
	return nil
}

func runGetNodesCommand(cmd *cobra.Command, args []string) error {
	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	if noHeaders, _ := cmd.Flags().GetBool("no-headers"); !noHeaders {
		cli.Output("%-16s %-8s %-16s %-20s %s\n", "EnbID", "Status", "Service Models", "E2T Controllers", "Cell ECGIs")
	}

	if watch, _ := cmd.Flags().GetBool("watch"); watch {
		stream, err := client.WatchNodes(context.Background(), &modelapi.WatchNodesRequest{NoReplay: false})
		if err != nil {
			return err
		}
		for {
			r, err := stream.Recv()
			if err != nil {
				break
			}
			node := r.Node
			cli.Output("%-16d %-8s %-16s %-20s %s\n", node.EnbID, node.Status,
				catStrings(node.ServiceModels), catStrings(node.Controllers), catECGIs(node.CellECGIs))
		}

	} else {

		stream, err := client.ListNodes(context.Background(), &modelapi.ListNodesRequest{})
		if err != nil {
			return err
		}

		for {
			r, err := stream.Recv()
			if err != nil {
				break
			}
			node := r.Node
			cli.Output("%-16d %-8s %-16s %-20s %s\n", node.EnbID, node.Status,
				catStrings(node.ServiceModels), catStrings(node.Controllers), catECGIs(node.CellECGIs))
		}
	}

	return nil
}

func optionsToNode(cmd *cobra.Command, node *types.Node, update bool) (*types.Node, error) {
	cells, _ := cmd.Flags().GetUintSlice("cells")
	if !update || cmd.Flags().Changed("cells") {
		node.CellECGIs = toECGIs(cells)
	}

	models, _ := cmd.Flags().GetStringSlice("service-models")
	if !update || cmd.Flags().Changed("service-models") {
		node.ServiceModels = models
	}

	controllers, _ := cmd.Flags().GetStringSlice("controllers")
	if !update || cmd.Flags().Changed("controllers") {
		node.Controllers = controllers
	}
	return node, nil
}

func runCreateNodeCommand(cmd *cobra.Command, args []string) error {
	enbid, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	node, err := optionsToNode(cmd, &types.Node{EnbID: types.EnbID(enbid)}, false)
	if err != nil {
		return err
	}

	_, err = client.CreateNode(context.Background(), &modelapi.CreateNodeRequest{Node: node})
	if err != nil {
		return err
	}
	cli.Output("Node %d created\n", enbid)
	return nil
}

func runUpdateNodeCommand(cmd *cobra.Command, args []string) error {
	enbid, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Get the node first to prime the update node with existing values and allow sparse update
	gres, err := client.GetNode(context.Background(), &modelapi.GetNodeRequest{EnbID: types.EnbID(enbid)})
	if err != nil {
		return err
	}

	node, err := optionsToNode(cmd, gres.Node, true)
	if err != nil {
		return err
	}

	_, err = client.UpdateNode(context.Background(), &modelapi.UpdateNodeRequest{Node: node})
	if err != nil {
		return err
	}
	cli.Output("Node %d updated\n", enbid)
	return nil
}

func outputNode(node *types.Node) {
	cli.Output("EnbID: %-16d\nStatus: %s\nService Models: %s\nControllers: %s\nCell EGGIs: %s\n",
		node.EnbID, node.Status, catStrings(node.ServiceModels), catStrings(node.Controllers), catECGIs(node.CellECGIs))
}

func runGetNodeCommand(cmd *cobra.Command, args []string) error {
	enbid, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()
	res, err := client.GetNode(context.Background(), &modelapi.GetNodeRequest{EnbID: types.EnbID(enbid)})
	if err != nil {
		return err
	}

	outputNode(res.Node)
	return nil
}

func runDeleteNodeCommand(cmd *cobra.Command, args []string) error {
	enbid, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = client.DeleteNode(context.Background(), &modelapi.DeleteNodeRequest{EnbID: types.EnbID(enbid)})
	if err != nil {
		return err
	}

	cli.Output("Node %d deleted\n", enbid)
	return nil
}

func runControlCommand(command string, cmd *cobra.Command, args []string) error {
	enbid, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	client, conn, err := getNodeClient(cmd)
	if err != nil {
		return err
	}
	defer conn.Close()

	request := &modelapi.AgentControlRequest{EnbID: types.EnbID(enbid), Command: command}
	res, err := client.AgentControl(context.Background(), request)
	if err != nil {
		return err
	}
	outputNode(res.Node)
	return nil
}

func runStartNodeCommand(cmd *cobra.Command, args []string) error {
	return runControlCommand("start", cmd, args)
}

func runStopNodeCommand(cmd *cobra.Command, args []string) error {
	return runControlCommand("stop", cmd, args)
}

func toECGIs(ids []uint) []types.ECGI {
	ecgis := make([]types.ECGI, 0, len(ids))
	for _, id := range ids {
		ecgis = append(ecgis, types.ECGI(id))
	}
	return ecgis
}

func catECGIs(ecgis []types.ECGI) string {
	s := ""
	for _, ecgi := range ecgis {
		s = s + fmt.Sprintf(",%d", ecgi)
	}
	if len(s) > 1 {
		return s[1:]
	}
	return s
}

func catStrings(strings []string) string {
	s := ""
	for _, string := range strings {
		s = s + "," + string
	}
	if len(s) > 1 {
		return s[1:]
	}
	return s
}
