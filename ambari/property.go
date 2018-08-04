// Copyright 2018 Oliver Szabo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ambari

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GetMinimalBlueprint obtain minimal blueprint - compare properties with stack default properties and get a minimal blueprint configuration
func (a AmbariRegistry) GetMinimalBlueprint(blueprint map[string]interface{}, stackDefaults map[string]StackConfig) []byte {
	if configurationsVal, ok := blueprint["configurations"]; ok {
		miniConfig := make(map[string]map[string]interface{})
		configEntries := configurationsVal.([]interface{})
		for _, configEntry := range configEntries {
			confI := configEntry.(map[string]interface{})
			for configType, props := range confI {
				stackDefaultsConfigType := stackDefaults[configType]
				propsI := props.(map[string]interface{})
				properties := propsI["properties"].(map[string]interface{})
				stackDefaultProperties := stackDefaultsConfigType.Properties
				for propertyKey, propertyVal := range properties {
					property := propertyVal.(string)
					for _, stackDefaultProperty := range stackDefaultProperties {
						if strings.Compare(stackDefaultProperty.Name, propertyKey) == 0 {
							if (propertyKey == "content" && strings.Compare(strings.TrimSpace(stackDefaultProperty.Value), strings.TrimSpace(property)) != 0) ||
								(propertyKey != "content" && strings.Compare(stackDefaultProperty.Value, property) != 0) {
								if properties, ok := miniConfig[configType]; ok {
									properties[propertyKey] = property
									miniConfig[configType] = properties
								} else {
									properties := make(map[string]interface{})
									properties[propertyKey] = property
									miniConfig[configType] = properties
								}
							}
						}
					}
				}
			}
		}
		if len(miniConfig) > 0 {
			configurations := convertToConfigurationsFromMap(miniConfig)
			blueprint["configurations"] = configurations
		} else {
			blueprint["configurations"] = make([]interface{}, 0)
		}
	}
	bodyBytes, err := json.Marshal(blueprint)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return bodyBytes
}

func convertToConfigurationsFromMap(minimalConfig map[string]map[string]interface{}) []map[string]map[string]interface{} {
	configurations := make([]map[string]map[string]interface{}, len(minimalConfig))
	cnt := 0
	for propertyKey, propertyValue := range minimalConfig {
		confMap := make(map[string]map[string]interface{})
		confMap[propertyKey] = propertyValue
		configurations[cnt] = confMap
		cnt++
	}
	return configurations
}
