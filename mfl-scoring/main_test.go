package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
							{TeamID: "1", TeamName: "Team 1", OwnerName: "Owner 1"},
							{TeamID: "2", TeamName: "Team 2", OwnerName: "Owner 2"},
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
					{TeamID: "1", TeamName: "Team 1", OwnerName: "Owner 1",
						PointsForString: "100.0", RecordWinsString: "5", RecordLossesString: "3", RecordTiesString: "2",
						PointsFor: 100.0, RecordWins: 5, RecordLosses: 3, RecordTies: 2,
					},
					{TeamID: "2", TeamName: "Team 2", OwnerName: "Owner 2",
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
				t.Errorf("Mismatch in test case %s: Expected %+v, got %+v", tc.name, tc.expected, result)
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
