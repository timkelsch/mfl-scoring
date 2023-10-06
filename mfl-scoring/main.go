package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/gocolly/colly"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type LeagueResponse struct {
	Version string `json:"version"`
	League  struct {
		Franchises struct {
			Count     string `json:"count"`
			Franchise []struct {
				Name      string `json:"name"`
				ID        string `json:"id"`
				OwnerName string `json:"owner_name"`
				Username  string `json:"username,omitempty"`
			} `json:"franchise"`
		} `json:"franchises"`
		ID      string `json:"id"`
		History struct {
			League []struct {
				URL  string `json:"url"`
				Year string `json:"year"`
			} `json:"league"`
		} `json:"history"`
		Name    string `json:"name"`
		H2H     string `json:"h2h"`
		BaseURL string `json:"baseURL"`
	} `json:"league"`
	Encoding string `json:"encoding"`
}

type LeagueStandingsResponse struct {
	Version         string `json:"version"`
	LeagueStandings struct {
		Franchise []struct {
			ID            string `json:"id"`
			RecordWins    string `json:"h2hw"`
			RecordLosses  string `json:"h2hl"`
			RecordTies    string `json:"h2ht"`
			PointsFor     string `json:"pf"`
			PointsAgainst string `json:"pa"`
		} `json:"franchise"`
	} `json:"leagueStandings"`
	Encoding string `json:"encoding"`
}

type Franchise struct {
	TeamID            string
	TeamName          string
	OwnerName         string
	RecordWins        int
	RecordLosses      int
	RecordTies        int
	PointsFor         float64
	PointsForString   string
	PointScore        float64
	PointScoreString  string
	RecordMagic       float64
	RecordScore       float64
	RecordScoreString string
	TotalScore        string
	AllPlayWins       int
	AllPlayLosses     int
	AllPlayTies       int
	AllPlayPercentage string
}

const (
	MflURL                  string = "https://www46.myfantasyleague.com/"
	LeagueYear              string = "2023"
	LeagueAPIQuery          string = "TYPE=league"
	LeagueStandingsAPIQuery string = "TYPE=leagueStandings"
	LeagueAPIPath           string = "export?"
	LeagueWebPath           string = "options?"
	PowerRankingsTableQuery string = "O=101"
	LeagueOutputSortQuery   string = "SORT=ALLPLAY"
	LeagueIDQuery           string = "L=15781"
	APIOutputTypeQuery      string = "JSON=1"
	APIKeySecretARN         string = "MflScoringApiKeySecret-x1mDJYYsWop9" //nolint:gosec // Not credentials
)

type Franchises []Franchise

type ByPointsFor struct{ Franchises }
type ByRecordMagic struct{ Franchises }
type ByTotalScore struct{ Franchises }

type Request = events.APIGatewayProxyRequest

type AllPlayTeamStats struct {
	FranchiseName     string
	AllPlayWins       string
	AllPlayLosses     string
	AllPlayTies       string
	AllPlayPercentage string
}

func (f Franchises) Len() int      { return len(f) }
func (f Franchises) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (f ByPointsFor) Less(i, j int) bool {
	return f.Franchises[i].PointsFor > f.Franchises[j].PointsFor
}

func (f ByRecordMagic) Less(i, j int) bool {
	return f.Franchises[i].RecordMagic > f.Franchises[j].RecordMagic
}

func (f ByTotalScore) Less(i, j int) bool {
	return f.Franchises[i].TotalScore > f.Franchises[j].TotalScore
}

func main() {
	lambda.Start(handler)
}

