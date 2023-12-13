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
	"strings"
	"sync"
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
	TeamID                  string
	TeamName                string
	OwnerName               string
	RecordWins              int
	RecordLosses            int
	RecordTies              int
	Record                  string
	PointsFor               float64
	PointsForString         string
	PointScore              float64
	PointScoreString        string
	RecordMagic             float64
	RecordScore             float64
	RecordScoreString       string
	TotalScoreString        string
	TotalScore              float64
	AllPlayWins             int
	AllPlayLosses           int
	AllPlayTies             int
	AllPlayRecord           string
	AllPlayPercentageString string
	AllPlayPercentage       float64
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
)

type AllPlayTeamStats struct {
	FranchiseName     string
	AllPlayWins       string
	AllPlayLosses     string
	AllPlayTies       string
	AllPlayPercentage string
}

func main() {
	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	secretCache, err := secretcache.New()
	if err != nil {
		log.Fatal(err)
	}

	var APIKeySecretID = os.Getenv("API_KEY_SECRET_ID")
	apiKey, err := secretCache.GetSecretString(APIKeySecretID)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var franchiseDetails LeagueResponse
	var leagueStandings LeagueStandingsResponse

	go func() {
		franchiseDetails = getFranchiseDetails(apiKey)
		wg.Done()
	}()

	go func() {
		leagueStandings = getLeagueStandings(apiKey)
		wg.Done()
	}()

	wg.Wait()

	checkResponseParity(franchiseDetails, leagueStandings)

	// Populate the slice of Franchise objects with league standing data
	franchisesWithStandings := associateStandingsWithFranchises(franchiseDetails, leagueStandings)
	populatedHeadToHeadRecords := populateHeadToHeadRecords(franchisesWithStandings)

	// Put teams in order of most fantasy points scored
	sort.Sort(ByPointsFor{populatedHeadToHeadRecords})
	// fmt.Printf("%+v \n", populatedHeadToHeadRecords)

	// Assign points to teams based on fantasy points scored, sharing points as necessary when teams tie
	calculatedPointScore := calculatePointsScore(populatedHeadToHeadRecords)

	// Put teams in order of best head to head record
	calculatedRecordMagic := calculateRecordMagic(calculatedPointScore)
	sort.Sort(ByRecordMagic{calculatedRecordMagic})

	// Assign points to teams based on head to head record, sharing points as necessary when teams tie
	calculatedRecordScore := calculateRecordScore(calculatedRecordMagic)

	// totalScore = points assigned for fantasy points + points assigned for record
	calculatedTotalScore := calculateTotalScore(calculatedRecordScore)

	allPlayTeamData := scrape()
	franchisesWithStandingsAndAllplay := appendAllPlay(calculatedTotalScore, allPlayTeamData)
	populatedAllPlayRecords := populateAllPlayRecords(franchisesWithStandingsAndAllplay)
	fmt.Println("populatedAllPlayRecords: ", populatedAllPlayRecords)

	sortedFranchises := sortFranchises(populatedAllPlayRecords)

	// fmt.Printf("requestContext.DomainName: %v\n", request.RequestContext.DomainName)
	// fmt.Printf("requestContext.QueryStringParameters: %v\n", request.QueryStringParameters)
	if outputFormat, exists := request.QueryStringParameters["output"]; exists {
		if outputFormat == "json" {
			headers := map[string]string{
				"content-type":                     "application/json",
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			}
			fmt.Println("headers: ", headers)
			body, err := json.Marshal(sortedFranchises)
			if err != nil {
				panic(err)
			}
			return events.APIGatewayProxyResponse{
				Headers:    headers,
				Body:       string(body),
				StatusCode: 200,
			}, nil
		}
	}

	if strings.Contains(request.RequestContext.DomainName, "execute-api") {
		return events.APIGatewayProxyResponse{
			Body:       printScoringTableCouthly(sortedFranchises),
			StatusCode: 200,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       printScoringTableUncouthly(sortedFranchises),
		StatusCode: 200,
	}, nil
}

type Franchises []Franchise

type ByPointsFor struct{ Franchises }
type ByRecordMagic struct{ Franchises }
type ByTotalScore struct{ Franchises }
type ByAllPlayPercentage struct{ Franchises }

func (f Franchises) Len() int      { return len(f) }
func (f Franchises) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (f ByPointsFor) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchises[j].PointsFor < f.Franchises[i].PointsFor
}

func (f ByRecordMagic) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchises[j].RecordMagic < f.Franchises[i].RecordMagic
}

func (f ByTotalScore) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchises[j].TotalScore < f.Franchises[i].TotalScore
}

func (f ByAllPlayPercentage) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchises[j].AllPlayPercentage < f.Franchises[i].AllPlayPercentage
}

