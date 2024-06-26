package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/constants"
	"github.com/akeyless-community/akeyless-ui-webhook-builder/internal/types"
)

func GenerateBashScript(mappings map[string][]types.Step) error {
	scriptContent := `#!/bin/bash

# The Rotated Secret Payload should look like this (the order of the fields is not important):
# {
#   "username": "my-username",
#   "password": "my-password",
#   "recording": {
#     "title": "Recording 5/3/2024 at 3:23:34 PM",
#     "steps": [
#       // ... (all the steps from the recording)
#    ]
#   }
# }
# For more information, visit https://docs.akeyless.io/docs/create-a-custom-rotated-secret

function run_rotate() {
    PAYLOAD=$(echo "$*" | base64 -d)
    PAYLOAD_VALUE=$(echo "$PAYLOAD" | jq -r .payload)
    
    # Extract current credentials
    CURRENT_USERNAME=$(echo "$PAYLOAD_VALUE" | jq -r .username)
    CURRENT_PASSWORD=$(echo "$PAYLOAD_VALUE" | jq -r .password)
    
    # Generate new password
    NEW_PASSWORD=$(dd bs=1000 count=1 if=/dev/urandom status=none | tr -dc '[:alnum:]' | head -c 15)
    
    # Use jq to update the recording JSON
    UPDATED_RECORDING=$(echo "$PAYLOAD_VALUE" | jq --arg CURRENT_USERNAME "$CURRENT_USERNAME" \
                                                  --arg CURRENT_PASSWORD "$CURRENT_PASSWORD" \
                                                  --arg NEW_PASSWORD "$NEW_PASSWORD" '
        .recording.steps |= map(
            %s
        )
    ')
    
    # Execute the updated recording using puppeteer-replay
    echo "$UPDATED_RECORDING" | jq -r .recording > updated_recording.json
    npx @puppeteer/replay updated_recording.json
    
    # Prepare the new payload
    NEW_PAYLOAD=$(echo "$UPDATED_RECORDING" | jq -r --arg NEW_PASSWORD "$NEW_PASSWORD" '{
        username: .username,
        password: $NEW_PASSWORD,
        recording: .recording
    }')
    
    # Format the payload as required by Akeyless
    PAYLOAD_JSON=$(echo -n "$NEW_PAYLOAD" | jq -Rsa . | sed -e 's/\\n//g' -e 's/\\t//g')
    PAYLOAD_JSON=$(echo -n "{ \"payload\": $PAYLOAD_JSON }")
    
    # Output the formatted payload
    echo -n "$PAYLOAD_JSON"
}
`

	jqConditions := []string{}
	for field, steps := range mappings {
		for _, step := range steps {
			selectorsJSON, _ := json.Marshal(step.Selectors)
			var value string
			switch field {
			case constants.FieldInputUsername:
				value = "$CURRENT_USERNAME"
			case constants.FieldInputPassword:
				value = "$CURRENT_PASSWORD"
			case constants.FieldNewPassword:
				value = "$NEW_PASSWORD"
			default:
				value = ".value" // Keep the original value for other fields
			}
			condition := fmt.Sprintf(`if .type == "change" and (.selectors | tostring) == %q then .value = %s`, string(selectorsJSON), value)
			jqConditions = append(jqConditions, condition)
		}
	}
	jqConditions = append(jqConditions, "else .")

	scriptContent = fmt.Sprintf(scriptContent, strings.Join(jqConditions, "\n            "))

	err := os.WriteFile("custom_logic.sh", []byte(scriptContent), 0644)
	if err != nil {
		fmt.Printf("Error writing bash script: %v\n", err)
		return err
	} else {
		fmt.Println("Bash script generated: custom_logic.sh")
		return nil
	}
}
