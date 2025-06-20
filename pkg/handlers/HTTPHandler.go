package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/client"
	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func HTTPHandler(method string, cmd *cobra.Command, args []string) {
	path := "/"
    if len(args) > 0 {
        path = args[0]
    }

    var body []byte
    if len(args) > 1 {
        body = []byte(args[1])
    }

	domain, err := config.LoadActiveDomain()
	if err != nil {
		println("Error loading domain:", err)
	}

	headers := make(map[string]string)
	res, body, err := client.MakeRequest(method, domain, path, body, headers)

	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	defer res.Body.Close()

	if len(body) == 0 {
		fmt.Printf("Code: %d. Body: empty. Headers:\n", res.StatusCode)
		for key, values := range res.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
		return
	}

	fmt.Printf("%s", string(body))
}
