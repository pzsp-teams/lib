package api

import (
	"context"

	msmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	"golang.org/x/sync/errgroup"

	"github.com/pzsp-teams/lib/internal/sender"
	"github.com/pzsp-teams/lib/search"
)

type entityFilter func(SearchEntity) bool

type messageFetcher func(ctx context.Context, e SearchEntity) (msmodels.ChatMessageable, *sender.RequestError)

func cloneSearchOpts(in *search.SearchMessagesOptions) *search.SearchMessagesOptions {
	if in == nil {
		return &search.SearchMessagesOptions{}
	}

	out := *in
	out.From = append([]string(nil), in.From...)
	out.NotFrom = append([]string(nil), in.NotFrom...)
	out.To = append([]string(nil), in.To...)
	out.NotTo = append([]string(nil), in.NotTo...)

	if in.SearchPage != nil {
		sp := *in.SearchPage
		out.SearchPage = &sp
	}

	return &out
}

type task struct {
	idx int
	e   SearchEntity
}

func enrichMessages(
	ctx context.Context,
	searchAPI SearchAPI,
	opts *search.SearchMessagesOptions,
	keep entityFilter,
	fetch messageFetcher,
) ([]*SearchMessage, *sender.RequestError, *int32) {
	localOpts := cloneSearchOpts(opts)

	resp, reqErr := searchAPI.SearchMessages(ctx, localOpts)
	if reqErr != nil {
		return nil, reqErr, nil
	}

	entities := extractMessages(resp)
	if len(entities) == 0 {
		return []*SearchMessage{}, nil, nil
	}

	tasks := prepareTasks(entities, keep)

	if len(tasks) == 0 {
		nextFrom := calcNextSearchFrom(localOpts, len(entities))
		return []*SearchMessage{}, nil, nextFrom
	}

	const maxConcurrent = -1

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrent)
	results := make([]*SearchMessage, len(entities))

	for _, t := range tasks {
		g.Go(func() error {
			msg, err := fetch(gctx, t.e)
			if err != nil {
				if err.StatusCode() == 404 {
					return nil
				}
				return err
			}
			results[t.idx] = &SearchMessage{
				Message:   msg,
				ChannelID: t.e.ChannelID,
				TeamID:    t.e.TeamID,
				ChatID:    t.e.ChatID,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		if reqErr, ok := err.(*sender.RequestError); ok {
			return nil, reqErr, nil
		}
		return nil, &sender.RequestError{Message: err.Error()}, nil
	}

	nextFrom := calcNextSearchFrom(localOpts, len(entities))
	return results, nil, nextFrom
}

func prepareTasks(entities []SearchEntity, keep entityFilter) []task {
	tasks := make([]task, 0, len(entities))
	seen := make(map[string]struct{}, len(entities))

	for i, e := range entities {
		if e.MessageID == nil {
			continue
		}
		if keep != nil && !keep(e) {
			continue
		}
		id := *e.MessageID
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		tasks = append(tasks, task{idx: i, e: e})
	}
	return tasks
}
