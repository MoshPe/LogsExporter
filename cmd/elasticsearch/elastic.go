package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	docker "github.com/LogsExporter/cmd"
	"github.com/LogsExporter/cmd/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var (
	ES *elasticsearch.Client
)

var (
	Auth             bool
	ElasticMasterUrl string
	Username         string
	Password         string
)

var Elastic *cobra.Command

func elasticCmd() *cobra.Command {
	var logsCmd = &cobra.Command{
		Use:   "elastic",
		Short: "Elastic command to create or export default logs",
		Long:  `Elastic command to create or export default logs`,
	}
	return logsCmd
}

func init() {
	cobra.OnInitialize(InitElastic)
	Elastic = elasticCmd()
	elasticClientFlags(Elastic)
	docker.RootCmd.AddCommand(Elastic)
}
func elasticClientFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&ElasticMasterUrl, "node", "N", "http://localhost:9200", "Elastic master node to connect (default: localhost:9200)")
	cmd.PersistentFlags().StringVarP(&Username, "username", "u", "elastic", "Elastic auth username (default: elastic)")
	cmd.PersistentFlags().StringVarP(&Password, "password", "p", "changeme", "Elastic auth password (default: changeme)")
	cmd.PersistentFlags().BoolVarP(&Auth, "auth", "a", false, "Elastic auth enable (default: false)")
}

func InitElastic() {
	retryBackoff := backoff.NewExponentialBackOff()
	cfg := elasticsearch.Config{
		Addresses: []string{
			ElasticMasterUrl,
		},
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			log.Printf(`Connection retry %d `, i)

			return retryBackoff.NextBackOff()
		},
		MaxRetries: 4,
	}
	if Auth {
		cfg.Username = Username
		cfg.Password = Password
	}

	var err error
	ES, err = elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	info := ES.Info
	var req esapi.InfoRequest
	info.WithHuman()(&req)
	resp, err := req.Do(context.Background(), ES)
	if err != nil {
		log.Fatalf("Error formatting client output: %s", err)
	}
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
}
