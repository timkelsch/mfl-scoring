package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jedib0t/go-pretty/v6/table"
)

type LeagueResponse struct {
	Version string `json:"version"`
	League  struct {
		CurrentWaiverType string `json:"currentWaiverType"`
		PlayerLimitUnit   string `json:"playerLimitUnit"`
		TaxiSquad         string `json:"taxiSquad"`
		EndWeek           string `json:"endWeek"`
		MaxWaiverRounds   string `json:"maxWaiverRounds"`
		Lockout           string `json:"lockout"`
		Franchises        struct {
			Count     string `json:"count"`
			Franchise []struct {
				Icon                 string `json:"icon,omitempty"`
				Name                 string `json:"name"`
				WaiverSortOrder      string `json:"waiverSortOrder"`
				LastVisit            string `json:"lastVisit"`
				Logo                 string `json:"logo,omitempty"`
				Email                string `json:"email"`
				BbidAvailableBalance string `json:"bbidAvailableBalance"`
				Id                   string `json:"id"`
				OwnerName            string `json:"owner_name"`
				Cell                 string `json:"cell,omitempty"`
				Country              string `json:"country,omitempty"`
				Phone                string `json:"phone,omitempty"`
				State                string `json:"state,omitempty"`
				Zip                  string `json:"zip,omitempty"`
				City                 string `json:"city,omitempty"`
				Address              string `json:"address,omitempty"`
				TwitterUsername      string `json:"twitterUsername,omitempty"`
				Abbrev               string `json:"abbrev,omitempty"`
				Sound                string `json:"sound,omitempty"`
				Url                  string `json:"url,omitempty"`
				Stadium              string `json:"stadium,omitempty"`
				PlayAudio            string `json:"play_audio,omitempty"`
				MailEvent            string `json:"mail_event,omitempty"`
				WirelessCarrier      string `json:"wireless_carrier,omitempty"`
				Username             string `json:"username,omitempty"`
				TimeZone             string `json:"time_zone,omitempty"`
				UseAdvancedEditor    string `json:"use_advanced_editor,omitempty"`
			} `json:"franchise"`
		} `json:"franchises"`
		StandingsSort string `json:"standingsSort"`
		Id            string `json:"id"`
		History       struct {
			League []struct {
				Url  string `json:"url"`
				Year string `json:"year"`
			} `json:"league"`
		} `json:"history"`
		RosterSize      string `json:"rosterSize"`
		Name            string `json:"name"`
		BbidSeasonLimit string `json:"bbidSeasonLimit"`
		RosterLimits    struct {
			Position []struct {
				Name  string `json:"name"`
				Limit string `json:"limit"`
			} `json:"position"`
		} `json:"rosterLimits"`
		BbidIncrement string `json:"bbidIncrement"`
		MobileAlerts  string `json:"mobileAlerts"`
		Starters      struct {
			Count    string `json:"count"`
			Position []struct {
				Name  string `json:"name"`
				Limit string `json:"limit"`
			} `json:"position"`
		} `json:"starters"`
		BestLineup            string `json:"bestLineup"`
		Precision             string `json:"precision"`
		LastRegularSeasonWeek string `json:"lastRegularSeasonWeek"`
		SurvivorPool          string `json:"survivorPool"`
		BbidTiebreaker        string `json:"bbidTiebreaker"`
		UsesContractYear      string `json:"usesContractYear"`
		MinKeepers            string `json:"minKeepers"`
		InjuredReserve        string `json:"injuredReserve"`
		BbidConditional       string `json:"bbidConditional"`
		StartWeek             string `json:"startWeek"`
		SurvivorPoolStartWeek string `json:"survivorPoolStartWeek"`
		SurvivorPoolEndWeek   string `json:"survivorPoolEndWeek"`
		RostersPerPlayer      string `json:"rostersPerPlayer"`
		BbidFCFSCharge        string `json:"bbidFCFSCharge"`
		LeagueLogo            string `json:"leagueLogo"`
		H2H                   string `json:"h2h"`
		UsesSalaries          string `json:"usesSalaries"`
		MaxKeepers            string `json:"maxKeepers"`
		BbidMinimum           string `json:"bbidMinimum"`
		BaseURL               string `json:"baseURL"`
		LoadRosters           string `json:"loadRosters"`
	} `json:"league"`
	Encoding string `json:"encoding"`
}

