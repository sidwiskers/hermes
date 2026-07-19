// Command botapi-schema derives a deterministic schema inventory from the
// official Telegram Bot API HTML documentation.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sidwiskers/hermes/internal/botapi"
)

func main() {
	source := flag.String("source", botapi.OfficialSource, "official HTML URL or local snapshot")
	output := flag.String("output", "", "output JSON path; stdout when empty")
	flag.Parse()

	data, err := readSource(*source)
	if err != nil {
		fatal(err)
	}
	schema, err := botapi.ParseHTML(data)
	if err != nil {
		fatal(err)
	}
	encoded, err := botapi.Marshal(schema)
	if err != nil {
		fatal(err)
	}
	if *output == "" {
		_, err = os.Stdout.Write(encoded)
	} else {
		err = os.WriteFile(*output, encoded, 0o644)
	}
	if err != nil {
		fatal(err)
	}
}

func readSource(source string) ([]byte, error) {
	if !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://") {
		return os.ReadFile(source)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", source, response.Status)
	}
	return io.ReadAll(io.LimitReader(response.Body, 16<<20))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
