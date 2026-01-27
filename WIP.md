# wip

for simplicity, I keep todo list and progress here for now

- [ ] books
  - [x] create book
  - [x] update book
  - [ ] review changes: accept|reject
  - [ ] view book
  - [ ] delete book
  - [x] list books


- [ ] Add topics to entities
- [ ] Add price to book
- [ ] Maybe group services by entity not by action (Ask opinions)
- [ ] Review “retrieve one” methods with clear naming: Get{Something} should return {Something} or an error; Find{Something} may return nil, nil.
- [x] Make tools to seed data
- [ ] Review "TODO!"s
- [ ] enable "unused" linter 
- [ ] Frontpage
- [ ] RELEASE
---

NEXT:
- [ ] Convert all TODO's into tasks/issues 
- [ ] Create a CLI command to set up a new dev environment
- [ ] Jobs
- [ ] Events
- [ ] Improve user profile
- [ ] Internal notifications
- [ ] Entity comments
- [ ] Entity likes
- [ ] Offline mode


Must to have
- [ ] Resend verification link (Sometimes email is lost in somewhere between the woods (Not mailman to blame))

Nice to have:
- [ ] If user could see his connected AOuth accounts, connect other accounts and disconnect them
- [ ] Rework workers so it is possible to display a list of workers, enable/disable them, see run status, last run time, logs, etc.
- [ ] Tests for frontend
- [ ] Render initial data on the backend (for example, when the books page is requested, render it fully and return it, instead of letting the frontend fetch data via the API).
