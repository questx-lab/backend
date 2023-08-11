package domain

import (
	"context"
	"regexp"
	"unicode"

	"github.com/fatih/structs"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func processValidationData(
	ctx context.Context, questFactory questclaim.Factory, includeSecret bool, quest *entity.Quest,
) error {
	processor, err := questFactory.LoadProcessor(ctx, includeSecret, *quest, quest.ValidationData)
	if err != nil {
		return err
	}

	quest.ValidationData = structs.Map(processor)
	return nil
}

func checkCommunityHandle(ctx context.Context, handle string) error {
	if len(handle) < 4 {
		return errorx.New(errorx.BadRequest, "Handle too short (at least 4 characters)")
	}

	if len(handle) > 32 {
		return errorx.New(errorx.BadRequest, "Handle too long (at most 32 characters)")
	}

	ok, err := regexp.MatchString("^[a-z0-9_]*$", handle)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot execute regex pattern: %v", err)
		return errorx.Unknown
	}

	if !ok {
		return errorx.New(errorx.BadRequest, "Name contains invalid characters")
	}

	return nil
}

func checkCommunityDisplayName(displayName string) error {
	if len(displayName) < 4 {
		return errorx.New(errorx.BadRequest, "Display name too short (at least 4 characters)")
	}

	return nil
}

func generateCommunityHandle(displayName string) string {
	handle := []rune{}
	for _, c := range displayName {
		if isAsciiLetter(c) {
			handle = append(handle, unicode.ToLower(c))
		} else if c == ' ' {
			handle = append(handle, '_')
		}
	}

	return string(handle)
}

func isAsciiLetter(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '_'
}

func checkUsername(ctx context.Context, userName string) error {
	if len(userName) < 4 {
		return errorx.New(errorx.BadRequest, "Username too short (at least 4 characters)")
	}

	if len(userName) > 32 {
		return errorx.New(errorx.BadRequest, "Username too long (at most 32 characters)")
	}

	ok, err := regexp.MatchString("^[A-Za-z0-9_]*$", userName)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot execute regex pattern: %v", err)
		return errorx.Unknown
	}

	if !ok {
		return errorx.New(errorx.BadRequest, "Name contains invalid characters")
	}

	return nil
}