func sortFranchises(teams Franchises) Franchises {
	// fmt.Println(teams)
	sort.Sort(ByAllPlayPercentage{teams})
	sort.Sort(ByPointsFor{teams})
	sort.Sort(ByTotalScore{teams})

	for _, team := range teams {
		fmt.Printf("totalScore: %g, recordScore: %g, allPlayPct: %g \n",
			team.TotalScore, team.RecordScore, team.AllPlayPercentage)
		fmt.Println("")
	}

	return teams
}

const (
	TotalPts      string = "Total Pts"
	Record        string = "W-L-T"
	FantasyPts    string = "Fantasy Pts"
	PtsScore      string = "Pts Score"
	RecScore      string = "Rcrd Score"
	AllPlayRecord string = "AllPlay W-L-T"
	AllPlayPct    string = "AllPlay %"
)

func printScoringTableUncouthly(teams Franchises) string {
	t := table.NewWriter()
	t.SetOutputMirror(&bytes.Buffer{})
	t.AppendHeader(table.Row{"Team Name", "Owner", Record, FantasyPts, PtsScore, RecScore, TotalPts,
		AllPlayRecord, AllPlayPct})
	for _, o := range teams {
		t.AppendRow([]interface{}{o.TeamName, o.OwnerName, o.Record, o.PointsForString, o.PointScore,
			o.RecordScoreString, o.TotalScoreString, o.AllPlayRecord, o.AllPlayPercentageString})
	}

	columnConfigs := []table.ColumnConfig{
		{Name: Record, Align: text.AlignCenter},
		{Name: FantasyPts, Align: text.AlignCenter},
		{Name: PtsScore, Align: text.AlignCenter},
		{Name: RecScore, Align: text.AlignCenter},
		{Name: TotalPts, Align: text.AlignCenter},
		{Name: AllPlayRecord, Align: text.AlignCenter},
		{Name: AllPlayPct, Align: text.AlignCenter},
	}

	// sortBy := []table.SortBy{
	// 	{Name: TotalPts, Mode: table.DscNumeric},
	// 	{Name: RecScore, Mode: table.DscNumeric},
	// 	{Name: AllPlayPct, Mode: table.DscNumeric},
	// }

	t.SetColumnConfigs(columnConfigs)
	// t.SortBy(sortBy)
	return t.Render()
}

// Hide uncouth team names for professional project.
func printScoringTableCouthly(teams Franchises) string {
	t := table.NewWriter()
	t.SetOutputMirror(&bytes.Buffer{})
	t.AppendHeader(table.Row{"Team ID", Record, FantasyPts, PtsScore, RecScore, TotalPts,
		AllPlayRecord, AllPlayPct})
	for _, o := range teams {
		t.AppendRow([]interface{}{o.TeamID, o.Record, o.PointsForString, o.PointScore,
			o.RecordScoreString, o.TotalScoreString, o.AllPlayRecord, o.AllPlayPercentageString})
	}

	columnConfigs := []table.ColumnConfig{
		{Name: Record, Align: text.AlignCenter},
		{Name: FantasyPts, Align: text.AlignCenter},
		{Name: PtsScore, Align: text.AlignCenter},
		{Name: RecScore, Align: text.AlignCenter},
		{Name: TotalPts, Align: text.AlignCenter},
		{Name: AllPlayRecord, Align: text.AlignCenter},
		{Name: AllPlayPct, Align: text.AlignCenter},
	}

	// sortBy := []table.SortBy{
	// 	{Name: TotalPts, Mode: table.DscNumeric},
	// 	{Name: RecScore, Mode: table.DscNumeric},
	// 	{Name: AllPlayPct, Mode: table.DscNumeric},
	// }

	t.SetColumnConfigs(columnConfigs)
	// t.SortBy(sortBy)
	return t.Render() +
		"\n\nTeam names are hidden. There are some weirdos in this league.  "
}

func calculateTotalScore(franchises Franchises) Franchises {
	for i := range franchises {
		// for i := 0; i < len(franchises); i++ {
		franchises[i].TotalScore = franchises[i].PointScore + franchises[i].RecordScore
		franchises[i].TotalScoreString =
			strconv.FormatFloat(franchises[i].PointScore+franchises[i].RecordScore, 'f', 1, 64)
	}

	return franchises
}

