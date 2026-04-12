# 🚀 Serving gopl.dev

### Contributing

We don't have a formal set of rules for contributions yet; everyone is welcome! We appreciate everything from critiques and suggestions to bug fixes and new features.

### Setup Your Own Instance

See [SETUP.md](SETUP.md) for detailed instructions on how to set up your own instance.

### Internal Tools

* **Reset Dev Environment**
    ```bash
    go run ./cmd/cli/main.go rde
    ```  
  Resets the development environment by recreating the database, applying migrations, and creating a default user. This is useful during active development if you need a clean state.


* **Database Seeding**
    ```bash
    go run ./cmd/cli/main.go sd
    ```  
  Seeds data into the database. By default, it seeds all available data. You can specify an entity and a count:  
  `go run ./cmd/cli/main.go sd users 1000`

  Run `go run ./cmd/cli/main.go ? sd` to see all available options and detailed descriptions.

---

License [MIT](LICENSE)