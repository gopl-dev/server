# wip

for simplicity, I keep todo list and progress here for now

- [ ] books
  - [x] create book
  - [x] update book
  - [ ] review changes: accept|reject
  - [x] view book
  - [ ] delete book
  - [x] list books
  - [x] Add topics to entities
  - [x] Strict format for release date
  - [x] Rework book to have multiple authors
  - [x] Remove visibility from book

- [ ] Minimal admin dashboard for current needs:
  - [x] List of new books to review
  - [ ] List of edit requests for books
  - [ ] List of edit requests for books
  - [ ] List activity log

- [x] Rework user_action_log and entity_change_log into event_log
- [ ] Maybe group services by entity not by action (Ask opinions)
- [ ] Review “retrieve one” methods with clear naming: Get{Something} should return {Something} or an error; Find{Something} may return nil, nil.
- [x] Make tools to seed data

**Before release**:
- [ ] Review "TODO!"s
- [ ] enable "unused" linter 
- [ ] Homepage
- [ ] RELEASE
---

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