type LeagueStandingsResponse struct {
	Version         string `json:"version"`
	LeagueStandings struct {
		Franchise []struct {
			RecordLosses  string `json:"h2hl"`
			PowerRank     string `json:"power_rank"`
			Dp            string `json:"dp"`
			PointsFor     string `json:"pf"`
			StreakLen     string `json:"streak_len"`
			PointsAgainst string `json:"pa"`
			Maxpa         string `json:"maxpa"`
			Id            string `json:"id"`
			RecordTies    string `json:"h2ht"`
			AllPlayL      string `json:"all_play_l"`
			RecordWins    string `json:"h2hw"`
			AllPlayW      string `json:"all_play_w"`
			Vp            string `json:"vp"`
			Altpwr        string `json:"altpwr"`
			Pp            string `json:"pp"`
			Pwr           string `json:"pwr"`
			Minpa         string `json:"minpa"`
			AllPlayT      string `json:"all_play_t"`
			StreakType    string `json:"streak_type"`
			Op            string `json:"op"`
		} `json:"franchise"`
	} `json:"leagueStandings"`
	Encoding string `json:"encoding"`
}

type LeagueWeeklyResultsResponse struct {
	Version  string `json:"version"`
	Schedule struct {
		WeeklySchedule []struct {
			Week    string `json:"week"`
			Matchup []struct {
				Franchise []struct {
					ID     string `json:"id"`
					Result string `json:"result"`
					IsHome string `json:"isHome"`
					Score  string `json:"score"`
				} `json:"franchise"`
			} `json:"matchup"`
		} `json:"weeklySchedule"`
	} `json:"schedule"`
	Encoding string `json:"encoding"`
}

type Franchise struct {
	TeamID        string
	TeamName      string
	OwnerName     string
	RecordWins    int
	RecordLosses  int
	RecordTies    int
	PointsAgainst float64
	PointsFor     float64
	PointScore    float64
	RecordMagic   float64
	RecordScore   float64
	TotalScore    float64
}

type Config struct {
	MflApiKey string `envconfig:"MFL_API_KEY" required:"true"`
}

const (
	MflUrl             string = "https://www46.myfantasyleague.com/"
	LeagueYear         string = "2022"
	LeagueApi          string = "TYPE=league"
	LeagueStandingsApi string = "TYPE=leagueStandings"
	LeagueScheduleApi  string = "TYPE=schedule"
	LeagueId           string = "L=15781"
	ApiOutputType      string = "JSON=1"
)

type Franchises []Franchise

func (f Franchises) Len() int      { return len(f) }
func (f Franchises) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

type ByPointsFor struct{ Franchises }
type ByRecordMagic struct{ Franchises }
type ByTotalScore struct{ Franchises }

var numFranchises int

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

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	franchiseDetails := getFranchiseDetails()
	leagueStandings := getLeagueStandings()

	// Nullify points assigned on double matchup weeks.
	// 1. Get a list of weekly result results
	weeklyResults := getLeagueWeeklyResults()

	numTeamsInWeeklyResults := checkNumTeamsInWeeklyResults(weeklyResults)
	checkResponseParity(franchiseDetails, leagueStandings, numTeamsInWeeklyResults)

	// 2. Zero team fantasy points for double matchup weeks
	nullifyDoubleMatchupFantasyPoints(weeklyResults)

	// 3. Populate the array of Franchise objects so that we know which team ID is which
	teamInfo := associateStandingsWithFranchises(franchiseDetails, leagueStandings)

	// 4. Add point totals to []Franchise
	tabulateFantasyPoints(teamInfo, weeklyResults)

	sort.Sort(ByPointsFor{teamInfo})
	// Assign points to teams based on fantasy points scored
	calculatePointsScore(teamInfo)

	// Assign imaginary points to teams based on their record (1 point per win, 0.5 point per tie) so we can get the tied teams next to each other
	calculateRecordMagic(teamInfo)
	sort.Sort(ByRecordMagic{teamInfo})

	// Assign points to teams based on their record. Teams with identical records share the points they collectively earned,
	// Ex: If there are two teams tied for the best record, both teams would receive 9.5 points (10 points for first place + 9 points for second place, divided by 2 teams)
	calculateRecordScore(teamInfo)

	// Add up points assigned for fantasy points and points assigned for record
	calculateTotalScore(teamInfo)
	sort.Sort(ByTotalScore{teamInfo})

	return events.APIGatewayProxyResponse{
		Body:       printTeam(teamInfo),
		StatusCode: 200,
	}, nil
}