func handler(_ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	secretCache, err := secretcache.New()
	if err != nil {
		log.Fatal(err)
	}
	apiKey, err := secretCache.GetSecretString(APIKeySecretARN)
	if err != nil {
		log.Fatal(err)
	}

	franchiseDetails := getFranchiseDetails(apiKey)
	leagueStandings := getLeagueStandings(apiKey)

	checkResponseParity(franchiseDetails, leagueStandings)

	// Populate the slice of Franchise objects with league standing data
	franchisesWithStandings := associateStandingsWithFranchises(franchiseDetails, leagueStandings)

	// Put teams in order of most fantasy points scored
	sort.Sort(ByPointsFor{franchisesWithStandings})
	fmt.Printf("%+v \n", franchisesWithStandings)

	// Assign points to teams based on fantasy points scored, sharing points as necessary when teams tie
	calculatePointsScore(franchisesWithStandings)

	// Assign points to teams based on their record (1 point per win, 0.5 point per tie)
	// and sort by most points so we can get the tied teams next to each other
	calculateRecordMagic(franchisesWithStandings)
	sort.Sort(ByRecordMagic{franchisesWithStandings})

	// Assign points to teams based on their record. Teams with identical records share the points they collectively earned,
	// Ex: If there are two teams tied for the best record, both teams would receive 9.5 points
	// (10 points for first place + 9 points for second place, divided by 2 teams)
	calculateRecordScore(franchisesWithStandings)

	// Add up points assigned for fantasy points and points assigned for record
	calculateTotalScore(franchisesWithStandings)

	// Sort by TotalScore, then by all-play percentage (whatever that ends up being per support case from MFL)
	sort.Sort(ByTotalScore{franchisesWithStandings})

	allPlayTeamData := scrape()
	franchisesWithStandingsAndAllplay := appendAllPlay(franchisesWithStandings, allPlayTeamData)
	fmt.Print(franchisesWithStandingsAndAllplay)

	return events.APIGatewayProxyResponse{
		Body:       printTeam(franchisesWithStandingsAndAllplay),
		StatusCode: 200,
	}, nil
}

func printTeam(teams Franchises) string {
	t := table.NewWriter()
	t.SetOutputMirror(&bytes.Buffer{})
	t.AppendHeader(table.Row{"Team Name", "Owner", "Wins", "Losses", "Ties", "Fantasy Points", "Points", "Record", "Total Points",
		"AllPlay Wins", "AllPlay Losses", "AllPlay Ties", "AllPlay %"})
	for _, o := range teams {
		t.AppendRow([]interface{}{o.TeamName, o.OwnerName, o.RecordWins, o.RecordLosses, o.RecordTies, o.PointsForString, o.PointScore,
			o.RecordScoreString, o.TotalScore, o.AllPlayWins, o.AllPlayLosses, o.AllPlayTies, o.AllPlayPercentage})
	}

	fantasyPoints := []table.ColumnConfig{
		{Name: "Wins", Align: text.AlignCenter},
		{Name: "Losses", Align: text.AlignCenter},
		{Name: "Ties", Align: text.AlignCenter},
		{Name: "Fantasy Points", Align: text.AlignCenter},
		{Name: "Points", Align: text.AlignCenter},
		{Name: "Record", Align: text.AlignCenter},
		{Name: "Total Points", Align: text.AlignCenter},
		{Name: "AllPlay Wins", Align: text.AlignCenter},
		{Name: "AllPlay Losses", Align: text.AlignCenter},
		{Name: "AllPlay Ties", Align: text.AlignCenter},
		{Name: "AllPlay %", Align: text.AlignCenter},
	}

	sortBy := []table.SortBy{
		{Name: "Total Points", Mode: table.DscNumeric},
		{Name: "AllPlay %", Mode: table.DscNumeric},
	}

	t.SetColumnConfigs(fantasyPoints)
	t.SortBy(sortBy)
	return t.Render()
}

func calculateTotalScore(franchises Franchises) Franchises {
	for i := 0; i < len(franchises); i++ {
		franchises[i].TotalScore = strconv.FormatFloat(franchises[i].PointScore+franchises[i].RecordScore, 'f', 1, 64)
	}

	return franchises
}

