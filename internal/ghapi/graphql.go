package ghapi

import (
	"context"

	"github.com/shurcooL/githubv4"
)

// PRDetail is the aggregated PR view fetched in ONE GraphQL query — the whole
// point of the GraphQL path: reviews + review threads + checks + labels +
// assignees together, instead of many REST round-trips.
type PRDetail struct {
	Number    int
	Title     string
	State     string
	Body      string
	Author    string
	Labels    []string
	Assignees []string
	Reviews   []PRDetailReview
	Threads   []PRDetailThread
	Checks    []PRDetailCheck
}

type PRDetailReview struct {
	Author string
	State  string
	Body   string
}

// PRDetailThread is a review-comment thread with its resolve state and GraphQL
// node ID (needed for the resolve/unresolve mutations — REST has no equivalent).
type PRDetailThread struct {
	ID         string
	Path       string
	Line       int
	IsResolved bool
	Comments   []PRDetailThreadComment
}

type PRDetailThreadComment struct {
	Author string
	Body   string
}

type PRDetailCheck struct {
	Name       string
	Status     string
	Conclusion string
}

// PRDetailGraphQL fetches everything about a PR in a single GraphQL query.
func (c *Client) PRDetailGraphQL(ctx context.Context, owner, repo string, number int) (*PRDetail, error) {
	var q struct {
		Repository struct {
			PullRequest struct {
				Number int
				Title  string
				State  string
				Body   string
				Author struct {
					Login string
				}
				Labels struct {
					Nodes []struct{ Name string }
				} `graphql:"labels(first: 50)"`
				Assignees struct {
					Nodes []struct{ Login string }
				} `graphql:"assignees(first: 20)"`
				Reviews struct {
					Nodes []struct {
						Author struct{ Login string }
						State  string
						Body   string
					}
				} `graphql:"reviews(first: 50)"`
				ReviewThreads struct {
					Nodes []struct {
						ID         string
						IsResolved bool
						Comments   struct {
							Nodes []struct {
								Author struct{ Login string }
								Body   string
								Path   string
								Line   int
							}
						} `graphql:"comments(first: 50)"`
					}
				} `graphql:"reviewThreads(first: 50)"`
				Commits struct {
					Nodes []struct {
						Commit struct {
							StatusCheckRollup struct {
								Contexts struct {
									Nodes []struct {
										CheckRun struct {
											Name       string
											Status     string
											Conclusion string
										} `graphql:"... on CheckRun"`
										StatusContext struct {
											Context string
											State   string
										} `graphql:"... on StatusContext"`
									}
								} `graphql:"contexts(first: 50)"`
							}
						}
					}
				} `graphql:"commits(last: 1)"`
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	vars := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"repo":   githubv4.String(repo),
		"number": githubv4.Int(number),
	}
	if err := c.gql.Query(ctx, &q, vars); err != nil {
		return nil, err
	}

	pr := q.Repository.PullRequest
	out := &PRDetail{
		Number: pr.Number,
		Title:  pr.Title,
		State:  pr.State,
		Body:   pr.Body,
		Author: pr.Author.Login,
	}
	for _, l := range pr.Labels.Nodes {
		out.Labels = append(out.Labels, l.Name)
	}
	for _, a := range pr.Assignees.Nodes {
		out.Assignees = append(out.Assignees, a.Login)
	}
	for _, r := range pr.Reviews.Nodes {
		out.Reviews = append(out.Reviews, PRDetailReview{Author: r.Author.Login, State: r.State, Body: r.Body})
	}
	for _, t := range pr.ReviewThreads.Nodes {
		th := PRDetailThread{ID: t.ID, IsResolved: t.IsResolved}
		for i, cm := range t.Comments.Nodes {
			if i == 0 {
				th.Path = cm.Path
				th.Line = cm.Line
			}
			th.Comments = append(th.Comments, PRDetailThreadComment{Author: cm.Author.Login, Body: cm.Body})
		}
		out.Threads = append(out.Threads, th)
	}
	if len(pr.Commits.Nodes) > 0 {
		for _, ctxNode := range pr.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes {
			if ctxNode.CheckRun.Name != "" {
				out.Checks = append(out.Checks, PRDetailCheck{
					Name: ctxNode.CheckRun.Name, Status: ctxNode.CheckRun.Status, Conclusion: ctxNode.CheckRun.Conclusion,
				})
			} else if ctxNode.StatusContext.Context != "" {
				out.Checks = append(out.Checks, PRDetailCheck{
					Name: ctxNode.StatusContext.Context, Status: "COMPLETED", Conclusion: ctxNode.StatusContext.State,
				})
			}
		}
	}
	return out, nil
}

// ResolveThread marks a review thread resolved (GraphQL-only; no REST equivalent).
func (c *Client) ResolveThread(ctx context.Context, threadID string) error {
	var m struct {
		ResolveReviewThread struct {
			Thread struct{ ID string }
		} `graphql:"resolveReviewThread(input: $input)"`
	}
	input := githubv4.ResolveReviewThreadInput{ThreadID: githubv4.ID(threadID)}
	return c.gql.Mutate(ctx, &m, input, nil)
}

// UnresolveThread reopens a resolved review thread.
func (c *Client) UnresolveThread(ctx context.Context, threadID string) error {
	var m struct {
		UnresolveReviewThread struct {
			Thread struct{ ID string }
		} `graphql:"unresolveReviewThread(input: $input)"`
	}
	input := githubv4.UnresolveReviewThreadInput{ThreadID: githubv4.ID(threadID)}
	return c.gql.Mutate(ctx, &m, input, nil)
}
