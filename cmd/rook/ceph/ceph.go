/*
Copyright 2018 The Rook Authors. All rights reserved.

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
package ceph

import (
	"github.com/coreos/pkg/capnslog"
	"github.com/spf13/cobra"

	"github.com/rook/rook/cmd/rook/rook"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/daemon/ceph/mon"
	osdconfig "github.com/rook/rook/pkg/operator/ceph/cluster/osd/config"
	"github.com/rook/rook/pkg/util/exec"
	"github.com/rook/rook/pkg/util/flags"
)

var Cmd = &cobra.Command{
	Use:    "ceph",
	Short:  "Main command for Ceph operator and daemons.",
	Hidden: true,
}

var (
	cfg         = &config{}
	clusterInfo mon.ClusterInfo
	logger      = capnslog.NewPackageLogger("github.com/rook/rook", "cephcmd")
)

type config struct {
	devices            string
	directories        string
	metadataDevice     string
	dataDir            string
	forceFormat        bool
	location           string
	cephConfigOverride string
	storeConfig        osdconfig.StoreConfig
	networkInfo        clusterd.NetworkInfo
	monEndpoints       string
	nodeName           string
}

func init() {
	AddCommands(Cmd)
}

func AddCommands(command *cobra.Command) {
	command.AddCommand(operatorCmd)
	command.AddCommand(agentCmd)
	command.AddCommand(monCmd)
	command.AddCommand(osdCmd)
	command.AddCommand(mgrCmd)
	command.AddCommand(rgwCmd)
	command.AddCommand(mdsCmd)
}

func createContext() *clusterd.Context {
	executor := &exec.CommandExecutor{}
	return &clusterd.Context{
		Executor:           executor,
		ConfigDir:          cfg.dataDir,
		ConfigFileOverride: cfg.cephConfigOverride,
		LogLevel:           rook.Cfg.LogLevel,
		NetworkInfo:        cfg.networkInfo.Simplify(),
	}
}

func addCephFlags(command *cobra.Command) {
	command.Flags().StringVar(&cfg.networkInfo.PublicAddr, "public-ip", "127.0.0.1", "public IP address for this machine")
	command.Flags().StringVar(&cfg.networkInfo.ClusterAddr, "private-ip", "127.0.0.1", "private IP address for this machine")
	command.Flags().StringVar(&clusterInfo.Name, "cluster-name", "rookcluster", "ceph cluster name")
	command.Flags().StringVar(&clusterInfo.FSID, "fsid", "", "the cluster uuid")
	command.Flags().StringVar(&clusterInfo.MonitorSecret, "mon-secret", "", "the cephx keyring for monitors")
	command.Flags().StringVar(&clusterInfo.AdminSecret, "admin-secret", "", "secret for the admin user (random if not specified)")
	command.Flags().StringVar(&cfg.monEndpoints, "mon-endpoints", "", "ceph mon endpoints")
	command.Flags().StringVar(&cfg.dataDir, "config-dir", "/var/lib/rook", "directory for storing configuration")
	command.Flags().StringVar(&cfg.cephConfigOverride, "ceph-config-override", "", "optional path to a ceph config file that will be appended to the config files that rook generates")

	// deprecated ipv4 format address
	// TODO: remove these legacy flags in the future
	command.Flags().StringVar(&cfg.networkInfo.PublicAddrIPv4, "public-ipv4", "127.0.0.1", "public IPv4 address for this machine")
	command.Flags().StringVar(&cfg.networkInfo.ClusterAddrIPv4, "private-ipv4", "127.0.0.1", "private IPv4 address for this machine")
	command.Flags().MarkDeprecated("public-ipv4", "Use --public-ip instead. Will be removed in a future version.")
	command.Flags().MarkDeprecated("private-ipv4", "Use --private-ip instead. Will be removed in a future version.")
}

func verifyRenamedFlags(cmd *cobra.Command) error {
	renamed := []flags.RenamedFlag{
		{NewFlagName: "public-ip", OldFlagName: "public-ipv4"},
		{NewFlagName: "private-ip", OldFlagName: "private-ipv4"},
	}
	return flags.VerifyRenamedFlags(cmd, renamed)
}