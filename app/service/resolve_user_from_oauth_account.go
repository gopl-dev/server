package service

import (
	"context"
	"errors"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/oauth/provider"
	"github.com/gopl-dev/server/test/factory/random"
	"github.com/markbates/goth"
)

var resolveUserFromOAuthAccount = z.Shape{
	"Email":    emailInputRules,
	"Provider": z.String().Required(z.Message("provider is required")),
	"UserID":   z.String().Required(z.Message("user_id is required")),
}

// ResolveUserFromOAuthAccount attempts to find an existing user associated with an OAuth provider.
// 1. If the OAuth account exists, it returns the linked user.
// 2. If the OAuth account is missing but the email exists, it links the OAuth account to that user.
// 3. If neither exists, it creates a new user (with a unique username) and links the OAuth account.
func (s *Service) ResolveUserFromOAuthAccount(ctx context.Context, authAcc goth.User) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "CreateOAuthUserAccount")
	defer span.End()

	in := &AuthenticateOAuthUserInput{&authAcc}
	err = Normalize(in)
	if err != nil {
		return
	}

	acc, err := s.db.GetOAuthUserAccount(ctx, provider.New(in.Provider), in.UserID)
	if err == nil {
		return s.db.FindUserByID(ctx, acc.UserID)
	}
	if !errors.Is(err, repo.ErrOAuthUserAccountNotFound) {
		return
	}

	// create new oauth account
	user, err = s.FindUserByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, repo.ErrUserNotFound) {
		return nil, err
	}

	// create new user
	if errors.Is(err, repo.ErrUserNotFound) {
		username, err := s.selectUsernameForOAuthUser(ctx, in.NickName, in.Name, in.Email)
		if err != nil {
			return nil, err
		}

		user = &ds.User{
			ID:             ds.NewID(),
			Username:       username,
			Email:          in.Email,
			EmailConfirmed: true,
			Password:       random.String(32), //nolint:mnd
			CreatedAt:      time.Now(),
			UpdatedAt:      nil,
			DeletedAt:      nil,
		}

		err = s.db.CreateUser(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	acc = &ds.OAuthUserAccount{
		ID:             ds.NewID(),
		UserID:         user.ID,
		Provider:       provider.New(in.Provider),
		ProviderUserID: in.UserID,
		CreatedAt:      time.Now(),
	}

	err = s.CreateOAuthUserAccount(ctx, acc)
	if err != nil {
		return
	}

	return user, nil
}

// ResolveUserFromOAuthAccountInput ...
type ResolveUserFromOAuthAccountInput struct {
	*goth.User
}

// Sanitize ...
func (in *ResolveUserFromOAuthAccountInput) Sanitize() {
	in.UserID = strings.TrimSpace(in.UserID)
	in.Provider = strings.TrimSpace(in.Provider)
	in.Email = strings.TrimSpace(in.Email)
}

// Validate ...
func (in *ResolveUserFromOAuthAccountInput) Validate() error {
	return validateInput(resolveUserFromOAuthAccount, in)
}

// selectUsernameForOAuthUser picks the first available username from the provided candidates.
// If all candidates are taken, it appends random suffixes and retries until a unique name is found.
func (s *Service) selectUsernameForOAuthUser(ctx context.Context, maybeNames ...string) (username string, err error) {
	const atSign = "@"
	names := make([]string, 0, len(maybeNames))

	for _, n := range maybeNames {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}

		if ii := strings.Index(n, atSign); ii > -1 {
			n = n[:ii]
		}
		names = append(names, n)
	}

	if len(names) == 0 {
		names = append(names, random.String(6)) //nolint:mnd
	}

	// The first candidate that isn't found in the DB is returned immediately
	for {
		for _, name := range names {
			_, err := s.db.FindUserByUsername(ctx, name)
			if errors.Is(err, repo.ErrUserNotFound) {
				return name, nil
			}
			if err != nil {
				return "", err
			}
		}

		// none of provided usernames is available,
		// use random one (and only one)
		names = []string{names[0] + "-" + random.String(4)} //nolint:mnd
	}
}
