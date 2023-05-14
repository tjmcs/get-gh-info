/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package utils

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
)

/*
 * a function that can be used to find the date corresponding to the start of
 * the week that corresponds to the input date; in this function he start of
 * the week is assumed to be at midnight on Monday, and the date returned will
 * be the Monday of the week that contains the input date (which is perfect if
 * we want to look back a certain amount of time from the input date but only
 * want to include data from complete weeks)
 */
func weekStartDate(date time.Time) time.Time {
	// first, determine the offset between the weekday for the input date
	// and previous Monday (i.e. the start of the week) in days
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	// then, add that offset to the input date to get the start of the week
	if offset < 0 {
		date = date.Add(time.Duration(offset*24) * time.Hour)
	}
	// and return the result
	return date
}

/*
 * a function that can be used to parse the "lookback time" string value can be passed in
 * on the command-line; supported time units include:
 *     - days (e.g. "7d")
 *     - weeks (e.g. "12w")
 *     - months (e.g. "3m"); here a month is assumed to be 30 days for convenience
 *     - quarters (e.g. "2q"); here a quarter is assumed to be 90 days for convenience
 *     - years (e.g. "1y")
 *
 * it should be noted that due to limitations in the GitHub GraphQL API, the maximum
 * lookback time is limited to one year
 */
func getLookbackDuration(lookBackStr string) time.Duration {
	// define a regular expression to parse the lookback string
	parsePattern := "^([+-]?[0-9]+)(d|w|m|q|y)$"
	re := regexp.MustCompile(parsePattern)
	// search for a match in the lookback string
	matches := re.FindStringSubmatch(lookBackStr)
	if matches == nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to parse duration '%s'; expected format is '[+-]?[0-9]+[dwmqy]'\n", lookBackStr)
		os.Exit(-1)
	}
	// if a match was found, grab the value
	durationVal, err := strconv.Atoi(matches[1])
	// and use the accompanying time unit to return the appropriate time.Duration value
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to parse duration '%s'; expected format is '[+-]?[0-9]+[dwmqy]'\n", lookBackStr)
		os.Exit(-1)
	}
	switch matches[2] {
	case "d":
		return time.Duration(durationVal) * 24 * time.Hour
	case "w":
		return time.Duration(durationVal) * 7 * 24 * time.Hour
	case "m":
		return time.Duration(durationVal) * 30 * 24 * time.Hour
	case "q":
		return time.Duration(durationVal) * 3 * 30 * 24 * time.Hour
	case "y":
		return time.Duration(durationVal) * 365 * 24 * time.Hour
	default:
		fmt.Fprintf(os.Stderr, "ERROR: unable to parse duration '%s'; expected format is '[+-]?[0-9]+[dwmqy]'\n", lookBackStr)
		os.Exit(-1)
	}
	return 0
}

/*
 * a function that can be used to get a time window to use for our queries; note that
 * the logic around this is a lot more difficult than it might seem at first because
 * we support both positive and negative lookback times, only pulling data from complete
 * weeks, and default behaviors when either the lookback time is not specified, the
 * reference date is not specified, or both are not specified (with some of these
 * edge cases triggering a "lookahead" mode rather than a "lookback" mode)
 */