func calculatePointsScore(franchises Franchises) Franchises {
	j, err := json.MarshalIndent(franchises, "", "    ")
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(string(j))
	for i := 0; i < len(franchises); {
		currentFantasyPoints := franchises[i].PointsFor
		var currentPointsForGrabs = float64(len(franchises) - i)
		var teamsTied float64 = 1
		for j := i + 1; j < len(franchises); j++ {
			if franchises[j].PointsFor == currentFantasyPoints {
				currentPointsForGrabs = currentPointsForGrabs + float64(len(franchises)) - float64(i) - teamsTied
				teamsTied++
			} else {
				break
			}
		}
		var pointsPerTeam = currentPointsForGrabs / teamsTied
		for k := 0; k < int(teamsTied); k++ {
			franchises[i+k].PointScore = pointsPerTeam
			franchises[i+k].PointScoreString = strconv.FormatFloat(pointsPerTeam, 'f', 1, 64)
		}
		i += int(teamsTied)
	}

	return franchises
}

func calculateRecordMagic(franchises Franchises) Franchises {
	for i := 0; i < len(franchises); i++ {
		franchises[i].RecordMagic = float64(franchises[i].RecordWins*1) + (float64(franchises[i].RecordTies) * 0.5)
	}

	return franchises
}

func calculateRecordScore(franchises Franchises) Franchises {
	for i := 0; i < len(franchises); {
		currentMagicPoints := franchises[i].RecordMagic
		var currentPointsForGrabs = float64(len(franchises) - i)
		var teamsTied float64 = 1
		for j := i + 1; j < len(franchises); j++ {
			if franchises[j].RecordMagic == currentMagicPoints {
				currentPointsForGrabs = currentPointsForGrabs + float64(len(franchises)) - float64(i) - teamsTied
				teamsTied++
			} else {
				break
			}
		}
		var pointsPerTeam = currentPointsForGrabs / teamsTied
		for k := 0; k < int(teamsTied); k++ {
			franchises[i+k].RecordScore = pointsPerTeam
			franchises[i+k].RecordScoreString = strconv.FormatFloat(pointsPerTeam, 'f', 1, 64)
		}
		i += int(teamsTied)
	}

	return franchises
}

func getFranchiseDetails(apiKey string) LeagueResponse {
	LeagueAPIURL := MflURL + LeagueYear + "/" + LeagueAPIPath + LeagueAPIQuery + "&" + LeagueIDQuery + "&" + APIOutputTypeQuery + "&APIKEY=" + apiKey
	fmt.Println("LeagueApiURL: " + LeagueAPIURL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, LeagueAPIURL, http.NoBody)
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var leagueResponse LeagueResponse
	err = json.Unmarshal(responseData, &leagueResponse)
	if err != nil {
		fmt.Println(err)
	}

	return leagueResponse
}

func getLeagueStandings(apiKey string) LeagueStandingsResponse {
	LeagueStandingsAPIURL := MflURL + LeagueYear + "/" + LeagueAPIPath + LeagueStandingsAPIQuery + "&" + LeagueIDQuery + "&" + APIOutputTypeQuery + "&APIKEY=" + apiKey
	fmt.Println("LeagueStandingsApiURL: " + LeagueStandingsAPIURL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, LeagueStandingsAPIURL, http.NoBody)
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var leagueStandingsResponse LeagueStandingsResponse
	err = json.Unmarshal(responseData, &leagueStandingsResponse)
	if err != nil {
		fmt.Println(err)
	}

	return leagueStandingsResponse
}

func checkResponseParity(leagueResponse LeagueResponse, leagueStandingsResponse LeagueStandingsResponse) {
	var numFranchises = convertStringToInteger(leagueResponse.League.Franchises.Count)
	numLeagueFranchises := len(leagueResponse.League.Franchises.Franchise)
	numLeagueStandingsFranchises := len(leagueStandingsResponse.LeagueStandings.Franchise)

	if numFranchises != numLeagueFranchises || numFranchises != numLeagueStandingsFranchises {
		fmt.Printf(
			"Responses don't have the same number of franchises:\n League API Franchises.Count: %d\n League API: %d\n LeagueStandings API: %d\n",
			numFranchises, numLeagueFranchises, numLeagueStandingsFranchises)
		os.Exit(3)
	}
}

