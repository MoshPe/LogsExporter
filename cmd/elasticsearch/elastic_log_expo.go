package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	docker "github.com/LogsExporter/cmd"
	"github.com/docker/docker/api/types"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/charmap"
	"io"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

type Event struct {
	Name string
	Time int64
}

var (
	index           string
	numWorkers      int
	flushBytes      int
	countSuccessful uint64
)

// logsCmd represents the logs command
func elasticLogExpoCmd() *cobra.Command {
	var logsCmd = &cobra.Command{
		Use:   "export",
		Short: "LogExpo elastic export [OPTIONS] CONTAINER",
		Long:  `Export the container logs to elastic index`,
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

			parts := strings.FieldsFunc(string(c), func(r rune) bool {
				return r == ',' || r == '\n'
			})

			var events []Event
			for i := 0; i < len(parts)-1; i += 2 {
				name := strings.TrimSpace(parts[i])
				timeStr := strings.TrimSpace(parts[i+1])

				// Convert the time string to uint32
				time, err := strconv.ParseInt(timeStr, 10, 64)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				//time, err := strconv.ParseUint(timeStr, 10, 32)
				//if err != nil {
				//	fmt.Println("Error parsing time:", err)
				//	continue
				//}

				// Create an Event object and append to the array
				event := Event{
					Name: name,
					Time: time,
				}
				events = append(events, event)
			}

			bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
				Index:         index,            // The default index name
				Client:        ES,               // The Elasticsearch client
				NumWorkers:    numWorkers,       // The number of worker goroutines
				FlushBytes:    flushBytes,       // The flush threshold in bytes
				FlushInterval: 30 * time.Second, // The periodic flush interval
			})
			if err != nil {
				log.Fatalf("Error creating the indexer: %s", err)
			}

			start := time.Now().UTC()
			for _, a := range events {
				// Prepare the data payload: encode article to JSON
				//
				data, err := json.Marshal(a)
				if err != nil {
					log.Fatalf("Cannot encode event %s: %s", a.Name, err)
				}

				err = bi.Add(
					context.Background(),
					esutil.BulkIndexerItem{
						// Action field configures the operation to perform (index, create, delete, update)
						Action: "index",

						// Body is an `io.Reader` with the payload
						Body: bytes.NewReader(data),

						// OnSuccess is called for each successful operation
						OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
							atomic.AddUint64(&countSuccessful, 1)
						},

						// OnFailure is called for each failed operation
						OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
							if err != nil {
								log.Printf("ERROR: %s", err)
							} else {
								log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
							}
						},
					},
				)
				if err != nil {
					log.Fatalf("Unexpected error: %s", err)
				}
			}

			if err := bi.Close(context.Background()); err != nil {
				log.Fatalf("Unexpected error: %s", err)
			}

			biStats := bi.Stats()

			// Report the results: number of indexed docs, number of errors, duration, indexing rate
			//
			log.Println(strings.Repeat("▔", 65))

			dur := time.Since(start)

			if biStats.NumFailed > 0 {
				log.Fatalf(
					"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
					humanize.Comma(int64(biStats.NumFlushed)),
					humanize.Comma(int64(biStats.NumFailed)),
					dur.Truncate(time.Millisecond),
					humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
				)
			} else {
				log.Printf(
					"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
					humanize.Comma(int64(biStats.NumFlushed)),
					dur.Truncate(time.Millisecond),
					humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
				)
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
	Elastic.AddCommand(elasticLogExpoCmd())
}
func logsFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("details", "", false, "Give extended privileges to the command")
	cmd.Flags().BoolP("timestamps", "", false, "Give extended privileges to the command")
	cmd.Flags().StringVarP(&index, "index", "i", "LogExpoLogs", "Elastic index to push results (default: LogExpoLogs")
	cmd.Flags().IntVarP(&numWorkers, "workers", "w", runtime.NumCPU(), "Number of indexer workers")
	cmd.Flags().IntVarP(&flushBytes, "flush", "f", 5e+6, "Flush threshold in bytes (default: "+strconv.FormatInt(5e+6, 10)+")")
}