func printTeam(teams Franchises) string {
	t := table.NewWriter()
	t.SetOutputMirror(&bytes.Buffer{})
	t.AppendHeader(table.Row{"Team Name", "Owner", "Wins", "Losses", "Ties", "Fantasy Points", "Points", "Record", "Total Points"})
	for _, o := range teams {
		t.AppendRow([]interface{}{o.TeamName, o.OwnerName, o.RecordWins, o.RecordLosses, o.RecordTies, o.PointsFor, o.PointScore, o.RecordScore, o.TotalScore})
	}

	return t.Render()
}

func calculateTotalScore(franchises Franchises) Franchises {
	for i := 0; i < len(franchises); i++ {
		franchises[i].TotalScore = franchises[i].PointScore + franchises[i].RecordScore
	}

	return franchises
}

func calculatePointsScore(franchises Franchises) Franchises {
	for i := 0; i < len(franchises); i++ {
		franchises[i].PointScore = float64(len(franchises) - i)
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
		}
		i += int(teamsTied)
	}

	return franchises
}

func getFranchiseDetails() LeagueResponse {
	LeagueApiURL := MflUrl + LeagueYear + "/export?" + LeagueApi + "&" + LeagueId + "&" + ApiOutputType + "&APIKEY=" + os.Getenv("MFL_API_KEY")
	fmt.Println("LeagueApiURL: " + LeagueApiURL)
	response, err := http.Get(LeagueApiURL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(2)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var leagueResponse LeagueResponse
	err = json.Unmarshal(responseData, &leagueResponse)
	if err != nil {
		log.Fatal(err)
	}

	return leagueResponse
}

func getLeagueStandings() LeagueStandingsResponse {
	LeagueStandingsApiURL := MflUrl + LeagueYear + "/export?" + LeagueStandingsApi + "&" + LeagueId + "&" + ApiOutputType + "&APIKEY=" + os.Getenv("MFL_API_KEY")
	fmt.Println("LeagueStandingsApiURL: " + LeagueStandingsApiURL)
	response, err := http.Get(LeagueStandingsApiURL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var leagueStandingsResponse LeagueStandingsResponse
	err = json.Unmarshal(responseData, &leagueStandingsResponse)
	if err != nil {
		log.Fatal(err)
	}

	return leagueStandingsResponse
}

func getLeagueWeeklyResults() LeagueWeeklyResultsResponse {
	LeagueWeeklyResultsApiURL := MflUrl + LeagueYear + "/export?" + LeagueScheduleApi + "&" + LeagueId + "&" + ApiOutputType + "&APIKEY=" + os.Getenv("MFL_API_KEY") //cfg.MflApiKey
	fmt.Println("LeagueWeeklyResultsApiURL: " + LeagueWeeklyResultsApiURL)
	response, err := http.Get(LeagueWeeklyResultsApiURL)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var leagueWeeklyResultsResponse LeagueWeeklyResultsResponse
	err = json.Unmarshal(responseData, &leagueWeeklyResultsResponse)
	if err != nil {
		log.Fatal(err)
	}

	return leagueWeeklyResultsResponse
}

func checkNumTeamsInWeeklyResults(leagueWeeklyResultsResponse LeagueWeeklyResultsResponse) int {
	var uniqueTeamIDs []string
	for week := 0; week < len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule); week++ {
		for matchup := 0; matchup < len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup); matchup++ {
			for franchise := 0; franchise < len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise); franchise++ {
				uniqueTeamIDs = appendIfMissing(uniqueTeamIDs, leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise[franchise].ID)
			}
		}
	}

	return len(uniqueTeamIDs)
}

func appendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func checkResponseParity(leagueResponse LeagueResponse, leagueStandingsResponse LeagueStandingsResponse, numTeamsInWeeklyResults int) {
	numFranchises, _ = strconv.Atoi(leagueResponse.League.Franchises.Count)
	numLeagueFranchises := len(leagueResponse.League.Franchises.Franchise)
	numLeagueStandingsFranchises := len(leagueStandingsResponse.LeagueStandings.Franchise)

	if numFranchises != numLeagueStandingsFranchises || numFranchises != numLeagueStandingsFranchises || numFranchises != numTeamsInWeeklyResults {
		fmt.Printf("Responses don't have the same number of franchises:\n League API Franchises.Count: %d\n League API: %d\n LeagueStandings API: %d\n ScheduleAPI: %d\n",
			numFranchises, numLeagueFranchises, numLeagueStandingsFranchises, numTeamsInWeeklyResults)
		os.Exit(3)
	}
}

func nullifyDoubleMatchupFantasyPoints(leagueWeeklyResultsResponse LeagueWeeklyResultsResponse) LeagueWeeklyResultsResponse {
	numWeeks := len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule)

	for week := 0; week < numWeeks; week++ {
		matchups := len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup)
		if matchups > (numFranchises / 2) {
			for matchup := 0; matchup < matchups; matchup++ {
				for franchise := 0; franchise < len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise); franchise++ {
					leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise[franchise].Score = "0"
				}
			}
		}
	}

	return leagueWeeklyResultsResponse
}

func associateStandingsWithFranchises(franchiseDetailsResponse LeagueResponse, leagueStandingsResponse LeagueStandingsResponse) []Franchise {
	numLFranchises := len(franchiseDetailsResponse.League.Franchises.Franchise)
	numLSFranchises := len(leagueStandingsResponse.LeagueStandings.Franchise)

	franchiseStore := make([]Franchise, numLFranchises)
	for i := 0; i < numLFranchises; i++ {
		franchiseStore[i].TeamID = franchiseDetailsResponse.League.Franchises.Franchise[i].Id
		franchiseStore[i].TeamName = franchiseDetailsResponse.League.Franchises.Franchise[i].Name
		franchiseStore[i].OwnerName = franchiseDetailsResponse.League.Franchises.Franchise[i].OwnerName
		for j := 0; j < numLSFranchises; j++ {
			if franchiseStore[i].TeamID == leagueStandingsResponse.LeagueStandings.Franchise[j].Id {
				franchiseStore[i].RecordWins, _ = strconv.Atoi(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordWins)
				franchiseStore[i].RecordLosses, _ = strconv.Atoi(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordLosses)
				franchiseStore[i].RecordTies, _ = strconv.Atoi(leagueStandingsResponse.LeagueStandings.Franchise[j].RecordTies)
			}
		}
	}

	return franchiseStore
}

func tabulateFantasyPoints(franchiseStore []Franchise, leagueWeeklyResultsResponse LeagueWeeklyResultsResponse) []Franchise {
	numWeeks := len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule)
	//numFranchises := len(franchiseStore)

	for franchiseStoreTeam := 0; franchiseStoreTeam < len(franchiseStore); franchiseStoreTeam++ {
		for week := 0; week < numWeeks; week++ {
			for matchup := 0; matchup < len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup); matchup++ {
				for franchise := 0; franchise < len(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise); franchise++ {
					if leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise[franchise].ID == franchiseStore[franchiseStoreTeam].TeamID {
						s, err := strconv.ParseFloat(leagueWeeklyResultsResponse.Schedule.WeeklySchedule[week].Matchup[matchup].Franchise[franchise].Score, 64)
						if err != nil {
							fmt.Print(err.Error())
							os.Exit(3)
						} else {
							franchiseStore[franchiseStoreTeam].PointsFor += s
						}
					}
				}
			}
		}
		franchiseStore[franchiseStoreTeam].PointsFor = roundFloat(franchiseStore[franchiseStoreTeam].PointsFor, 1)
	}

	return franchiseStore
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