func associateStandingsWithFranchises(franchiseDetailsResponse LeagueResponse, leagueStandingsResponse LeagueStandingsResponse) []Franchise {
	numLFranchises := len(franchiseDetailsResponse.League.Franchises.Franchise)
	numLSFranchises := len(leagueStandingsResponse.LeagueStandings.Franchise)

	franchiseStore := make([]Franchise, numLFranchises)
	for i := 0; i < numLFranchises; i++ {
		franchiseStore[i].TeamID = franchiseDetailsResponse.League.Franchises.Franchise[i].ID
		franchiseStore[i].TeamName = franchiseDetailsResponse.League.Franchises.Franchise[i].Name
		franchiseStore[i].OwnerName = franchiseDetailsResponse.League.Franchises.Franchise[i].OwnerName
		for j := 0; j < numLSFranchises; j++ {
			if franchiseStore[i].TeamID != leagueStandingsResponse.LeagueStandings.Franchise[j].ID {
				continue
			}

			franchiseStore[i].RecordWins = convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordWins)
			franchiseStore[i].RecordLosses = convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordLosses)
			franchiseStore[i].RecordTies = convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordTies)
			var err error
			franchiseStore[i].PointsFor, err = strconv.ParseFloat(leagueStandingsResponse.LeagueStandings.Franchise[j].PointsFor, 64)
			if err != nil {
				log.Fatal(err)
			}
			franchiseStore[i].PointsForString = leagueStandingsResponse.LeagueStandings.Franchise[j].PointsFor
		}
	}

	return franchiseStore
}

func convertStringToInteger(str string) int {
	integer, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal(err)
	}

	return integer
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func scrape() []AllPlayTeamStats {
	// c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	c := colly.NewCollector()

	c.WithTransport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   90 * time.Second,
			KeepAlive: 60 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	var allPlayTeamsStats []AllPlayTeamStats

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Scraping: ", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Status: ", r.StatusCode)
	})

	c.OnHTML("table.report > tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			allPlayTeamStats := AllPlayTeamStats{
				FranchiseName:     el.ChildText("td:nth-child(1)"),
				AllPlayWins:       el.ChildText("td:nth-child(13)"),
				AllPlayLosses:     el.ChildText("td:nth-child(14)"),
				AllPlayTies:       el.ChildText("td:nth-child(15)"),
				AllPlayPercentage: el.ChildText("td:nth-child(16)"),
			}
			allPlayTeamsStats = append(allPlayTeamsStats, allPlayTeamStats)
		})
	})

	_ = c.Visit(MflURL + LeagueYear + "/" + LeagueWebPath + LeagueIDQuery + //nolint:errcheck // error is checked next line
		"&" + PowerRankingsTableQuery + "&" + LeagueOutputSortQuery) //nolint:errcheck // error is checked next line

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	var allPlayTeamsStatsReturn []AllPlayTeamStats

	re := regexp.MustCompile(`^[a-zA-Z]`)

	for i := range allPlayTeamsStats {
		checker := re.FindString(allPlayTeamsStats[i].FranchiseName)

		if checker != "" {
			allPlayTeamsStatsReturn = append(allPlayTeamsStatsReturn, allPlayTeamsStats[i])
		}
	}

	fmt.Println("allPlayTeamsStatsReturn: ", allPlayTeamsStatsReturn)
	return allPlayTeamsStatsReturn
}

func appendAllPlay(franchises []Franchise, allPlayTeamData []AllPlayTeamStats) []Franchise {
	for franchise := range franchises {
		for team := range allPlayTeamData {
			if franchises[franchise].TeamName == allPlayTeamData[team].FranchiseName {
				franchises[franchise].AllPlayWins = convertStringToInteger(allPlayTeamData[team].AllPlayWins)
				franchises[franchise].AllPlayLosses = convertStringToInteger(allPlayTeamData[team].AllPlayLosses)
				franchises[franchise].AllPlayTies = convertStringToInteger(allPlayTeamData[team].AllPlayTies)
				franchises[franchise].AllPlayPercentage = allPlayTeamData[team].AllPlayPercentage
			}
		}
	}

	return franchises
}
