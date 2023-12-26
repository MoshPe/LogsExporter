package cmd

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"os"
)

type hostTLS struct {
	HostIPAndPort string
	TLSCaCertPath string
	TLSCertPath   string
	TLSKeyPath    string
}

var (
	host hostTLS
)

var Client *client.Client

// RootCmd represents the docker command
var RootCmd = &cobra.Command{
	Use:   "LogExpo",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	host = hostTLS{}
	cobra.OnInitialize(InitClient)
	RootCmd.PersistentFlags().BoolP("tlsverify", "", false, "Use TLS and verify the remote")
	RootCmd.PersistentFlags().StringVarP(&host.HostIPAndPort, "host", "H", "", "Daemon socket to connect to")
	RootCmd.PersistentFlags().StringVar(&host.TLSCaCertPath, "tlscacert", os.ExpandEnv("$CLIENT_CONFIG/ca.pem"), "Trust certs signed only by this CA")
	RootCmd.PersistentFlags().StringVar(&host.TLSCertPath, "tlscert", os.ExpandEnv("$CLIENT_CONFIG/cert.pem"), "Path to TLS certificate file")
	RootCmd.PersistentFlags().StringVar(&host.TLSKeyPath, "tlskey", os.ExpandEnv("$CLIENT_CONFIG/key.pem"), "Path to TLS key file")
}

func InitClient() {
	var err error
	var clientWithHost client.Opt
	var clientWithTLS client.Opt
	if RootCmd.Flag("tlsverify").Changed {
		clientWithTLS = client.WithTLSClientConfig(
			os.ExpandEnv(host.TLSCaCertPath),
			os.ExpandEnv(host.TLSCertPath),
			os.ExpandEnv(host.TLSKeyPath))

	}

	if host.HostIPAndPort != "" {
		host.HostIPAndPort = "://" + os.ExpandEnv(host.HostIPAndPort)
		clientWithHost = client.WithHost(host.HostIPAndPort)
	}

	if clientWithTLS != nil && clientWithHost != nil {
		Client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(),
			clientWithHost, clientWithTLS)
	} else if clientWithHost != nil {
		Client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation(),
			clientWithHost)
	} else {
		Client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
