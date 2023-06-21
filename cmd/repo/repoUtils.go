/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package repo

import (
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/tjmcs/get-gh-info/cmd"
	"github.com/tjmcs/get-gh-info/utils"
)

type IssueOrPullRequest interface {
	*Issue | *PullRequest
	IsClosed() bool
	GetCreatedAt() githubv4.DateTime
	GetClosedAt() githubv4.DateTime
	GetComments() cmd.Comments
}

func IsClosed[C IssueOrPullRequest](contrib C) bool {
	return contrib.IsClosed()
}

/*
 * Define a generic function that we can use to get the time of the first response to an issue
 * or pull request. The arguments to this function are as follows:
 *
 *   - contrib: the issue or pull request for which we want to get the time of the first response
 *   - endDateTime: the end date/time of the query window; this is used to determine the default
 *         time to first response if no response is found
 *   - fromTeamOnly: a boolean flag that indicates whether or not we should only count comments
 *         from immediate team members
 *   - teamIds: a slice of strings that contains the GitHub IDs of the members of the team that
 *         owns the repository that contains the issue or pull request for which we want to get
 *         the time of the first response
 *
 */
func GetFirstResponseTime[C IssueOrPullRequest](contrib C, endDateTime githubv4.DateTime, fromTeamOnly bool, teamIds []string) time.Duration {
	// define a variable to hold the time of the first response
	var firstRespTime time.Duration
	// grab the time that this contribution created, the time when it was was closed
	// and the flag indicating whether or not it actually was closed
	contribCreatedAt := contrib.GetCreatedAt()
	contribClosedAt := contrib.GetClosedAt()
	contribIsClosed := contrib.IsClosed()
	// then use those values to set a default "first response time" based on the
	// difference between the time that the contribution was closed (if it was closed
	// before the end of our query window) or the end of our query window (if it was
	// not) and the creation time for this contrib
	if contribIsClosed && contribClosedAt.Time.Before(endDateTime.Time) {
		firstRespTime = contribClosedAt.Time.Sub(contribCreatedAt.Time)
	} else {
		firstRespTime = endDateTime.Time.Sub(contribCreatedAt.Time)
	}
	// and get the comments for this contrib
	comments := contrib.GetComments()
	// if no comments were found for this contrib, then use the default time we just defined
	if len(comments.Nodes) == 0 {
		return firstRespTime
	}
	// otherwise, loop over the comments for this contrib, looking for the first comment from
	// a team member (note that it is assumed here that the comments are sorted in ascending
	// order by the time they were last updated)
	for _, comment := range comments.Nodes {
		// if the comment was created after the contrib was closed (if it is closed) or the
		// comment was created after the end of our query window (if it is not closed),
		// then we've reached the end of the time where a user could have responded within
		// our time window, so we should break out of the loop and just use the default
		// which we defined (above)
		if (contribIsClosed && comment.CreatedAt.After(contribClosedAt.Time)) ||
			comment.CreatedAt.After(endDateTime.Time) {
			break
		}
		// if the comment has an author (it should)
		if len(comment.Author.Login) > 0 {
			// if the flag to only count comments from the immediate team was
			// set, then only count comments from immediate team members
			if fromTeamOnly {
				// if here, looking only for comments only from immediate team members,
				// so if this comment is not from an immediate team member skip it
				idx := utils.FindIndexOf(comment.Author.Login, teamIds)
				if idx < 0 {
					continue
				}
			} else {
				// otherwise (by default), we're looking for comments from anyone who is
				// an owner of this repository, a member of the organization that owns this
				// repository, or collaborator on this repository; if that's not the case
				// for this comment, then skip it
				if comment.AuthorAssociation != "OWNER" &&
					comment.AuthorAssociation != "MEMBER" &&
					comment.AuthorAssociation != "COLLABORATOR" {
					continue
				}
			}
			// if get here, then we've found a comment from a member of the team that was
			// created before the end of our query window, so calculate the time to first
			// response and break out of the loop
			firstRespTime = comment.CreatedAt.Time.Sub(contribCreatedAt.Time)
			break
		}
	}
	// and return the time of the first response that we found (or the default if we
	// didn't find one)
	return firstRespTime
}

