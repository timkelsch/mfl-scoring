## To Do:
- [ ] Finish unit tests
- [ ] Find a better way of dealing with int/float issue in structs
- [ ] Consolidate standard and allplay w/l/t into single fields
- [ ] Build minimal front end
- [ ] Add year selector on front end that defaults to current year but allows checking scores for earlier years
- [x] Dockerize this app  
- [x] Add points tie breaker
- [x] Implement golangci-lint
- [x] Automate testing
- [x] Finalize tie break rules (1st tiebreak is fantasy points, 2nd is all-play %)
- [x] Rip out schedule API
- [x] Update from 2022 to 2023 league year
- [x] Allow teams with tied fantasy points to share championship points
- [x] Update API Key
- [x] Migrate runtime from go1.x to provided.al2
- [x] Move API key to secretsmanager protected by KMS
- [x] Implement tie break with all-play record percentage
- [x] Scrape allplay data from front end instead of schedule API
- [x] Standardize decimal format per column
- [x] Horizontally center column values
- [x] Add golangci-lint
- [x] Fix manual formatting - use github.com/jedib0t/go-pretty/v6/table to automate

## Requirements

* AWS CLI already configured with Administrator permission
* [Docker installed](https://www.docker.com/community-edition)
* [Golang](https://golang.org)


### Testing

We use `testing` package that is built-in in Golang and you can simply run the following command to run our tests:

```shell
go test -v ./hello-world/
```
# Appendix

### Golang installation

Please ensure Go 1.x (where 'x' is the latest version) is installed as per the instructions on the official golang website: https://golang.org/doc/install

A quickstart way would be to use Homebrew, chocolatey or your linux package manager.

#### Homebrew (Mac)

Issue the following command from the terminal:

```shell
brew install golang
```

If it's already installed, run the following command to ensure it's the latest version:

```shell
brew update
brew upgrade golang
```
