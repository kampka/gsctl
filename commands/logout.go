package commands

import (
	"errors"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/giantswarm/gsclientgen"
	"github.com/spf13/cobra"

	"github.com/giantswarm/gsctl/config"
)

var (
	// LogoutCommand performs a logout
	LogoutCommand = &cobra.Command{
		Use:     "logout",
		Short:   "Sign out the current user",
		Long:    `This will terminate the current user's session and invalidate the authentication token.`,
		PreRunE: checkLogout,
		Run:     logout,
	}
)

func checkLogout(cmd *cobra.Command, args []string) error {
	if config.Config.Token == "" && cmdToken == "" {
		return errors.New("You are not logged in")
	}
	return nil
}

func logout(cmd *cobra.Command, args []string) {
	client := gsclientgen.NewDefaultApi()

	// if token is set via flags, we unauthenticate using this token
	authHeader := "giantswarm " + config.Config.Token
	if cmdToken != "" {
		authHeader = "giantswarm " + cmdToken
	}

	logoutResponse, apiResponse, err := client.UserLogout(authHeader)
	if err != nil {
		fmt.Println("Info: The client doesn't handle the API's 401 response yet.")
		fmt.Println("Seeing this error likely means: The passed token was no longer valid.")
		fmt.Println("Error details:")
		log.Fatal(err)
	}
	if logoutResponse.StatusCode == 10007 {
		// remove token from settings
		// unless we unathenticated the token from flags
		if cmdToken == "" {
			config.Config.Token = ""
			config.Config.Email = ""
		}
		fmt.Println(color.GreenString("Successfully logged out"))
	} else {
		fmt.Printf("Unhandled response code: %v", logoutResponse.StatusCode)
		fmt.Printf("Status text: %v", logoutResponse.StatusText)
		fmt.Printf("apiResponse: %s\n", apiResponse)
	}

	config.WriteToFile()
}
