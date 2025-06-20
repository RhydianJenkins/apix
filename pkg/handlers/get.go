package handlers

import (
	"fmt"

	"github.com/rhydianjenkins/apix/pkg/client"
	"github.com/rhydianjenkins/apix/pkg/config"
	"github.com/spf13/cobra"
)

func GetHandler(cmd *cobra.Command, args []string) {
	path := "/"

	if len(args) > 0 {
		path = args[0]
	}

	domain, err := config.LoadActiveDomain()

	if err != nil {
		println("Error loading domain:", err)
	}

	headers := make(map[string]string)
	res, body, err := client.MakeRequest("GET", domain, path, nil, headers)

	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	defer res.Body.Close()

	if len(body) == 0 {
		fmt.Println("Response body is empty. Headers:")
		for key, values := range res.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
		return
	}

	fmt.Printf("%s", string(body))
}
