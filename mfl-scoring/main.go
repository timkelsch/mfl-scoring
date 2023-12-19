package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	Version  string `json:"version"`
	League   League `json:"league"`
	Encoding string `json:"encoding"`
}

type League struct {
	Franchises Franchises `json:"franchises"`
	ID         string     `json:"id"`
	History    History    `json:"history"`
	Name       string     `json:"name"`
	H2H        string     `json:"h2h"`
	BaseURL    string     `json:"baseURL"`
}

type Franchises struct {
	Franchise []Franchise `json:"franchise"`
}

type History struct {
	League []struct {
		URL  string `json:"url"`
		Year string `json:"year"`
	} `json:"league"`
}

type LeagueStandingsResponse struct {
	Version         string          `json:"version"`
	LeagueStandings LeagueStandings `json:"leagueStandings"`
	Encoding        string          `json:"encoding"`
}

type LeagueStandings struct {
	Franchise []Franchise `json:"franchise"`
}

type Franchise struct {
	TeamID                  string `json:"id"`
	TeamName                string `json:"name"`
	OwnerName               string `json:"owner_name"`
	RecordWins              int
	RecordWinsString        string `json:"h2hw"`
	RecordLosses            int
	RecordLossesString      string `json:"h2hl"`
	RecordTies              int
	RecordTiesString        string `json:"h2ht"`
	Record                  string
	PointsFor               float64
	PointsForString         string `json:"pf"`
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
		return events.APIGatewayProxyResponse{}, err
	}

	var APIKeySecretID = os.Getenv("API_KEY_SECRET_ID")
	apiKey, err := secretCache.GetSecretString(APIKeySecretID)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
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

	// Populate the slice of Franchise objects with league standing data
	franchisesWithStandings, err := associateStandingsWithFranchises(franchiseDetails, leagueStandings)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
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
	franchisesWithStandingsAndAllplay, err := appendAllPlay(calculatedTotalScore, allPlayTeamData)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	populatedAllPlayRecords := populateAllPlayRecords(franchisesWithStandingsAndAllplay)
	// fmt.Println("populatedAllPlayRecords: ", populatedAllPlayRecords)

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

// type Franchises struct { []Franchise }

type ByPointsFor struct{ Franchises }
type ByRecordMagic struct{ Franchises }
type ByTotalScore struct{ Franchises }
type ByAllPlayPercentage struct{ Franchises }

func (f Franchises) Len() int      { return len(f.Franchise) }
func (f Franchises) Swap(i, j int) { f.Franchise[i], f.Franchise[j] = f.Franchise[j], f.Franchise[i] }

func (f ByPointsFor) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchise[j].PointsFor < f.Franchise[i].PointsFor
}

func (f ByRecordMagic) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchise[j].RecordMagic < f.Franchise[i].RecordMagic
}

func (f ByTotalScore) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchise[j].TotalScore < f.Franchise[i].TotalScore
}

func (f ByAllPlayPercentage) Less(i, j int) bool {
	// Sort descending so j < i
	return f.Franchise[j].AllPlayPercentage < f.Franchise[i].AllPlayPercentage
}

