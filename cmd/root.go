// cmd/root.go

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/constants"
	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/generator"
	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/processor"
	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/types"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "akeyless-ui-webhook-builder",
	Short: "A CLI tool to build Akeyless UI webhooks from Chrome DevTools recordings",
	Run:   runMain,
}

var jsonFilePath string

func init() {
	rootCmd.Flags().StringVarP(&jsonFilePath, "file", "f", "", "Path to the recording JSON file (required)")
	rootCmd.MarkFlagRequired("file")
}

func Execute() error {
	return rootCmd.Execute()
}

func runMain(cmd *cobra.Command, args []string) {
	recording, err := processor.ReadRecordingFile(jsonFilePath)
	if err != nil {
		fmt.Printf("Error reading recording file: %v\n", err)
		os.Exit(1)
	}

	mappings, err := processor.GetUserMappings(recording.Steps)
	if err != nil {
		// User likely cancelled the operation
		os.Exit(0)
	}

	err = generator.GenerateBashScript(mappings)
	if err != nil {
		fmt.Printf("Error generating bash script: %v\n", err)
		os.Exit(1)
	}

	var username, oldPassword, newPassword string
	for field, steps := range mappings {
		if len(steps) > 0 {
			value := extractValue(steps, recording.Steps)
			switch field {
			case constants.FieldInputUsername:
				username = value
			case constants.FieldInputPassword:
				oldPassword = value
			case constants.FieldNewPassword:
				newPassword = value
			}
		}
	}

	// Create the payload object
	payload := map[string]interface{}{
		"username":  username,
		"password":  newPassword, // Use newPassword as the current password
		"recording": recording,
	}

	// Convert the payload to JSON
	payloadJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Printf("Error creating JSON payload: %v\n", err)
		os.Exit(1)
	}

	// Write the payload to a file
	err = os.WriteFile("payload.json", payloadJSON, 0644)
	if err != nil {
		fmt.Printf("Error writing payload to file: %v\n", err)
		os.Exit(1)
	}

	// Output information to stdout
	fmt.Println("\nOutput files generated successfully:")
	fmt.Println("1. custom_logic.sh - The bash script for credential rotation")
	fmt.Println("2. payload.json - The initial payload for the rotation process")

	fmt.Printf("\nPassword change summary:\n")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Old password: %s\n", oldPassword)
	fmt.Printf("New password field value: %s\n", newPassword)
	fmt.Println("\nNote: The 'password' field in payload.json contains the new password for the NEXT rotation.")
}

func extractValue(mappedSteps []types.Step, allSteps []types.Step) string {
	if len(mappedSteps) == 0 {
		return "PLACEHOLDER_VALUE"
	}
	for _, step := range allSteps {
		if step.Type == "change" && sameSelectors(step.Selectors, mappedSteps[0].Selectors) {
			return step.Value
		}
	}
	return "PLACEHOLDER_VALUE"
}

func sameSelectors(s1, s2 [][]string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if len(s1[i]) != len(s2[i]) {
			return false
		}
		for j := range s1[i] {
			if s1[i][j] != s2[i][j] {
				return false
			}
		}
	}
	return true
}