/*
 * Define a generic function that we can use to get the time of the latest response (or staleness)
 * to an issue or pull request. The arguments to this function are as follows:
 *
 *   - contrib: the issue or pull request for which we want to get the time of the first response
 *   - endDateTime: the end date/time of the query window; this is used to determine the default
 *         time to first response if no response is found
 *   - fromTeamOnly: a boolean flag that indicates whether or not we should only count comments
 *         from immediate team members
 *   - teamIds: a slice of strings that contains the GitHub IDs of the members of the team that
 *         owns the repository that contains the issue or pull request for which we want to get
 *         the time of the first response
 *
 */
func GetLatestResponseTime[C IssueOrPullRequest](contrib C, endDateTime githubv4.DateTime, fromTeamOnly bool, teamIds []string) time.Duration {
	// grab a few values from this contribution that we'll need later
	contribCreatedAt := contrib.GetCreatedAt()
	contribIsClosed := contrib.IsClosed()
	contribClosedAt := contrib.GetClosedAt()
	// next, set a default "latest response time" based on the difference either the time
	// that the contrib was closed (if it's closed) or the end of our query window and the
	// creation time for this contrib; if no response is found then this will be the
	// staleness time that we return
	var stalenessTime time.Duration
	if contribIsClosed && contribClosedAt.Before(endDateTime.Time) {
		stalenessTime = contribClosedAt.Sub(contribCreatedAt.Time)
	} else {
		stalenessTime = endDateTime.Sub(contribCreatedAt.Time)
	}
	// and get the comments for this contrib
	comments := contrib.GetComments()
	// if no comments were found for this contrib, then return the default staleness time
	if len(comments.Nodes) == 0 {
		return stalenessTime
	}
	// loop over the comments for this contrib, looking for the latest comment from a team member
	// (note that it is assumed here that the comments are sorted in descending order by the time
	// they were last updated)
	for _, comment := range comments.Nodes {
		// if this comment was created after the time when the contrib was closed
		//  or the contrib is not closed and the comment was created after the
		// reference time time, then skip it
		if (contribIsClosed && comment.CreatedAt.After(contribClosedAt.Time)) ||
			comment.CreatedAt.After(endDateTime.Time) {
			continue
		}
		// if the comment has an author (it should)
		if len(comment.Author.Login) > 0 {
			// if the flag to only count comments from the immediate team was
			// set, then only count comments from immediate team members
			if fromTeamOnly {
				// if here, looking only for comments only from immediate team members,
				// so if this comment is not from an immediate team member skip it
				idx := utils.FindIndexOf(comment.Author.Login, teamIds)
				if idx < 0 {
					continue
				}
			} else {
				// otherwise (by default), we're looking for comments from anyone who is
				// an owner of this repository, a member of the organization that owns this
				// repository, or collaborator on this repository; if that's not the case
				// for this comment, then skip it
				if comment.AuthorAssociation != "OWNER" &&
					comment.AuthorAssociation != "MEMBER" &&
					comment.AuthorAssociation != "COLLABORATOR" {
					continue
				}
			}
			// if get here, then we've found a comment from a member of the team,
			// so use the time the comment was closed or the end time of our query window
			// (whichever is less) to calculate a staleness value for this contrib
			if contribIsClosed && contribClosedAt.Before(endDateTime.Time) {
				// if the contrib is closed before the end time of our time window, then use
				// the time the contrib was closed to determine the staleness time
				stalenessTime = contribClosedAt.Time.Sub(comment.CreatedAt.Time)
			} else {
				// otherwise use the reference time for our time window
				stalenessTime = endDateTime.Sub(comment.CreatedAt.Time)
			}
			break
		}
	}
	// and return the time of the latest response that we found (or the default if we
	// didn't find one)
	return stalenessTime
}
