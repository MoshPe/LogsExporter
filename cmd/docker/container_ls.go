package docker

import (
	"context"
	"fmt"
	docker "github.com/LogsExporter/cmd"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

var containerName string

// lsCmd represents the ls command
func containerLsCmd() *cobra.Command {
	var containerLsCmd = &cobra.Command{
		Use:   "ls",
		Short: "docker container ls",
		Long:  `List containers`,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				containers []types.Container
				err        error
				//imageLen   int
				//portLen    int
			)
			ctx := context.Background()
			//can add filtering options

			if cmd.Flag("filter").Changed && cmd.Flag("all").Changed {
				containers, err = docker.Client.ContainerList(ctx, types.ContainerListOptions{
					All: cmd.Flag("all").Changed,
					Filters: filters.NewArgs(
						serviceFilter(containerName),
					),
				})
				if err != nil {
					panic(err)
				}
			} else if cmd.Flag("filter").Changed {
				containers, err = docker.Client.ContainerList(ctx, types.ContainerListOptions{
					Filters: filters.NewArgs(
						serviceFilter(containerName),
					),
				})
				if err != nil {
					panic(err)
				}
			} else {
				containers, err = docker.Client.ContainerList(ctx, types.ContainerListOptions{
					All: cmd.Flag("all").Changed,
				})
				if err != nil {
					panic(err)
				}
			}

			if len(containers) == 0 {
				fmt.Println("There are no containers available")
				return
			}

			// Create a tabwriter
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

			// Print header
			fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tPORTS\tNAMES")

			// Print container information
			for _, container := range containers {
				createdTime := time.Since(time.Unix(container.Created, 0)).Round(time.Second)
				ports := formatPorts(container.Ports)

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
					container.ID[:12], container.Image, container.Command,
					formatDuration(createdTime.String()), ports, strings.Join(container.Names, ", "))
			}

			// Flush the tabwriter
			w.Flush()

			//imageLen = findMaxImageLength(containers)
			//if imageLen < len("CONTAINER IMAGE") {
			//	imageLen = 3
			//}
			//portLen, ports := findMaxPortLength(containers)
			//if portLen < len("PORTS") {
			//	imageLen = 3
			//}
			//for i, container := range containers {
			//	printContainers[i+1] = ""
			//	printContainers[i+1] = fmt.Sprintf("%3.12s %-*s%-*s%s%s%s%s%s",
			//		container.ID, 4, container.Image, imageLen, container.Command, time.Unix(container.Created, 0).Format(time.RFC850), container.Status, ports[i],
			//		strings.Repeat(" ", portLen), container.Names[0][1:])
			//}
			//printContainers[0] = fmt.Sprintf("CONTAINER ID   CONTAINER IMAGE%s"+
			//	"COMMAND   CREATED   STATUS   PORTS   %s   NAMES", strings.Repeat(" ", imageLen), strings.Repeat(" ", portLen))
			//
			//for _, container := range printContainers {
			//	fmt.Println(container)
			//}
		},
	}
	lsFlags(containerLsCmd)
	return containerLsCmd
}

func serviceFilter(serviceName string) filters.KeyValuePair {
	return filters.Arg("name", serviceName)
}

func findMaxImageLength(containers []types.Container) int {
	max := 0
	for _, container := range containers {
		temp := len(container.Image)
		if temp > max {
			max = temp
		}
	}
	return max
}

func formatPorts(ports []types.Port) string {
	var formattedPorts []string
	for _, port := range ports {
		formattedPorts = append(formattedPorts, fmt.Sprintf("%s:%d->%s:%d", port.IP, port.PublicPort, port.IP, port.PrivatePort))
	}
	return strings.Join(formattedPorts, ", ")
}

func formatDuration(durationString string) string {
	duration, err := time.ParseDuration(durationString)
	if err != nil {
		fmt.Println("Error parsing duration:", err)
		return ""
	}
	switch {
	case duration.Seconds() < 60:
		return fmt.Sprintf("%d seconds ago", int(duration.Seconds()))
	case duration.Minutes() < 60:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration.Hours() < 24:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	default:
		// Calculate days and weeks
		days := int(duration.Hours() / 24)
		weeks := int(duration.Hours() / (24 * 7))

		// Choose the appropriate format based on the duration
		switch {
		case days < 7:
			return fmt.Sprintf("%d days ago", days)
		default:
			return fmt.Sprintf("%d weeks ago", weeks)
		}
	}
}

func findMaxPortLength(containers []types.Container) (int, []string) {
	max := 0
	portStr := make([]string, len(containers))
	for i, container := range containers {
		for _, port := range container.Ports {
			portStr[i] += fmt.Sprintf("%s:%d->%d/%s, ", port.IP, port.PublicPort, port.PrivatePort, port.Type)
			temp := len(portStr)
			if temp > max {
				max = temp
			}
		}
	}
	return max, portStr
}

func init() {
	docker.RootCmd.AddCommand(containerLsCmd())
}

func lsFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&containerName, "name", "", "", "Get all the containers(with exit codes)")
	cmd.Flags().BoolP("filter", "f", false, "Get all the containers(with exit codes)")
	cmd.Flags().BoolP("all", "", false, "Get all the containers(with exit codes)")
}