func GetQueryTimeWindow() (githubv4.DateTime, githubv4.DateTime) {
	// setup a few variables that we'll be using in this function
	var refDateTime time.Time
	var startDateTime time.Time
	var endDateTime time.Time
	var lookBackDuration time.Duration
	showCompleteWeeksOnly := viper.GetBool("completeWeeks")
	// first, get the date to start looking back from (this value should
	// have been passed in on the command-line, but defaults to the empty string)
	referenceDate := viper.Get("referenceDate").(string)
	// next, look for the "lookbackTime" value that we should use (this value can
	// be passed in on the command-line, but defaults to the empty string)
	lookBackStr := viper.Get("lookbackTime").(string)
	if referenceDate != "" {
		dateTime, err := time.Parse("2006-01-02", referenceDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: unable to parse end date '%s'; expected format is '2006-01-02'\n", referenceDate)
			os.Exit(-1)
		}
		refDateTime = dateTime
	} else {
		// If here, then no end-date was specified, so choose a default value of
		// of the current day at midnight (UTC) and make that the ending date time
		// for our query
		refDateTime = time.Now().UTC().Truncate(time.Hour * 24)
	}
	// if a lookback time was specified, then grab it
	if lookBackStr != "" {
		// get the lookback duration
		lookBackDuration = getLookbackDuration(lookBackStr)
	}
	// if the user has requested only complete weeks, then we need to shift our reference
	// date to the start of the week we're interested in
	if showCompleteWeeksOnly {
		// if a lookback time was provided and it is less than zero, or if a reference
		// date was provided but a loopback time was not, then we're going to be looking
		// ahead rather than looking back and we need to shift the reference date one
		// more week into the future to only include complete weeks in our queries
		if (lookBackStr != "" && lookBackDuration < 0) || (lookBackStr == "" && referenceDate != "") {
			// if the offset between the weekday for the input date and the previous
			// (i.e. the start of the week) in days is not zero, then we need to shift
			// the reference date forward by one day so that the lookback logic will work
			// for a look ahead time window
			offset := (int(time.Monday) - int(refDateTime.Weekday()) - 7) % 7
			if offset != 0 {
				refDateTime = refDateTime.Add(7 * 24 * time.Hour)
			}
		}
		// now, shift the reference date to the start of the week we're interested in
		refDateTime = weekStartDate(refDateTime)
		fmt.Fprintf(os.Stderr, "WARN: only complete weeks requested, reference date set to '%s'\n", refDateTime.Format("2006-01-02"))
	}
	// now that we have the reference date set appropriately, use that and the lookback
	// time (if it was set) to setup our time window
	if lookBackStr != "" {
		// if a negative lookback time was specified, then the start of our query window is the
		// reference date time and the end of our query window is the absolute value of that lookback
		// time added to the start date time
		if lookBackDuration < 0 {
			startDateTime = refDateTime
			refDateTime = refDateTime.Add(-lookBackDuration)
			if showCompleteWeeksOnly {
				// since it's an end date for the window, we just need to truncate so that we only
				// see data from complete weeks in our output
				refDateTime = weekStartDate(refDateTime)
				fmt.Fprintf(os.Stderr, "WARN: only complete weeks requested, end date set to '%s'\n", refDateTime.Format("2006-01-02"))
			}
		} else {
			// otherwise, subtract the lookback time from the reference date time to get the
			// start of our query window
			startDateTime = refDateTime.Add(-lookBackDuration)
			if showCompleteWeeksOnly {
				// if the start date is not the start of the week, then we need to shift
				// by a week and truncate to the start of the week to ensure we only get
				// complete weeks in our output data
				offset := (int(time.Monday) - int(startDateTime.Weekday()) - 7) % 7
				if offset != 0 {
					startDateTime = weekStartDate(startDateTime.Add(7 * 24 * time.Hour))
					fmt.Fprintf(os.Stderr, "WARN: only complete weeks requested, start date set to '%s'\n", startDateTime.Format("2006-01-02"))
				}
			}
		}
		endDateTime = refDateTime
	} else {
		// if a lookback time was not specified, but a reference time was, then the start
		// of our query window is the reference date time and the end of our query window
		// is the curren date time
		if referenceDate != "" {
			startDateTime = refDateTime
			endDateTime = time.Now().UTC().Truncate(time.Hour * 24)
			if showCompleteWeeksOnly {
				// since it's an end date for the window, we just need to truncate so that we only
				// see data from complete weeks in our output
				endDateTime = weekStartDate(endDateTime)
				fmt.Fprintf(os.Stderr, "WARN: only complete weeks requested, end date set to '%s'\n", endDateTime.Format("2006-01-02"))
			}
			fmt.Fprintf(os.Stderr, "WARN: no lookback time specified; using reference date as start of time window\n")
		} else {
			// otherwise, if neither a lookback time nor a reference time was specified, then
			// assume a default lookback time of 90 days from the current date time
			fmt.Fprintf(os.Stderr, "WARN: no lookback time or reference date specified; using default lookback time of 90 days\n")
			startDateTime = refDateTime.Add(-defaultLookbackDays * 24 * time.Hour)
			if showCompleteWeeksOnly {
				// since it's a start date for the window, we need to shift by a week and
				// truncate to the start of the week to ensure we only get complete weeks
				// in our output data
				startDateTime = weekStartDate(startDateTime.Add(7 * 24 * time.Hour))
				fmt.Fprintf(os.Stderr, "WARN: only complete weeks requested, start date set to '%s'\n", startDateTime.Format("2006-01-02"))
			}
			endDateTime = refDateTime
		}
	}
	// if the start time for our query window is in the future, we should exit with an error
	// since no data will be available
	currentDateTime := time.Now().UTC()
	if startDateTime.After(currentDateTime) {
		fmt.Fprintf(os.Stderr, "ERROR: defined start date for query window is in the future; no data will be available\n")
		os.Exit(-1)
	} else if endDateTime.After(currentDateTime) {
		// if the end time for our query window is in the future, then we should warn the user
		fmt.Fprintf(os.Stderr, "WARN: defined end date for query window is in the future; results only cover %s through %s\n", startDateTime.Format("2006-01-02"), currentDateTime.Format("2006-01-02"))
	}
	fmt.Fprintf(os.Stderr, "INFO: time window for query is %s through %s\n", startDateTime.Format("2006-01-02"), endDateTime.Format("2006-01-02"))
	// otherwise, return the start and end date times for our query window
	return githubv4.DateTime{startDateTime}, githubv4.DateTime{refDateTime}
}
