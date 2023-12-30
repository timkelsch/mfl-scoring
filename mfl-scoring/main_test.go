package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/html"
)

const (
	Team1Name  = "Team 1"
	Team2Name  = "Team 2"
	Team1Owner = "Owner 1"
	Team2Owner = "Owner 2"
)

func TestRoundFloat(t *testing.T) {
	type args struct {
		value     float64
		precision uint
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "small", args: args{-0.34567, 3}, want: -0.346},
		{name: "large", args: args{4923487768956.98234779857, 8}, want: 4923487768956.98234780},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundFloat(tt.args.value, tt.args.precision)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateRecordMagic(t *testing.T) {
	testCases := []struct {
		name       string
		franchises Franchises
		expected   Franchises
	}{
		{
			name: "one",
			franchises: Franchises{
				Franchise: []Franchise{
					{RecordWins: 6, RecordTies: 1},
					{RecordWins: 3, RecordTies: 7},
					{RecordWins: 1, RecordTies: 0},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{RecordWins: 6, RecordTies: 1, RecordMagic: 6.5},
					{RecordWins: 3, RecordTies: 7, RecordMagic: 6.5},
					{RecordWins: 1, RecordTies: 0, RecordMagic: 1},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateRecordMagic(tc.franchises)
			for i := range result.Franchise {
				if result.Franchise[i].RecordMagic != tc.expected.Franchise[i].RecordMagic {
					t.Errorf("Mismatch in test case %s for franchise %d: Expected %f, got %f",
						tc.name, i, tc.expected.Franchise[i].RecordMagic, result.Franchise[i].RecordMagic)
				}
			}
		})
	}
}

func TestCalculateTotalScore(t *testing.T) {
	testCases := []struct {
		name       string
		franchises Franchises
		expected   Franchises
	}{
		{
			name: "one",
			franchises: Franchises{
				Franchise: []Franchise{
					{PointScore: 3, RecordScore: 4.5},
					{PointScore: 7, RecordScore: 9},
					{PointScore: 2, RecordScore: 1.5},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{PointScore: 3, RecordScore: 4.5, TotalScoreString: "7.5"},
					{PointScore: 7, RecordScore: 9, TotalScoreString: "16.0"},
					{PointScore: 2, RecordScore: 1.5, TotalScoreString: "3.5"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateTotalScore(tc.franchises)
			for i := range result.Franchise {
				if result.Franchise[i].TotalScoreString != tc.expected.Franchise[i].TotalScoreString {
					t.Errorf("Mismatch in test case %s for franchise %d: Expected %s, got %s",
						tc.name, i, tc.expected.Franchise[i].TotalScoreString, result.Franchise[i].TotalScoreString)
				}
			}
		})
	}
}

func TestCalculateRecordScore(t *testing.T) {
	testCases := []struct {
		name       string
		franchises Franchises
		expected   Franchises
	}{
		{
			name: "Test with ties",
			franchises: Franchises{
				Franchise: []Franchise{
					{RecordMagic: 8.5},
					{RecordMagic: 8.5},
					{RecordMagic: 7},
					{RecordMagic: 5},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{RecordMagic: 8.5, RecordScore: 3.5, RecordScoreString: "3.5"},
					{RecordMagic: 8.5, RecordScore: 3.5, RecordScoreString: "3.5"},
					{RecordMagic: 7, RecordScore: 2, RecordScoreString: "2.0"},
					{RecordMagic: 5, RecordScore: 1, RecordScoreString: "1.0"},
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateRecordScore(tc.franchises)
			for i := range result.Franchise {
				if result.Franchise[i].RecordScore != tc.expected.Franchise[i].RecordScore {
					t.Errorf("Mismatch in test case %s for franchise %d in field %s: Expected %f, got %f",
						tc.name, i, "RecordScore", tc.expected.Franchise[i].RecordScore, result.Franchise[i].RecordScore)
				}
				if result.Franchise[i].RecordScoreString != tc.expected.Franchise[i].RecordScoreString {
					t.Errorf("Mismatch in test case %s for franchise %d in field %s: Expected %s, got %s",
						tc.name, i, "RecordScoreString", tc.expected.Franchise[i].RecordScoreString, result.Franchise[i].RecordScoreString)
				}
			}
		})
	}
}

func TestCalculatePointsScore(t *testing.T) {
	testCases := []struct {
		name       string
		franchises Franchises
		expected   Franchises
	}{
		{
			name: "Test with ties",
			franchises: Franchises{
				Franchise: []Franchise{
					{PointsFor: 15},
					{PointsFor: 15},
					{PointsFor: 10},
					{PointsFor: 5},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{PointsFor: 15, PointScore: 3.5, PointScoreString: "3.5"},
					{PointsFor: 15, PointScore: 3.5, PointScoreString: "3.5"},
					{PointsFor: 10, PointScore: 2, PointScoreString: "2.0"},
					{PointsFor: 5, PointScore: 1, PointScoreString: "1.0"},
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculatePointsScore(tc.franchises)
			for i := range result.Franchise {
				if result.Franchise[i].PointScore != tc.expected.Franchise[i].PointScore {
					t.Errorf("Mismatch in test case %s for franchise %d: Expected %f, got %f",
						tc.name, i, tc.expected.Franchise[i].PointScore, result.Franchise[i].PointScore)
				}
			}
		})
	}
}

func TestSortFranchises(t *testing.T) {
	testCases := []struct {
		name       string
		franchises Franchises
		expected   Franchises
	}{
		{
			name: "Test with different scores",
			franchises: Franchises{
				Franchise: []Franchise{
					{TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.5},
					{TotalScore: 30, PointsFor: 30, AllPlayPercentage: 0.4},
					{TotalScore: 20, PointsFor: 10, AllPlayPercentage: 0.6},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{TotalScore: 30, PointsFor: 30, AllPlayPercentage: 0.4},
					{TotalScore: 20, PointsFor: 10, AllPlayPercentage: 0.6},
					{TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.5},
				},
			},
		},
		{
			name: "Test with same scores",
			franchises: Franchises{
				Franchise: []Franchise{
					{TeamID: "3", TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.56},
					{TeamID: "1", TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.59},
					{TeamID: "0", TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.51},
					{TeamID: "2", TotalScore: 10, PointsFor: 21, AllPlayPercentage: 0.51},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{TeamID: "2", TotalScore: 10, PointsFor: 21, AllPlayPercentage: 0.51},
					{TeamID: "1", TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.59},
					{TeamID: "3", TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.56},
					{TeamID: "0", TotalScore: 10, PointsFor: 20, AllPlayPercentage: 0.51},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errorString := "Mismatch in test case %s checking %s for franchise %d: Expected %f, got %f"
			result := sortFranchises(tc.franchises)
			for i := range result.Franchise {
				if result.Franchise[i].TotalScore != tc.expected.Franchise[i].TotalScore {
					t.Errorf(errorString, tc.name,
						"TeamID", i, tc.expected.Franchise[i].TeamID, result.Franchise[i].TeamID)
				}
			}

			for i := range result.Franchise {
				if result.Franchise[i].TotalScore != tc.expected.Franchise[i].TotalScore {
					t.Errorf(errorString, tc.name,
						"TotalScore", i, tc.expected.Franchise[i].TotalScore, result.Franchise[i].TotalScore)
				}
			}

			for i := range result.Franchise {
				if result.Franchise[i].PointsFor != tc.expected.Franchise[i].PointsFor {
					t.Errorf(errorString, tc.name,
						"PointsFor", i, tc.expected.Franchise[i].PointsFor, result.Franchise[i].PointsFor)
				}
			}

			for i := range result.Franchise {
				if result.Franchise[i].AllPlayPercentage != tc.expected.Franchise[i].AllPlayPercentage {
					t.Errorf(errorString, tc.name,
						"AllPlayPercentage", i, tc.expected.Franchise[i].AllPlayPercentage, result.Franchise[i].AllPlayPercentage)
				}
			}
		})
	}
}

func TestPrintScoringTableUncouthly(t *testing.T) {
	teams := Franchises{
		Franchise: []Franchise{
			{
				TeamName:                Team1Name,
				OwnerName:               Team1Owner,
				Record:                  "12-6-0",
				PointsForString:         "2068.4",
				PointScore:              10.0,
				RecordScoreString:       "9.0",
				TotalScoreString:        "19.0",
				AllPlayRecord:           "111-33-0",
				AllPlayPercentageString: ".771",
			},
			// Add more teams as needed...
		},
	}

	expected := `+-----------+---------+--------+-------------+-----------+------------+-----------+---------------+-----------+
| TEAM NAME | OWNER   | W-L-T  | FANTASY PTS | PTS SCORE | RCRD SCORE | TOTAL PTS | ALLPLAY W-L-T | ALLPLAY % |
+-----------+---------+--------+-------------+-----------+------------+-----------+---------------+-----------+
| Team 1    | Owner 1 | 12-6-0 |    2068.4   |     10    |     9.0    |    19.0   |    111-33-0   |    .771   |
+-----------+---------+--------+-------------+-----------+------------+-----------+---------------+-----------+`

	result := printScoringTableUncouthly(teams)
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestPrintScoringTableCouthly(t *testing.T) {
	teams := Franchises{
		Franchise: []Franchise{
			{
				TeamID:                  "0003",
				Record:                  "9-9-0",
				PointsForString:         "1603.5",
				PointScore:              3,
				RecordScoreString:       "5.5",
				TotalScoreString:        "8.5",
				AllPlayRecord:           "69-74-1",
				AllPlayPercentageString: ".483",
			},
			// Add more teams as needed...
		},
	}

	expected := `+---------+-------+-------------+-----------+------------+-----------+---------------+-----------+
| TEAM ID | W-L-T | FANTASY PTS | PTS SCORE | RCRD SCORE | TOTAL PTS | ALLPLAY W-L-T | ALLPLAY % |
+---------+-------+-------------+-----------+------------+-----------+---------------+-----------+
| 0003    | 9-9-0 |    1603.5   |     3     |     5.5    |    8.5    |    69-74-1    |    .483   |
+---------+-------+-------------+-----------+------------+-----------+---------------+-----------+

Team names are hidden. There are some weirdos in this league. `

	result := printScoringTableCouthly(teams)
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestAssociateStandingsWithFranchises(t *testing.T) {
	testCases := []struct {
		name                     string
		franchiseDetailsResponse LeagueResponse
		leagueStandingsResponse  LeagueStandingsResponse
		expected                 Franchises
		expectError              bool
	}{
		{
			name: "Test case 1",
			franchiseDetailsResponse: LeagueResponse{
				League: League{
					Franchises: Franchises{
						Franchise: []Franchise{
							{TeamID: "1", TeamName: Team1Name, OwnerName: Team1Owner}, // define a constant for these values and use them to replace all instances of these strings
							{TeamID: "2", TeamName: Team2Name, OwnerName: Team2Owner},
						},
					},
				},
			},
			leagueStandingsResponse: LeagueStandingsResponse{
				LeagueStandings: LeagueStandings{
					Franchise: []Franchise{
						{TeamID: "1", PointsForString: "100.0", RecordWinsString: "5", RecordLossesString: "3", RecordTiesString: "2"},
						{TeamID: "2", PointsForString: "200.0", RecordWinsString: "6", RecordLossesString: "4", RecordTiesString: "0"},
					},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{TeamID: "1", TeamName: Team1Name, OwnerName: Team1Owner,
						PointsForString: "100.0", RecordWinsString: "5", RecordLossesString: "3", RecordTiesString: "2",
						PointsFor: 100.0, RecordWins: 5, RecordLosses: 3, RecordTies: 2,
					},
					{TeamID: "2", TeamName: Team2Name, OwnerName: Team2Owner,
						PointsForString: "200.0", RecordWinsString: "6", RecordLossesString: "4", RecordTiesString: "0",
						PointsFor: 200.0, RecordWins: 6, RecordLosses: 4, RecordTies: 0,
					},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := associateStandingsWithFranchises(tc.franchiseDetailsResponse, tc.leagueStandingsResponse)
			if (err != nil) != tc.expectError {
				t.Errorf("associateStandingsWithFranchises() error = %v, expectError %v", err, tc.expectError)
				return
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Mismatch in test case %s:\n\n Expected %+v,\n\n Got      %+v \n\n", tc.name, tc.expected, result)
			}
		})
	}
}

func TestAppendAllPlay(t *testing.T) {
	testCases := []struct {
		name            string
		franchises      Franchises
		allPlayTeamData []AllPlayTeamStats
		expected        Franchises
		expectError     bool
	}{
		{
			name: "Test case 1: Valid data",
			franchises: Franchises{
				Franchise: []Franchise{
					{TeamName: Team1Name, OwnerName: Team1Owner, TeamID: "1", PointsForString: "100.0",
						RecordWinsString: "5", RecordLossesString: "3", RecordTiesString: "2"},
					{TeamName: Team2Name, OwnerName: Team2Owner, TeamID: "2", PointsForString: "200.0",
						RecordWinsString: "6", RecordLossesString: "4", RecordTiesString: "0"},
				},
			},
			allPlayTeamData: []AllPlayTeamStats{
				{FranchiseName: Team1Name, AllPlayWins: "5", AllPlayLosses: "3", AllPlayTies: "0",
					AllPlayPercentage: "62.5"},
				{FranchiseName: Team2Name, AllPlayWins: "7", AllPlayLosses: "1", AllPlayTies: "0",
					AllPlayPercentage: "87.5"},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{TeamName: Team1Name, OwnerName: Team1Owner, TeamID: "1", PointsForString: "100.0",
						RecordWinsString: "5", RecordLossesString: "3", RecordTiesString: "2",
						AllPlayWins: 5, AllPlayLosses: 3, AllPlayTies: 0, AllPlayPercentageString: "62.5",
						AllPlayPercentage: 62.5},
					{TeamName: Team2Name, OwnerName: Team2Owner, TeamID: "2", PointsForString: "200.0",
						RecordWinsString: "6", RecordLossesString: "4", RecordTiesString: "0",
						AllPlayWins: 7, AllPlayLosses: 1, AllPlayTies: 0, AllPlayPercentageString: "87.5",
						AllPlayPercentage: 87.5},
				},
			},
			expectError: false,
		},
		{
			name: "Test case 2: Invalid data",
			franchises: Franchises{
				Franchise: []Franchise{
					{TeamName: Team1Name},
					{TeamName: Team2Name},
				},
			},
			allPlayTeamData: []AllPlayTeamStats{
				{FranchiseName: Team1Name, AllPlayWins: "5", AllPlayLosses: "3", AllPlayTies: "0", AllPlayPercentage: "invalid"},
				{FranchiseName: Team2Name, AllPlayWins: "7", AllPlayLosses: "1", AllPlayTies: "0", AllPlayPercentage: "87.5"},
			},
			expected:    Franchises{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := appendAllPlay(tc.franchises, tc.allPlayTeamData)
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected:\n%v\nGot:\n%v", tc.expected, result)
			}
		})
	}
}

func TestCopyStandingsDetails(t *testing.T) {
	testCases := []struct {
		name      string
		franchise Franchise
		standing  Franchise
		expected  Franchise
		expectErr bool
	}{
		{
			name: "Valid standings details",
			franchise: Franchise{
				TeamID:    "1",
				TeamName:  Team1Name,
				OwnerName: Team1Owner,
			},
			standing: Franchise{
				RecordWinsString:   "10",
				RecordLossesString: "5",
				RecordTiesString:   "0",
				PointsForString:    "500.5",
			},
			expected: Franchise{
				TeamID:             "1",
				TeamName:           Team1Name,
				OwnerName:          Team1Owner,
				RecordWinsString:   "10",
				RecordLossesString: "5",
				RecordTiesString:   "0",
				PointsForString:    "500.5",
				RecordWins:         10,
				RecordLosses:       5,
				RecordTies:         0,
				PointsFor:          500.5,
			},
			expectErr: false,
		},
		{
			name: "Invalid standings details",
			franchise: Franchise{
				TeamID:    "1",
				TeamName:  Team1Name,
				OwnerName: Team1Owner,
			},
			standing: Franchise{
				RecordWinsString:   "ten",
				RecordLossesString: "five",
				RecordTiesString:   "zero",
				PointsForString:    "five hundred point five",
			},
			expected:  Franchise{},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := copyStandingsDetails(tc.franchise, tc.standing)
			if (err != nil) != tc.expectErr {
				t.Fatalf("Expected error status %v, got %v", tc.expectErr, err != nil)
			}
			if !tc.expectErr && result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestPopulateHeadToHeadRecords(t *testing.T) {
	testCases := []struct {
		name     string
		input    Franchises
		expected Franchises
	}{
		{
			name: "Test case 1",
			input: Franchises{
				Franchise: []Franchise{
					{RecordWins: 6, RecordLosses: 4, RecordTies: 2},
					{RecordWins: 3, RecordLosses: 7, RecordTies: 0},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{RecordWins: 6, RecordLosses: 4, RecordTies: 2, Record: "6-4-2"},
					{RecordWins: 3, RecordLosses: 7, RecordTies: 0, Record: "3-7-0"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := populateHeadToHeadRecords(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("populateHeadToHeadRecords() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestConvertStringToInteger(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{
			name:     "Test case 1: Valid integer string",
			input:    "123",
			expected: 123,
			hasError: false,
		},
		{
			name:     "Test case 2: Zero integer string",
			input:    "0",
			expected: 0,
			hasError: false,
		},
		{
			name:     "Test case 3: Invalid integer string",
			input:    "abc",
			expected: 0,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertStringToInteger(tc.input)
			if (err != nil) != tc.hasError {
				t.Errorf("convertStringToInteger(%v) error = %v, wantErr %v", tc.input, err, tc.hasError)
				return
			}
			if result != tc.expected {
				t.Errorf("convertStringToInteger(%v) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestPopulateAllPlayRecords(t *testing.T) {
	testCases := []struct {
		name     string
		input    Franchises
		expected Franchises
	}{
		{
			name: "Test case 1: Single franchise",
			input: Franchises{
				Franchise: []Franchise{
					{
						AllPlayWins:   10,
						AllPlayLosses: 5,
						AllPlayTies:   2,
					},
				},
			},
			expected: Franchises{
				Franchise: []Franchise{
					{
						AllPlayWins:   10,
						AllPlayLosses: 5,
						AllPlayTies:   2,
						AllPlayRecord: "10-5-2",
					},
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := populateAllPlayRecords(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("populateAllPlayRecords() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCheckResponseParity(t *testing.T) {
	testCases := []struct {
		name                    string
		leagueResponse          LeagueResponse
		leagueStandingsResponse LeagueStandingsResponse
		expectError             bool
	}{
		{
			name: "Test case 1: Equal number of franchises",
			leagueResponse: LeagueResponse{
				League: League{
					Franchises: Franchises{
						Franchise: []Franchise{
							{TeamName: Team1Name},
							{TeamName: Team2Name},
						},
					},
				},
			},
			leagueStandingsResponse: LeagueStandingsResponse{
				LeagueStandings: LeagueStandings{
					Franchise: []Franchise{
						{TeamName: Team1Name},
						{TeamName: Team2Name},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Test case 2: Unequal number of franchises",
			leagueResponse: LeagueResponse{
				League: League{
					Franchises: Franchises{
						Franchise: []Franchise{
							{TeamName: Team1Name},
						},
					},
				},
			},
			leagueStandingsResponse: LeagueStandingsResponse{
				LeagueStandings: LeagueStandings{
					Franchise: []Franchise{
						{TeamName: Team1Name},
						{TeamName: Team2Name},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkResponseParity(tc.leagueResponse, tc.leagueStandingsResponse)
			if (err != nil) != tc.expectError {
				t.Errorf("checkResponseParity() error = %v, expectError %v", err, tc.expectError)
				return
			}
		})
	}
}

func TestNewCollector(t *testing.T) {
	c := newCollector()
	if c == nil {
		t.Errorf("newCollector() = %v, want non-nil", c)
	}
	if reflect.TypeOf(c) != reflect.TypeOf(&colly.Collector{}) {
		t.Errorf("newCollector() = %T, want *colly.Collector", c)
	}
}

// MockHTMLElement is a mock of colly.HTMLElement.
type MockHTMLElement struct {
	mock.Mock
}

func (m *MockHTMLElement) NewHTMLElementFromSelectionNode(resp *colly.Response, s *goquery.Selection, n *html.Node, idx int) *colly.HTMLElement {
	args := m.Called(resp, s, n, idx)
	return args.Get(0).(*colly.HTMLElement)
}

func (m *MockHTMLElement) Attr(k string) string {
	args := m.Called(k)
	return args.String(0)
}

func (m *MockHTMLElement) ChildAttr(goquerySelector, attrName string) string {
	args := m.Called(goquerySelector, attrName)
	return args.String(0)
}

func (m *MockHTMLElement) ChildText(goquerySelector string) string {
	args := m.Called(goquerySelector)
	return args.String(0)
}

func (m *MockHTMLElement) ChildAttrs(goquerySelector, attrName string) []string {
	args := m.Called(goquerySelector, attrName)
	return args.Get(0).([]string)
}

func (m *MockHTMLElement) ForEach(goquerySelector string, callback func(int, *colly.HTMLElement)) {
	m.Called(goquerySelector, callback)
}

func (m *MockHTMLElement) ForEachWithBreak(goquerySelector string, callback func(int, *colly.HTMLElement) bool) {
	m.Called(goquerySelector, callback)
}

func (m *MockHTMLElement) Unmarshal(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func TestParseRow(t *testing.T) {
	mockHTMLElement := new(MockHTMLElement)

	// Setup expectations
	mockHTMLElement.On("ChildText", "td:nth-child(1)").Return("Test Franchise")
	mockHTMLElement.On("ChildText", "td:nth-child(13)").Return("10")
	mockHTMLElement.On("ChildText", "td:nth-child(14)").Return("5")
	mockHTMLElement.On("ChildText", "td:nth-child(15)").Return("2")
	mockHTMLElement.On("ChildText", "td:nth-child(16)").Return("0.66")

	// Call the function with the mock
	result := parseRow(mockHTMLElement)

	// Assert that the expectations were met
	mockHTMLElement.AssertExpectations(t)

	// Assert that the result is what you expect
	if result.FranchiseName != "Test Franchise" {
		t.Errorf("Expected FranchiseName to be 'Test Franchise', got '%s'", result.FranchiseName)
	}

	if result.AllPlayWins != "10" {
		t.Errorf("Expected AllPlayWins to be '10', got '%s'", result.AllPlayWins)
	}

	if result.AllPlayLosses != "5" {
		t.Errorf("Expected AllPlayLosses to be '5', got '%s'", result.AllPlayLosses)
	}

	if result.AllPlayTies != "2" {
		t.Errorf("Expected AllPlayTies to be '2', got '%s'", result.AllPlayTies)
	}

	if result.AllPlayPercentage != "0.66" {
		t.Errorf("Expected AllPlayPercentage to be '0.66', got '%s'", result.AllPlayPercentage)
	}
}

func TestFilterTeams(t *testing.T) {
	// Create a slice of AllPlayTeamStats
	allPlayTeamsStats := []AllPlayTeamStats{
		{FranchiseName: Team1Name},
		{FranchiseName: "2nd Team"},
		{FranchiseName: "Third Team"},
		{FranchiseName: "_Fourth Team"},
	}

	// Call filterTeams
	result := filterTeams(allPlayTeamsStats)

	// Check that the result only includes the AllPlayTeamStats whose FranchiseName starts with a letter
	expected := []AllPlayTeamStats{
		{FranchiseName: Team1Name},
		{FranchiseName: "Third Team"},
	}
	if len(result) != len(expected) {
		t.Fatalf("Expected result length to be %d, got %d", len(expected), len(result))
	}
	for i, v := range result {
		if v.FranchiseName != expected[i].FranchiseName {
			t.Errorf("Expected FranchiseName at index %d to be '%s', got '%s'", i, expected[i].FranchiseName, v.FranchiseName)
		}
	}
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestGetFranchiseDetails(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	leagueAPIURL := "http://example.com"

	testLeagueResponse, _ := json.Marshal(LeagueResponse{
		League: League{
			Name: "fantasmo",
		},
	})
	t.Run("successful request", func(t *testing.T) {
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(string(testLeagueResponse))),
		}
		mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

		result, err := getFranchiseDetails(mockHTTPClient, leagueAPIURL)
		fmt.Printf("%+v\n", result)
		mockHTTPClient.AssertExpectations(t)

		if err != nil {
			t.Errorf("Expected no error, got '%s'", err)
		}

		if result.League.Name != "fantasmo" {
			t.Errorf("Expected result to be '%s', got '%s'", "fantasmo", result.League.Name)
		}
	})

	t.Run("failed request", func(t *testing.T) {
		// Setup expectations
		mockHTTPClient.On("Do", mock.Anything).Return(nil, errors.New("network error"))

		// Call the function with the mock
		_, err := getFranchiseDetails(mockHTTPClient, leagueAPIURL)

		// Assert that the expectations were met
		mockHTTPClient.AssertExpectations(t)

		// Assert that an error was returned
		if err == nil {
			t.Error("Expected an error, got nil")
		}
	})
}

func TestGetLeagueStandings(t *testing.T) {
	mockHTTPClient := new(MockHTTPClient)
	leagueAPIURL := "http://example.com"

	testLeagueStandings, _ := json.Marshal(LeagueStandingsResponse{
		LeagueStandings: LeagueStandings{
			Franchise: []Franchise{
				{TeamID: "1", TeamName: Team1Name, OwnerName: Team1Owner, RecordWinsString: "10", RecordLossesString: "5", RecordTiesString: "2", PointsForString: "0.66"},
			},
		},
	})

	t.Run("successful request", func(t *testing.T) {
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(string(testLeagueStandings))),
		}
		mockHTTPClient.On("Do", mock.Anything).Return(mockResponse, nil)

		result, err := getLeagueStandings(mockHTTPClient, leagueAPIURL)
		fmt.Printf("%+v", result)
		mockHTTPClient.AssertExpectations(t)

		if err != nil {
			t.Errorf("Expected no error, got '%s'", err)
		}

		if result.LeagueStandings.Franchise[0].TeamID != "1" {
			t.Errorf("Expected result to be '%s', got '%s'", "1", result.LeagueStandings.Franchise[0].TeamID)
		}
	})

	t.Run("failed request", func(t *testing.T) {
		// Setup expectations
		mockHTTPClient.On("Do", mock.Anything).Return(nil, errors.New("network error"))

		// Call the function with the mock
		_, err := getFranchiseDetails(mockHTTPClient, leagueAPIURL)

		// Assert that the expectations were met
		mockHTTPClient.AssertExpectations(t)

		// Assert that an error was returned
		if err == nil {
			t.Error("Expected an error, got nil")
		}
	})
}
