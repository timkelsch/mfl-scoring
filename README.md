## Scoring Links
Stage: https://3xi97uokw5.execute-api.us-east-1.amazonaws.com/stage/mfl-scoring </br>
Prod: https://3xi97uokw5.execute-api.us-east-1.amazonaws.com/prod/mfl-scoring

## My Fantasy League Custom Fantasy Football Scoring

Provides a custom scoring solution to calculate live league championship scoring based on the following rules:
- Points are awarded for "total fantasy points scored" and "head to head record".
- The team with the most fantasy points gets [# of teams] points. The team with the next best record gets [(# of teams) - 1] points, and so on until all teams' points have been allocated.
- The same logic applies to calculating score based on fantasy points.
- If two or more teams have the same record or total fantasy points, tied teams equally share the points that would have been allocated to those places. For example, if there are 10 teams and the top three teams have the same record, each team would receive 9 points (10 points for best record + 9 points for second best record + 8 points for the third best record, divided by three teams = 9 points)
- If there is a tie in total championship points, there are currently two tiebreakers:
  - First, the team with the greatest total fantasy points wins
  - Second, the team with the better AllPlay percentage wins (see AllPlay explanation below)
- The team with the most total championship points wins.

What is AllPlay percentage? 

First, we'll start with explaining what AllPlay record is: AllPlay is calculated by comparing each team's fantasy points to every other team's fantasy points each week of the season. An AllPlay win is accumulated when your team has a greater number of fantasy points than another team on any given week of the season. The same goes for ties and losses. The AllPlay records for the year are calculated across all weeks of the season. 

Example: if my team was tied for the 4th most points in week 1 of this season, my AllPlay record would be 5 wins (I had more points than teams with 6th to 10th most points), 3 losses (I had less points than teams with 1st to 3rd most points), and 1 tie (I had the same number of points as one team).

AllPlay percentage for each team is calculated as follows:
(1 * AllPlay wins) + (0.5 * AllPlay Ties) / (AllPlay wins + AllPlay ties + AllPlay losses)


## Disclaimer
I am not responsible for creation of this rule set - I merely automated the calculation of it.


## To Do:
- [ ] Change front end API from stage to prod and make mflstage.<domain> into stage
- [ ] Put team names in front of staging stage
- [ ] Write unit test(s) for tie breaker sorting
- [ ] Use json files for unit tests
- [ ] Get test coverage above 80%
- [-] Docker container move from alpine to scratch (tried and didn't reduce image size)
- [x] Make GET requests to MFL in parallel to improve performance
- [x] Wait until API responds to render front end
- [x] Build minimal front end
- [x] Finish implementing sorting for tie breakers
- [x] Figure out how to get alerts when prod goes down
- [x] Use for range where possible
- [x] Consolidate standard and AllPlay w/l/t into single fields
- [x] Enable JSON output based on queryStringParameters
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
- [x] Scrape AllPlay data from front end instead of schedule API
- [x] Standardize decimal format per column
- [x] Horizontally center column values
- [x] Add golangci-lint
- [x] Fix manual formatting - use github.com/jedib0t/go-pretty/v6/table to automate


## Requirements

* AWS CLI already configured with Administrator permission
* [Docker installed](https://www.docker.com/community-edition)
* [Golang](https://golang.org)


### Testing

```shell
make test
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
