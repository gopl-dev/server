# wip

for simplicity, I keep todo list and progress here for now

- [ ] books
  - [x] create book
  - [x] update book
  - [x] review changes: accept|reject
  - [x] view book
  - [ ] delete book
  - [x] list books
  - [x] Add topics to entities
  - [x] Strict format for release date
  - [x] Rework book to have multiple authors
  - [x] Remove visibility from book

- [x] Minimal admin dashboard for current needs:
  - [x] List of new books to review
  - [x] List of edit requests for books

- [ ] Change requests 
    - [x] Review, apply, and test complex types (pictures — need preview; topics — references to another entity; authors — array: what was added, removed, or updated)
    - [ ] In the public activity log, let everyone see what was changed (at the moment the changes are applied)
    - [ ] Learn how to create a patch (when a typo is fixed within a large text, there’s no need to display the whole diff, but only the area around the fix)

- [ ] Maybe group services by entity not by action (Ask opinions)
- [ ] Review “retrieve one” methods with clear naming: Get{Something} should return {Something} or an error; Find{Something} may return nil, nil.

**Before release**:
- [ ] Review "TODO!"s
- [ ] enable "unused" linter 
- [ ] Homepage
- [ ] Licence
- [ ] RELEASE
---

AFTER RELEASE CHECKLIST:
- [ ] Setup & test email
- [ ] Setup & test auth & registration with google
- [ ] Setup & test auth & registration with github
- [ ] Setup & test tracing

NEXT:
- [ ] Resend verification link (Sometimes email is lost in somewhere between the woods (Not mailman to blame))
- [ ] Create new topic when creating/editing entity
- [ ] Pages
- [ ] Convert all TODO's into tasks/issues 
- [ ] Create a CLI command to set up a new dev environment
- [ ] Jobs
- [ ] Events
- [ ] Software
- [ ] Improve user profile
- [ ] Internal notifications
- [ ] Entity comments
- [ ] Entity likes
- [ ] Offline mode
- [ ] Make one instance of input validation for frontend and backend
- [ ] If user could see his connected AOuth accounts, connect other accounts and disconnect them
- [ ] Rework workers so it is possible to display a list of workers, enable/disable them, see run status, last run time, logs, etc.
- [ ] Tests for frontend
- [ ] Render initial data on the backend (for example, when the books page is requested, render it fully and return it, instead of letting the frontend fetch data via the API).
- [ ] Add server version to frontend
- [ ] Plain theme
- [ ] Notification system for selecting notification types and delivery channels (web, email, Telegram, etc.)
- [ ] Selective approve changes (user might reject some properties from request and apply some)
- [ ] Let user continue work on reject entity and proposed changes
- [ ] Review "delete account" test. Right now, it passes even if models belonging to the user still exist.