func calculatePointsScore(franchises Franchises) Franchises {
	/* j, err := json.MarshalIndent(franchises, "", "    ")
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(string(j)) */
	for i := 0; i < len(franchises); {
		currentFantasyPoints := franchises[i].PointsFor
		var currentPointsForGrabs = float64(len(franchises) - i)
		var teamsTied float64 = 1
		for j := i + 1; j < len(franchises); j++ {
			if franchises[j].PointsFor == currentFantasyPoints {
				currentPointsForGrabs = currentPointsForGrabs + float64(len(franchises)) -
					float64(i) - teamsTied
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
	for i := range franchises {
		franchises[i].RecordMagic = float64(franchises[i].RecordWins*1) +
			(float64(franchises[i].RecordTies) * 0.5)
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
				currentPointsForGrabs = currentPointsForGrabs + float64(len(franchises)) -
					float64(i) - teamsTied
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
	LeagueAPIURL := MflURL + LeagueYear + "/" + LeagueAPIPath + LeagueAPIQuery + "&" +
		LeagueIDQuery + "&" + APIOutputTypeQuery + "&APIKEY=" + apiKey
	// fmt.Println("LeagueApiURL: " + LeagueAPIURL) asdf

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
	LeagueStandingsAPIURL := MflURL + LeagueYear + "/" + LeagueAPIPath + LeagueStandingsAPIQuery + "&" +
		LeagueIDQuery + "&" + APIOutputTypeQuery + "&APIKEY=" + apiKey
	// fmt.Println("LeagueStandingsApiURL: " + LeagueStandingsAPIURL)

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

func associateStandingsWithFranchises(franchiseDetailsResponse LeagueResponse,
	leagueStandingsResponse LeagueStandingsResponse) []Franchise {
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

			franchiseStore[i].RecordWins =
				convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordWins)
			franchiseStore[i].RecordLosses =
				convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordLosses)
			franchiseStore[i].RecordTies =
				convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordTies)
			var err error
			franchiseStore[i].PointsFor, err =
				strconv.ParseFloat(leagueStandingsResponse.LeagueStandings.Franchise[j].PointsFor, 64)
			if err != nil {
				log.Fatal(err)
			}
			franchiseStore[i].PointsForString = leagueStandingsResponse.LeagueStandings.Franchise[j].PointsFor
		}
	}

	return franchiseStore
}

func populateHeadToHeadRecords(franchises []Franchise) []Franchise {
	for index := range franchises {
		franchises[index].Record =
			strconv.Itoa(franchises[index].RecordWins) + "-" +
				strconv.Itoa(franchises[index].RecordLosses) + "-" +
				strconv.Itoa(franchises[index].RecordTies)
	}

	return franchises
}

func convertStringToInteger(str string) int {
	integer, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("CSTI - str: %s - int: %d\n", str, integer)

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

	_ = c.Visit(MflURL + LeagueYear + "/" + LeagueWebPath + LeagueIDQuery + //nolint:errcheck // err checked next line
		"&" + PowerRankingsTableQuery + "&" + LeagueOutputSortQuery) //nolint:errcheck // err checked next line

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

	// fmt.Println("allPlayTeamsStatsReturn: ", allPlayTeamsStatsReturn)
	return allPlayTeamsStatsReturn
}

func appendAllPlay(franchises []Franchise, allPlayTeamData []AllPlayTeamStats) []Franchise {
	fmt.Println("allPlayTeamData: ", allPlayTeamData)
	for indexA := range franchises {
		for indexB := range allPlayTeamData {
			if franchises[indexA].TeamName == allPlayTeamData[indexB].FranchiseName {
				franchises[indexA].AllPlayWins = convertStringToInteger(allPlayTeamData[indexB].AllPlayWins)
				franchises[indexA].AllPlayLosses = convertStringToInteger(allPlayTeamData[indexB].AllPlayLosses)
				franchises[indexA].AllPlayTies = convertStringToInteger(allPlayTeamData[indexB].AllPlayTies)
				franchises[indexA].AllPlayPercentageString = allPlayTeamData[indexB].AllPlayPercentage
				allPlayPct, err := strconv.ParseFloat(allPlayTeamData[indexB].AllPlayPercentage, 64)
				if err != nil {
					log.Fatal(err)
				}
				franchises[indexA].AllPlayPercentage = allPlayPct
				// Here AllPlay W/L/T is correct
				// fmt.Printf("Franchise: %s - Team: %s - FALP: %d - TALP: %d\n",
				// franchiseObj.TeamName, teamObj.FranchiseName, franchises[indexA].AllPlayWins,
				// convertStringToInteger(allPlayTeamData[indexB].AllPlayWins))
				// Here AllPlay W/L/T is gone. WTF
				// I had put populateAllPlayRecords() way too early in main
				// fmt.Println("franchisest: ", franchises)
			}
		}
	}
	return franchises
}

func populateAllPlayRecords(franchises []Franchise) []Franchise {
	for index := range franchises {
		franchises[index].AllPlayRecord =
			strconv.Itoa(franchises[index].AllPlayWins) + "-" +
				strconv.Itoa(franchises[index].AllPlayLosses) + "-" +
				strconv.Itoa(franchises[index].AllPlayTies)
	}

	return franchises
}
