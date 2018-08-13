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

// ConvertStingsToMap generate a map from strings (like key=value)
func ConvertStingsToMap(keyValueStrings []string) map[string]string {
	responseMap := make(map[string]string)
	for _, keyValue := range keyValueStrings {
		if len(keyValue) > 0 {
			keyValuePair := strings.SplitN(keyValue, "=", 2)
			if len(keyValuePair) == 2 {
				key := strings.TrimSpace(keyValuePair[0])
				value := keyValuePair[1]
				responseMap[key] = value
			}
		}
	}
	return responseMap
}

// GetConfigValue get a value from the blueprint for a speficfic config property with a config type
func GetConfigValue(blueprint map[string]interface{}, configType string, configProperty string) string {
	if configurationsVal, ok := blueprint["configurations"]; ok {
		configEntries := configurationsVal.([]interface{})
		for _, configEntry := range configEntries {
			confI := configEntry.(map[string]interface{})
			for confType, props := range confI {
				propsI := props.(map[string]interface{})
				properties := propsI["properties"].(map[string]interface{})
				for propertyKey, propertyVal := range properties {
					property := propertyVal.(string)
					if confType == configType && propertyKey == configProperty {
						return property
					}
				}
			}
		}
	}
	return ""
}

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
						fillConfWithChangedProperties(stackDefaultProperty, propertyKey, property, miniConfig, configType)
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

func fillConfWithChangedProperties(stackDefaultProperty StackProperty, propertyKey string, property string, miniConfig map[string]map[string]interface{}, configType string) {
	if strings.Compare(stackDefaultProperty.Name, propertyKey) == 0 {
		if (propertyKey == "content" && strings.Compare(strings.TrimSpace(stackDefaultProperty.Value), strings.TrimSpace(property)) != 0) ||
			(propertyKey != "content" && strings.Compare(stackDefaultProperty.Value, property) != 0) {
			if configTypeVal, ok := miniConfig[configType]; ok {
				properties := configTypeVal["properties"].(map[string]interface{})
				properties[propertyKey] = property
				miniConfig[configType]["properties"] = properties
			} else {
				properties := make(map[string]interface{})
				propertyAttributes := make(map[string]interface{})
				propertiesAndAttributes := make(map[string]interface{})
				properties[propertyKey] = property
				propertiesAndAttributes["properties"] = properties
				propertiesAndAttributes["properties_attributes"] = propertyAttributes
				miniConfig[configType] = propertiesAndAttributes
			}
		}
	}
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
