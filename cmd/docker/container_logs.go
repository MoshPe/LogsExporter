package docker

import (
	"context"
	"errors"
	"fmt"
	docker "github.com/LogsExporter/cmd"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/charmap"
	"io"
	"strings"
	"unicode/utf8"
)

// logsCmd represents the logs command
func logsCmd() *cobra.Command {
	var logsCmd = &cobra.Command{
		Use:   "logs CONTAINER",
		Short: "docker container logs [OPTIONS] CONTAINER",
		Long:  `Fetch the logs of a container`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			containerId := args[0]
			options := types.ContainerLogsOptions{
				ShowStdout: true,
				Timestamps: cmd.Flag("timestamps").Changed,
				Details:    cmd.Flag("details").Changed,
			}
			out, err := docker.Client.ContainerLogs(ctx, containerId, options)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", err.Error())
				return
			}

			a, _ := io.ReadAll(out)
			if !utf8.Valid(a) {
				fmt.Printf("%#v\n", out)
			}
			str := strings.ReplaceAll(string(a), "\\n", "\n")

			c := make([]byte, 0)

			for _, r := range str {
				if e, ok := charmap.ISO8859_1.EncodeRune(r); ok {
					c = append(c, e)
				}
			}

			reader := strings.NewReader(string(c))

			_, err = io.Copy(cmd.OutOrStdout(), reader)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", err.Error())
				return
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("required a container ID / name")
			}
			return nil
		},
	}
	logsFlags(logsCmd)
	return logsCmd
}

func init() {
	docker.RootCmd.AddCommand(logsCmd())
}
func logsFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("details", "", false, "Give extended privileges to the command")
	cmd.Flags().BoolP("timestamps", "", false, "Give extended privileges to the command")
}
