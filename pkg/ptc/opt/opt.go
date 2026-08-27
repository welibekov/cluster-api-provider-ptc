package opt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/welibekov/cluster-api-provider-ptc/pkg/ptc/task/types"
)

// Rest2Task converts a raw payload into a types.Task, stripping metadata fields starting with "_".
func Rest2Task(src interface{}, dest *types.Task) error {
	if src == nil || dest == nil {
		return errors.New("source payload and destination task pointer must not be nil")
	}

	// 1. Marshal source to JSON bytes
	srcBytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal source payload: %w", err)
	}

	// 2. Unmarshal into map to filter ignored keys (starting with "_")
	var rawMap map[string]interface{}
	if err := json.Unmarshal(srcBytes, &rawMap); err != nil {
		return fmt.Errorf("failed to unmarshal payload to map: %w", err)
	}

	filteredMap := make(map[string]interface{}, len(rawMap))
	for key, value := range rawMap {
		if !strings.HasPrefix(key, "_") {
			filteredMap[key] = value
		}
	}

	// 3. Remarshal filtered map
	filteredBytes, err := json.Marshal(filteredMap)
	if err != nil {
		return fmt.Errorf("failed to marshal filtered payload: %w", err)
	}

	// 4. Decode into destination struct
	decoder := json.NewDecoder(bytes.NewReader(filteredBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("failed to decode payload into Task: %w", err)
	}

	return nil
}
