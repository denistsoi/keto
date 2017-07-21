/*
Copyright 2017 The Keto Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/UKHomeOffice/keto/pkg/cloudprovider"
	"github.com/UKHomeOffice/keto/pkg/constants"
	"github.com/UKHomeOffice/keto/pkg/controller"
	"github.com/UKHomeOffice/keto/pkg/userdata"

	"github.com/spf13/cobra"
)

var (
	// errNotImplemented is an error for not implemented features.
	errNotImplemented = errors.New("not implemented")

	// KetoCmd represents the root command when called without any subcommands
	KetoCmd = &cobra.Command{
		Use:   "keto",
		Short: "Kubernetes clusters manager",
		Long:  "Kubernetes clusters manager",
		RunE: func(c *cobra.Command, args []string) error {
			if c.Flags().Changed("version") {
				versionCmdFunc()
				return nil
			}
			return c.Usage()
		},
	}

	// subcommand aliases
	clusterCmdAliases     = []string{"cl", "clusters"}
	masterPoolCmdAliases  = []string{"mp", "master", "masters", "masterpools"}
	computePoolCmdAliases = []string{"cp", "compute", "computes", "computepools"}
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := KetoCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

// cli respresents keto cli client.
type cli struct {
	logger      *log.Logger
	debugLogger *log.Logger
	ctrl        *controller.Controller
}

// newCLI returns a new instance of cli. It is expected to be used by
// keto cli subcommands.
func newCLI(c *cobra.Command) (*cli, error) {
	if !c.Flags().Changed("cloud") {
		return &cli{}, fmt.Errorf("cloud provider name is not specified")
	}

	logger := log.New(os.Stdout, "", 0)
	debugLogger := log.New(os.Stderr, "[debug] ", log.Ldate|log.Ltime|log.Lshortfile)
	debug, err := c.Flags().GetBool("debug")
	if err != nil {
		return &cli{}, err
	}
	if !debug {
		debugLogger.SetOutput(ioutil.Discard)
	}

	cloudName, err := c.Flags().GetString("cloud")
	if err != nil {
		return &cli{}, err
	}

	cloud, err := cloudprovider.InitCloudProvider(cloudName, debugLogger)
	if err != nil {
		return &cli{}, err
	}

	ud := userdata.New(debugLogger)
	ctrl := controller.New(
		controller.Config{
			Logger:   debugLogger,
			Cloud:    cloud,
			UserData: ud,
		})

	return &cli{
		logger:      logger,
		debugLogger: debugLogger,
		ctrl:        ctrl,
	}, nil
}

func init() {
	// Local flags
	KetoCmd.Flags().BoolP("help", "h", false, "Help message")
	KetoCmd.Flags().BoolP("version", "v", false, "Print version")

	// Global flags
	KetoCmd.PersistentFlags().String("cloud", "",
		"Cloud provider name. Supported providers: "+strings.Join(cloudprovider.CloudProviders(), ", "))
	// TODO: set default to false once we're happy with the tool.
	KetoCmd.PersistentFlags().Bool("debug", true, "Enable debug logging")

	KetoCmd.AddCommand(
		getCmd,
		createCmd,
		deleteCmd,
		describeCmd,
		updateCmd,
		versionCmd,
	)
}

// addClusterFlag adds a cluster flag
func addClusterFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("cluster", "", "Cluster name")
	}
}

// addInternalFlag adds an internal flag
func addInternalFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().Bool("internal", false, "Create an internal cluster")
	}
}

// addNetworksFlag adds a networks flag
func addNetworksFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().StringSlice("networks", []string{}, "Cloud specific list of comma separated networks")
	}
}

// addCoreOSVersionFlag adds an OS flag
func addCoreOSVersionFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("coreos-version", "", fmt.Sprintf("Operating system (default %q)", constants.DefaultCoreOSVersion))
	}
}

// addSSHKeyFlag adds an ssh-key flag
func addSSHKeyFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("ssh-key", "", "Public SSH key or name (dependent on cloud provider)")
	}
}

// addDiskSizeFlag adds a disk-size flag
func addDiskSizeFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().Int("disk-size", 0, fmt.Sprintf("Node boot disk size in GB (default %d)", constants.DefaultDiskSizeInGigabytes))
	}
}

// addMachineTypeFlag adds a machine type flag
func addMachineTypeFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("machine-type", "", "Machine type")
	}
}

// addPoolSizeFlag adds a size flag
func addPoolSizeFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().Int("pool-size", 0,
			fmt.Sprintf("Number of nodes in the compute pool (default %d)", constants.DefaultComputePoolSize))
	}
}

// addDNSZoneFlag adds a DNS zone flag
func addDNSZoneFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("dns-zone", "", "Hosted DNS zone name")
	}
}

// addLabelsFlag adds labels flag
func addLabelsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().StringSlice("labels", []string{}, "List of labels in a comma separated key=value format")
	}
}

// addTaintsFlag adds taints flag
func addTaintsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().StringSlice("taints", []string{}, "List of taints in a comma separated key=value format")
	}
}

// addKubeVersionFlag adds a kubernetes version flag
func addKubeVersionFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("kube-version", constants.DefaultKubeVersion, "Kubernetes version")
	}
}

// addAssetsDirFlag adds an assets dir flag.
func addAssetsDirFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("assets-dir", "", "The path to etcd/kube CA certs and keys")
	}
}

// addComputePoolsFlag adds a compute pools flag
func addComputePoolsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().Int("compute-pools", 1, "Number of compute pools to create")
	}
}

// addKubeletExtraArgsFlag adds a kubelet extra arguments flag
func addKubeletExtraArgsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("kubelet-extra-args", "", "Kubelet extra arguments")
	}
}

// addAPIServerExtraArgsFlag adds an api-server extra arguments flag
func addAPIServerExtraArgsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("api-server-extra-args", "", "Kubernetes api-server extra arguments")
	}
}

// addControllerManagerExtraArgsFlag adds a controller-manager extra arguments flag
func addControllerManagerExtraArgsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("controller-manager-extra-args", "", "Kubernetes controller-manager extra arguments")
	}
}

// addSchedulerExtraArgsFlag adds a scheduler extra arguments flag
func addSchedulerExtraArgsFlag(c ...*cobra.Command) {
	for _, i := range c {
		i.Flags().String("scheduler-extra-args", "", "Kubernetes scheduler extra arguments")
	}
}
