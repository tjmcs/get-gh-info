/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
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

/*
 * a utility function that can be used to return the stats (min, max, median, average
 * first quartile, and third quartile) of a slice of data along with the length of
 * the slice
 */
func GetJsonDurationStats(data []time.Duration) (map[string]JsonDuration, int) {
	sliceLen := len(data)
	// initialize a variable to hold the results
	zeroDuration := JsonDuration{time.Duration(0)}
	results := map[string]JsonDuration{"minimum": zeroDuration,
		"firstQuartile": zeroDuration, "median": zeroDuration, "average": zeroDuration,
		"thirdQuartile": zeroDuration, "maximum": zeroDuration}
	// if the slice is empty, just return the results
	if sliceLen == 0 {
		return results, sliceLen
	}
	// if we have more than one value, then sort it from least to greatest
	if sliceLen > 1 {
		sort.Slice(data, func(i, j int) bool {
			return data[i] < data[j]
		})
	}
	// grab the two obvious values
	results["minimum"] = JsonDuration{data[0]}
	results["maximum"] = JsonDuration{data[sliceLen-1]}
	results["average"] = JsonDuration{GetAverageDuration(data)}
	// and grab the remaining stats based on the length of the slice; note that
	// the first quartile is defined as the median of the lower half of the
	// slice of durations and the third quartile is the median of the uppoer
	// half fo that same slice, so the behavior changes a bit depending on
	// the slice length
	switch sliceLen {
	case 1:
		// only one value; all three (first quartile, median, and third quartile)
		// are the same
		results["median"] = JsonDuration{data[0]}
		results["firstQuartile"] = results["median"]
		results["thirdQuartile"] = results["median"]
	case 2:
		// if two values, the median is the average of those two numbers
		results["median"] = JsonDuration{(data[0] + data[1]) / 2}
		fallthrough
	case 3:
		// if three values, then the median is the middle value
		results["median"] = JsonDuration{data[1]}
		// for either a two or three value slice, the median of lower
		// half is the first value and the median of the second half is
		// the last value
		results["firstQuartile"] = JsonDuration{data[0]}
		results["thirdQuartile"] = JsonDuration{data[sliceLen-1]}
	default:
		// for the rest, we have to know if this is an odd-length slice or
		// an even-length slice
		if sliceLen%2 > 0 {
			// if here, it's a odd length slice so things are easy
			results["median"] = JsonDuration{data[sliceLen/2+1]}
			results["firstQuartile"] = JsonDuration{data[sliceLen/4+1]}
			results["thirdQuartile"] = JsonDuration{data[(sliceLen*3)/4+1]}
		} else {
			// if here, it's an even length slice so we have to do a little more work
			// in this case the values we want are the average of the two values around
			// where the actual value would be in an odd-length list
			results["median"] = JsonDuration{(data[sliceLen/2] + data[sliceLen/2+1]) / 2}
			results["firstQuartile"] = JsonDuration{(data[sliceLen/4] + data[sliceLen/4+1]) / 2}
			results["thirdQuartile"] = JsonDuration{(data[(sliceLen*3)/4] + data[(sliceLen*3)/4+1]) / 2}
		}
	}
	// finally, return the results
	return results, sliceLen
}

/*
 * defind a type that lets us dump out a time.Duration as a
 * formatted string in JSON
 */
type JsonDuration struct {
	time.Duration
}

func (j JsonDuration) format() string {
	d := j.Duration
	if d >= (time.Hour * 24) {
		return fmt.Sprintf("%.2fd", d.Hours()/24)
	} else if d >= time.Hour {
		return fmt.Sprintf("%.2fh", d.Hours())
	} else if d >= time.Minute {
		return fmt.Sprintf("%.2fm", d.Minutes())
	} else if d >= time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else if d >= time.Millisecond {
		return fmt.Sprintf("%dms", d.Round(time.Millisecond))
	} else if d >= time.Microsecond {
		return fmt.Sprintf("%dus", d.Round(time.Microsecond))
	} else if d == 0 {
		return "0s"
	}
	// else, if we get here, just return the number of nanoseconds
	return fmt.Sprintf("%dns", d.Nanoseconds())
}

func (j JsonDuration) MarshalText() ([]byte, error) {
	return []byte(j.format()), nil
}

func (j JsonDuration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.format() + `"`), nil
}
