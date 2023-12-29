package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LogsExporter/cmd/utils"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var (
	shards   int
	replicas int
)

// logsCmd represents the logs command
func elasticIndexCmd() *cobra.Command {
	var logsCmd = &cobra.Command{
		Use:   "index",
		Short: "LogExpo elastic index [Flags]",
		Long:  `Create an index in elastic with given flags as index configuration`,
		Run: func(cmd *cobra.Command, args []string) {
			InitElastic()
			index := args[0]

			mapping := `{
					"properties": {
						"Time": { "type": "date" },
						"Name": { "type": "keyword" }
					}
				}`

			requestBody := fmt.Sprintf(`{
					"mappings": %s
				}`, mapping)
			resp, err := ES.Indices.Create(
				index,
				ES.Indices.Create.WithBody(strings.NewReader(requestBody)),
				ES.Indices.Create.WithPretty(),
			)

			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					fmt.Println("Error closing response body:", closeErr)
				}
			}()
			var data interface{}
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				fmt.Println("Error decoding JSON:", err)
				return
			}
			output, err := utils.FormatOutput(data)
			if err != nil {
				log.Fatalf("Error formatting client output: %s", err)
			}
			log.Println(output)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("required an index name")
			}
			return nil
		},
	}
	elasticFlags(logsCmd)
	return logsCmd
}

func init() {
	Elastic.AddCommand(elasticIndexCmd())
}
func elasticFlags(cmd *cobra.Command) {
	cmd.Flags().IntVarP(&shards, "shards", "s", 1, "Number of shards for creating an index (default: 1)")
	cmd.Flags().IntVarP(&replicas, "replicas", "r", 1, "Number of replicas for creating an index (default: 1)")
}
