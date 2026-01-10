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

type enrichResult struct {
	idx int
	e   SearchEntity
	msg msmodels.ChatMessageable
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

	const maxConcurrent = 16
	sem := make(chan struct{}, maxConcurrent)

	g, gctx := errgroup.WithContext(ctx)
	outCh := make(chan enrichResult, len(tasks))

	for _, t := range tasks {
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			msg, err := fetch(gctx, t.e)
			if err != nil {
				if err.StatusCode() == 404 {
					return nil
				}
				return err
			}

			select {
			case outCh <- enrichResult{idx: t.idx, e: t.e, msg: msg}:
				return nil
			case <-gctx.Done():
				return gctx.Err()
			}
		})
	}

	go func() {
		_ = g.Wait()
		close(outCh)
	}()

	byIdx := make(map[int]enrichResult, len(tasks))
	for r := range outCh {
		byIdx[r.idx] = r
	}

	if err := g.Wait(); err != nil {
		if reqErr, ok := err.(*sender.RequestError); ok {
			return nil, reqErr, nil
		}
		return nil, &sender.RequestError{Message: err.Error()}, nil
	}

	results := aggregateResults(entities, byIdx)

	nextFrom := calcNextSearchFrom(localOpts, len(entities))
	return results, nil, nextFrom
}

func aggregateResults(entities []SearchEntity, byIdx map[int]enrichResult) []*SearchMessage {
	results := make([]*SearchMessage, 0, len(byIdx))
	for i := range entities {
		r, ok := byIdx[i]
		if !ok {
			continue
		}
		results = append(results, &SearchMessage{
			Message:   r.msg,
			ChannelID: r.e.ChannelID,
			TeamID:    r.e.TeamID,
			ChatID:    r.e.ChatID,
		})
	}
	return results
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
