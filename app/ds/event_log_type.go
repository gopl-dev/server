package ds

// EventLogType defines a stable identifier of a system or user event.
type EventLogType string

const (
	// EventLogSystem represents a system-level event that is not directly
	// initiated by a user action (e.g. deployment, maintenance, background jobs).
	EventLogSystem EventLogType = "system_event"

	// EventLogUserAccountCreated is recorded when a user account record
	// is created in the system (email/password or OAuth).
	EventLogUserAccountCreated EventLogType = "user_account_created"

	// EventLogUserEmailConfirmed is recorded when a user successfully
	// confirms ownership of their email address.
	EventLogUserEmailConfirmed EventLogType = "user_email_confirmed"

	// EventLogUserRequestPasswordReset is recorded when a user initiates
	// a password reset flow.
	EventLogUserRequestPasswordReset EventLogType = "user_request_password_reset"

	// EventLogUserPasswordChangedByResetRequest is recorded when a user
	// changes their password via a password reset request.
	EventLogUserPasswordChangedByResetRequest EventLogType = "user_password_changed_by_reset_request"

	// EventLogUserPasswordChanged is recorded when a user changes their
	// password while authenticated.
	EventLogUserPasswordChanged EventLogType = "user_password_changed"

	// EventLogUserEmailChangeRequested is recorded when a user initiates
	// an email address change flow.
	EventLogUserEmailChangeRequested EventLogType = "user_email_change_requested"

	// EventLogUserEmailChanged is recorded when a user's email address
	// has been successfully changed.
	EventLogUserEmailChanged EventLogType = "user_email_changed"

	// EventLogUserUsernameChanged is recorded when a user changes
	// their username.
	EventLogUserUsernameChanged EventLogType = "user_username_changed"

	// EventLogUserAccountActivated is recorded when a user account becomes
	// active and eligible to appear in public-facing activity feeds.
	EventLogUserAccountActivated EventLogType = "user_account_activated"

	// EventLogEntitySubmitted is recorded when an entity is submitted
	// for moderation or review.
	EventLogEntitySubmitted EventLogType = "entity_submitted"

	// EventLogEntityApproved is recorded when an entity is approved
	// by a moderator or administrator.
	EventLogEntityApproved EventLogType = "entity_approved"

	// EventLogEntityRejected is recorded when an entity submission
	// is rejected during moderation.
	EventLogEntityRejected EventLogType = "entity_rejected"

	// EventLogEntityAdded is recorded when an entity becomes visible
	// and available to other users (published/accepted).
	EventLogEntityAdded EventLogType = "entity_added"

	// EventLogEntityUpdated is recorded when an existing entity
	// is updated.
	EventLogEntityUpdated EventLogType = "entity_updated"
)

// EventLogTypes lists all supported event log types.
var EventLogTypes = []EventLogType{
	EventLogSystem,
	// User events
	EventLogUserAccountCreated,
	EventLogUserEmailConfirmed,
	EventLogUserAccountActivated,
	EventLogUserRequestPasswordReset,
	EventLogUserPasswordChangedByResetRequest,
	EventLogUserPasswordChanged,
	EventLogUserEmailChangeRequested,
	EventLogUserEmailChanged,
	EventLogUserUsernameChanged,
	// Entity events
	EventLogEntitySubmitted,
	EventLogEntityApproved,
	EventLogEntityRejected,
	EventLogEntityAdded,
	EventLogEntityUpdated,
}

// Verb returns a short, human-readable verb describing the event.
func (t EventLogType) Verb() string {
	switch t {
	case EventLogUserAccountCreated:
		return "created account"
	case EventLogUserRequestPasswordReset:
		return "password reset request"
	case EventLogUserPasswordChangedByResetRequest:
		return "changed password by reset request"
	case EventLogUserPasswordChanged:
		return "changed password changed"
	case EventLogUserEmailConfirmed:
		return "email confirmed"
	case EventLogUserEmailChangeRequested:
		return "email change requested"
	case EventLogUserEmailChanged:
		return "email changed"
	case EventLogUserUsernameChanged:
		return "username changed"
	case EventLogUserAccountActivated:
		return "joined"
	case EventLogEntitySubmitted:
		return "submitted"
	case EventLogEntityApproved:
		return "approved"
	case EventLogEntityRejected:
		return "rejected"
	case EventLogEntityAdded:
		return "created"
	case EventLogEntityUpdated:
		return "updated"
	}

	return ""
}
