package processor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/constants"
	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/generator"
	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/types"
	"github.com/charmbracelet/huh"
)

func ReadRecordingFile(filePath string) (types.Recording, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return types.Recording{}, err
	}

	var recording types.Recording
	err = json.Unmarshal(data, &recording)
	if err != nil {
		return types.Recording{}, err
	}

	return recording, nil
}

func ProcessFile(filePath string) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var recording types.Recording
	if err := json.Unmarshal(jsonData, &recording); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	changeSteps := filterChangeSteps(recording.Steps)
	mappings, err := GetUserMappings(changeSteps)
	if err != nil {
		fmt.Printf("Error getting user mappings: %v\n", err)
		os.Exit(1)
	}

	genMappings := make(map[string][]types.Step)
	for k, v := range mappings {
		genSteps := make([]types.Step, len(v))
		for i, step := range v {
			genSteps[i] = types.Step(step)
		}
		genMappings[k] = genSteps
	}

	generator.GenerateBashScript(genMappings)
}

func filterChangeSteps(steps []types.Step) []types.Step {
	var changeSteps []types.Step
	for _, step := range steps {
		if step.Type == "change" {
			changeSteps = append(changeSteps, step)
		}
	}
	return changeSteps
}

func GetUserMappings(steps []types.Step) (map[string][]types.Step, error) {
	mappings := make(map[string][]types.Step)

	// Create and run the landing page
	landingForm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Akeyless UI Webhook Builder").
				Description(
					"Welcome to the Akeyless UI Webhook Builder!\n\n" +
						"This tool will guide you through mapping input fields from your recording captured using\n" +
						"the Chrome DevTools Recorder to the necessary credential rotation fields needed for the Akeyless Custom Rotator.\n\n" +
						"Important Rules:\n" +
						"1. There are three required fields: Input Username, Input Password, and New Password.\n" +
						"2. Each required field can be mapped to one or more input fields from your recording.\n" +
						"3. Multiple mappings for a field are useful for processes with double authentication or similar scenarios.\n" +
						"4. There is an optional 'Required Static Input' field for any additional static inputs needed in the rotation process.\n" +
						"5. If there are no duplicate fields to map, you can select 'Done' to move to the next field.\n\n" +
						"Mapping Process:\n" +
						"- For each field, you'll be presented with a list of available input fields from your recording.\n" +
						"- Select the appropriate input(s) for each field.\n" +
						"- If you've selected all necessary inputs for a field, choose 'Done' to move to the next field.\n" +
						"- For required fields, you must select at least one input before you can choose 'Done'.\n\n" +
						"Output Files:\n" +
						"1. custom_logic.sh - The bash script for credential rotation\n" +
						"2. payload.json - The initial payload for the rotation process\n\n" +
						"Let's begin the mapping process!\n\n",
				).
				Next(true).
				NextLabel("Start Mapping"),
		),
	)

	err := landingForm.Run()
	if err != nil {
		fmt.Printf("Error displaying landing page: %v\n", err)
		return mappings, err
	}

	availableSteps := make(map[string]types.Step)
	for _, step := range steps {
		if step.Type == "change" {
			key := generateStepKey(step)
			availableSteps[key] = step
		}
	}

	// Process required fields
	for _, field := range constants.RequiredFields {
		mappings[field], err = getMultipleSelections(availableSteps, field, true)
		if err != nil {
			return mappings, err
		}
	}

	// Process optional fields if there are any steps left
	if len(availableSteps) > 0 {
		for _, field := range constants.OptionalFields {
			mappings[field], err = getMultipleSelections(availableSteps, field, false)
			if err != nil {
				return mappings, err
			}
			if len(availableSteps) == 0 {
				break
			}
		}
	}

	return mappings, nil
}

// internal/processor/processor.go

func getMultipleSelections(availableSteps map[string]types.Step, field string, required bool) ([]types.Step, error) {
	var selectedSteps []types.Step

	for {
		options := make([]huh.Option[string], 0, len(availableSteps)+1)
		for key, step := range availableSteps {
			label := formatStepLabel(step)
			options = append(options, huh.NewOption(label, key))
		}

		// Add 'Done' option
		doneOption := huh.NewOption("Done", "done")

		var selectedKey string

		if !required || len(selectedSteps) > 0 {

			options = append(options, doneOption)
			if len(selectedSteps) > 0 {
				selectedKey = "done" // Set 'Done' as default if field is already set
			}
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Select a step for %s%s", field, getDoneMessage(required, len(selectedSteps)))).
					Options(options...).
					Value(&selectedKey),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Printf("Error getting user input: %v\n", err)
			return selectedSteps, err
		}

		if selectedKey == "done" {
			if required && len(selectedSteps) == 0 {
				fmt.Printf("%s is required. Please select at least one step.\n", field)
				continue
			}
			break
		}

		if step, exists := availableSteps[selectedKey]; exists {
			selectedSteps = append(selectedSteps, step)
			delete(availableSteps, selectedKey)
		}
	}

	return selectedSteps, nil
}

func formatStepLabel(step types.Step) string {
	selectorSummary := make([]string, len(step.Selectors))
	for i, selector := range step.Selectors {
		if len(selector) > 0 {
			selectorSummary[i] = selector[0]
		}
	}
	return fmt.Sprintf("Input Field ID: %s | Selectors: %s", step.Value, strings.Join(selectorSummary, ", "))
}

func getDoneMessage(required bool, selections int) string {
	if required && selections == 0 {
		return " (required)"
	}
	return " (select another or choose 'Done' if no duplicate fields are left to map for this input)"
}

func generateStepKey(step types.Step) string {
	// Create a unique key for each step based on its selectors
	h := sha256.New()
	for _, selector := range step.Selectors {
		h.Write([]byte(strings.Join(selector, ",")))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
