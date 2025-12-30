# wip

for simplicity, I keep todo list and progress here for now

---
- [ ] core user features:
    - [x] register user
    - [x] confirm email
    - [x] login
    - [x] logout
    - [x] handle frontend logic for user auth status 
    - [x] change password
    - [x] reset password
    - [x] change email
    - [x] change username
    - [x] delete account
    - [x] login/register by google & github

- [x] user activity log
- [ ] books
  - [ ] create book
  - [ ] update book
  - [ ] review changes: accept|reject
  - [ ] view book
  - [ ] delete book
  - [ ] list books

- [x] workers boilerplate
  - [x] cleanup expired password change requests
  - [x] cleanup email change requests
  - [x] delete users with unconfirmed emails
  - [x] cleanup expired user session
  - [x] cleanup deleted user accounts
  
- [ ] Separate "web" handlers from "api" and rename to "view" endpoints
- [x] Validation and sanitation should be done at service layer
- [x] uuid v7 for IDs
- [ ] Review "TODO!"
- [ ] enable "unused" linter 
- [ ] RELEASE
---

Next features:
- [ ] Jobs
- [ ] Events
- [ ] Comments
- [ ] Likes
- [ ] Offline mode


Must to have
- [ ] Resend verification link (Sometimes email is lost in somewhere between the woods (Not mailman to blame))

Nice to have:
- [ ] If user could see his connected AOuth accounts, connect other accounts and disconnect them
