/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// define a default lookback time of 90 days
const defaultLookbackDays = 90

// and a method to see if a slice of strings contains a given string
func SliceContains(sl []string, name string) bool {
	for _, v := range sl {
		if v == name {
			return true
		}
	}
	return false
}

/*
 * a utility function that finds the smallest index at which val == strSlice[i],
 * (or -1 if there is no such index)
 */
func FindIndexOf(val string, strSlice []string) int {
	for i, n := range strSlice {
		if val == n {
			return i
		}
	}
	return -1
}

/*
 * a utility function used to get the average duration of a list of durations
 */
func GetAverageDuration(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, duration := range durations {
		total += duration
	}
	return total / time.Duration(len(durations))
}

/*
 * a function that can be used to read a generic YAML file
 */

func ReadYamlFile(fileName string) []map[string]interface{} {
	// read the contents of the file
	yfile, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: while reading input YAML file '%s'; %s\n", fileName, err)
		os.Exit(-6)
	}
	data := make([]map[interface{}]interface{}, 5)
	err2 := yaml.Unmarshal(yfile, &data)
	if err2 != nil {
		fmt.Fprintf(os.Stderr, "ERROR: while unmarshaling data from input YAML file '%s'; %s\n", fileName, err)
		os.Exit(-6)
	}
	// convert the list of maps of interfaces to interfaces into a list of maps of strings
	listOfStringMaps := []map[string]interface{}{}
	for _, item := range data {
		listOfStringMaps = append(listOfStringMaps, convInterToInterMapToStringToInterMap(item))
	}
	return listOfStringMaps
}

/*
 * a utility function that can be used to convert a map of interfaces to interfaces to
 * a map of strings to interfaces
 */
func convInterToInterMapToStringToInterMap(inputMap map[interface{}]interface{}) map[string]interface{} {
	outputMap := map[string]interface{}{}
	for key, val := range inputMap {
		outputMap[key.(string)] = val
	}
	return outputMap
}

/*
 * a function that can be used to dump out the results of the query as a
 * formatted JSON string
 */

func DumpMapAsJSON(results interface{}) {
	// first, get the JSON encoding of the results
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to marshal results to JSON: %v", err)
		os.Exit(-7)
	}
	// then, print the results to stdout
	fmt.Println(string(jsonBytes))
}
