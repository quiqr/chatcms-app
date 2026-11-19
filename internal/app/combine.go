package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/quiqr/chatcms-app/internal/aigateway"
	"github.com/quiqr/chatcms-app/internal/httpapi"
	"github.com/quiqr/chatcms-app/internal/policy"
	"github.com/quiqr/chatcms-app/internal/sitesync"
	"github.com/quiqr/chatcms-app/internal/store"
)

type aiCombiner struct {
	gateway interface {
		Ask(ctx context.Context, caller aigateway.Caller, messages []aigateway.Message) (string, error)
	}
	queries *store.Queries
}

const combineInstructions = `You combine two edits of the same field of a web page, made by two people at the same time.
Keep what each of them meant: every fact, change and addition from both, in the language they wrote in.
When they contradict each other, prefer the second version's facts but keep anything the first added that does not conflict.
Keep the form of the field: a short title stays a short title, Markdown stays Markdown.
Answer with the combined text only, with no quotes, no explanation and no code fences.`

func combinePrompt(page, field, mine, theirs string) string {
	return fmt.Sprintf("Page: %s\nField: %s\n\nFirst version:\n%s\n\nSecond version:\n%s", page, field, mine, theirs)
}

func tidyCombined(said string) string {
	text := strings.TrimSpace(said)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if newline := strings.IndexByte(text, '\n'); newline >= 0 {
			text = text[newline+1:]
		}
		text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	}
	return strings.TrimSpace(text)
}

var refusedCouldNotCombine = &httpapi.RefusedError{
	Key:    "refused.could-not-combine",
	Reason: "the assistant could not combine these; choose one of the two",
}

func (c aiCombiner) Combine(ctx context.Context, actor policy.Actor, projectID uuid.UUID, page, field, mine, theirs string) (string, error) {
	person, err := uuid.Parse(actor.PersonID)
	if err != nil {
		return "", httpapi.ErrNotAllowed
	}
	row, err := c.queries.GetProject(ctx, pgUUID(projectID))
	if err != nil {
		return "", err
	}
	caller := aigateway.Caller{Person: person, Project: projectID, Organisation: uuid.UUID(row.OrganisationID.Bytes)}
	said, err := c.gateway.Ask(ctx, caller, []aigateway.Message{
		{Role: "system", Content: combineInstructions},
		{Role: "user", Content: combinePrompt(page, field, mine, theirs)},
	})
	switch {
	case errors.Is(err, aigateway.ErrLimited):
		return "", &httpapi.RefusedError{Key: "refused.ai-limit-reached", Reason: err.Error()}
	case errors.Is(err, aigateway.ErrNotEnabled):
		return "", &httpapi.RefusedError{Key: "refused.assistant-not-available", Reason: "the assistant is not available on this installation"}
	case err != nil:
		return "", refusedCouldNotCombine
	}
	combined := tidyCombined(said)
	if combined == "" {
		return "", refusedCouldNotCombine
	}
	return combined, nil
}

var _ sitesync.Combiner = aiCombiner{}