func sortFranchises(teams Franchises) Franchises {
	// for _, team := range teams.Franchise {
	// 	fmt.Printf("teamID: %s, totalScore: %g, recordScore: %g, allPlayPct: %g \n",
	// 		team.TeamID, team.TotalScore, team.RecordScore, team.AllPlayPercentage)
	// 	fmt.Println("")
	// }

	sort.Sort(ByAllPlayPercentage{teams})
	sort.Sort(ByPointsFor{teams})
	sort.Sort(ByTotalScore{teams})

	// for _, team := range teams.Franchise {
	// 	fmt.Printf("teamID: %s, totalScore: %g, recordScore: %g, allPlayPct: %g \n",
	// 		team.TeamID, team.TotalScore, team.RecordScore, team.AllPlayPercentage)
	// 	fmt.Println("")
	// }

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
	for _, o := range teams.Franchise {
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
	for _, o := range teams.Franchise {
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
		"\n\nTeam names are hidden. There are some weirdos in this league."
}

func calculateTotalScore(franchises Franchises) Franchises {
	for i := range franchises.Franchise {
		// for i := 0; i < len(franchises); i++ {
		franchises.Franchise[i].TotalScore = franchises.Franchise[i].PointScore + franchises.Franchise[i].RecordScore
		franchises.Franchise[i].TotalScoreString =
			strconv.FormatFloat(franchises.Franchise[i].PointScore+franchises.Franchise[i].RecordScore, 'f', 1, 64)
	}

	return franchises
}

func calculatePointsScore(franchises Franchises) Franchises {
	/* j, err := json.MarshalIndent(franchises, "", "    ")
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Println(string(j)) */
	for i := 0; i < len(franchises.Franchise); {
		currentFantasyPoints := franchises.Franchise[i].PointsFor
		var currentPointsForGrabs = float64(len(franchises.Franchise) - i)
		var teamsTied float64 = 1
		for j := i + 1; j < len(franchises.Franchise); j++ {
			if franchises.Franchise[j].PointsFor == currentFantasyPoints {
				currentPointsForGrabs = currentPointsForGrabs + float64(len(franchises.Franchise)) -
					float64(i) - teamsTied
				teamsTied++
			} else {
				break
			}
		}
		var pointsPerTeam = currentPointsForGrabs / teamsTied
		for k := 0; k < int(teamsTied); k++ {
			franchises.Franchise[i+k].PointScore = pointsPerTeam
			franchises.Franchise[i+k].PointScoreString = strconv.FormatFloat(pointsPerTeam, 'f', 1, 64)
		}
		i += int(teamsTied)
	}

	return franchises
}

func calculateRecordMagic(franchises Franchises) Franchises {
	for i := range franchises.Franchise {
		franchises.Franchise[i].RecordMagic = float64(franchises.Franchise[i].RecordWins*1) +
			(float64(franchises.Franchise[i].RecordTies) * 0.5)
	}

	return franchises
}

func calculateRecordScore(franchises Franchises) Franchises {
	for i := 0; i < len(franchises.Franchise); {
		currentMagicPoints := franchises.Franchise[i].RecordMagic
		var currentPointsForGrabs = float64(len(franchises.Franchise) - i)
		var teamsTied float64 = 1
		for j := i + 1; j < len(franchises.Franchise); j++ {
			if franchises.Franchise[j].RecordMagic == currentMagicPoints {
				currentPointsForGrabs = currentPointsForGrabs + float64(len(franchises.Franchise)) -
					float64(i) - teamsTied
				teamsTied++
			} else {
				break
			}
		}
		var pointsPerTeam = currentPointsForGrabs / teamsTied
		for k := 0; k < int(teamsTied); k++ {
			franchises.Franchise[i+k].RecordScore = pointsPerTeam
			franchises.Franchise[i+k].RecordScoreString = strconv.FormatFloat(pointsPerTeam, 'f', 1, 64)
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
	numLeagueFranchises := len(leagueResponse.League.Franchises.Franchise)
	numLeagueStandingsFranchises := len(leagueStandingsResponse.LeagueStandings.Franchise)

	if numLeagueFranchises != numLeagueStandingsFranchises {
		fmt.Printf(
			"Responses don't have the same number of franchises:\n League API: %d\n LeagueStandings API: %d\n",
			numLeagueFranchises, numLeagueStandingsFranchises)
		os.Exit(3)
	}
}

func associateStandingsWithFranchises(franchiseDetailsResponse LeagueResponse,
	leagueStandingsResponse LeagueStandingsResponse) (Franchises, error) {
	checkResponseParity(franchiseDetailsResponse, leagueStandingsResponse)

	numLFranchises := len(franchiseDetailsResponse.League.Franchises.Franchise)
	numLSFranchises := len(leagueStandingsResponse.LeagueStandings.Franchise)

	var err error
	franchiseStore := Franchises{
		Franchise: make([]Franchise, numLFranchises),
	}
	for i := 0; i < numLFranchises; i++ {
		franchiseStore.Franchise[i].TeamID = franchiseDetailsResponse.League.Franchises.Franchise[i].TeamID
		franchiseStore.Franchise[i].TeamName = franchiseDetailsResponse.League.Franchises.Franchise[i].TeamName
		franchiseStore.Franchise[i].OwnerName = franchiseDetailsResponse.League.Franchises.Franchise[i].OwnerName
		for j := 0; j < numLSFranchises; j++ {
			if franchiseStore.Franchise[i].TeamID != leagueStandingsResponse.LeagueStandings.Franchise[j].TeamID {
				continue
			}

			franchiseStore.Franchise[i].RecordWins, err =
				convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordWinsString)
			if err != nil {
				return Franchises{}, err
			}
			franchiseStore.Franchise[i].RecordLosses, err =
				convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordLossesString)
			if err != nil {
				return Franchises{}, err
			}
			franchiseStore.Franchise[i].RecordTies, err =
				convertStringToInteger(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordTiesString)
			if err != nil {
				return Franchises{}, err
			}
			franchiseStore.Franchise[i].PointsFor, err =
				strconv.ParseFloat(leagueStandingsResponse.LeagueStandings.Franchise[j].PointsForString, 64)
			if err != nil {
				return Franchises{}, err
			}

			franchiseStore.Franchise[i].PointsForString = leagueStandingsResponse.LeagueStandings.Franchise[j].PointsForString
			franchiseStore.Franchise[i].RecordWinsString = leagueStandingsResponse.LeagueStandings.Franchise[j].RecordWinsString
			franchiseStore.Franchise[i].RecordLossesString = leagueStandingsResponse.LeagueStandings.Franchise[j].RecordLossesString
			franchiseStore.Franchise[i].RecordTiesString = leagueStandingsResponse.LeagueStandings.Franchise[j].RecordTiesString
		}
	}

	return franchiseStore, nil
}

func populateHeadToHeadRecords(franchises Franchises) Franchises {
	for i := 0; i < len(franchises.Franchise); i++ {
		franchises.Franchise[i].Record =
			strconv.Itoa(franchises.Franchise[i].RecordWins) + "-" +
				strconv.Itoa(franchises.Franchise[i].RecordLosses) + "-" +
				strconv.Itoa(franchises.Franchise[i].RecordTies)
	}

	return franchises
}

func convertStringToInteger(str string) (int, error) {
	integer, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return integer, nil
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

func appendAllPlay(franchises Franchises, allPlayTeamData []AllPlayTeamStats) (Franchises, error) {
	var err error

	for indexA := range franchises.Franchise {
		for indexB := range allPlayTeamData {
			if franchises.Franchise[indexA].TeamName != allPlayTeamData[indexB].FranchiseName {
				continue
			}

			franchises.Franchise[indexA].AllPlayWins, err = convertStringToInteger(allPlayTeamData[indexB].AllPlayWins)
			if err != nil {
				return Franchises{}, err
			}
			franchises.Franchise[indexA].AllPlayLosses, err = convertStringToInteger(allPlayTeamData[indexB].AllPlayLosses)
			if err != nil {
				return Franchises{}, err
			}
			franchises.Franchise[indexA].AllPlayTies, err = convertStringToInteger(allPlayTeamData[indexB].AllPlayTies)
			if err != nil {
				return Franchises{}, err
			}
			franchises.Franchise[indexA].AllPlayPercentageString = allPlayTeamData[indexB].AllPlayPercentage
			if err != nil {
				return Franchises{}, err
			}
			var allPlayPct float64
			allPlayPct, err = strconv.ParseFloat(allPlayTeamData[indexB].AllPlayPercentage, 64)
			if err != nil {
				return Franchises{}, err
			}
			franchises.Franchise[indexA].AllPlayPercentage = allPlayPct
		}
	}

	return franchises, nil
}

func populateAllPlayRecords(franchises Franchises) Franchises {
	for index := range franchises.Franchise {
		franchises.Franchise[index].AllPlayRecord =
			strconv.Itoa(franchises.Franchise[index].AllPlayWins) + "-" +
				strconv.Itoa(franchises.Franchise[index].AllPlayLosses) + "-" +
				strconv.Itoa(franchises.Franchise[index].AllPlayTies)
	}

	return franchises
}
