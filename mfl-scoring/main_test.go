package main

import (
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
				{RecordWins: 6, RecordTies: 1},
				{RecordWins: 3, RecordTies: 7},
				{RecordWins: 1, RecordTies: 0},
			},
			expected: Franchises{
				{RecordWins: 6, RecordTies: 1, RecordMagic: 6.5},
				{RecordWins: 3, RecordTies: 7, RecordMagic: 6.5},
				{RecordWins: 1, RecordTies: 0, RecordMagic: 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateRecordMagic(tc.franchises)
			for i := range result {
				if result[i].RecordMagic != tc.expected[i].RecordMagic {
					t.Errorf("Mismatch in test case %s for franchise %d: Expected %f, got %f",
						tc.name, i, tc.expected[i].RecordMagic, result[i].RecordMagic)
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
				{PointScore: 3, RecordScore: 4.5},
				{PointScore: 7, RecordScore: 9},
				{PointScore: 2, RecordScore: 1.5},
			},
			expected: Franchises{
				Franchise{PointScore: 3, RecordScore: 4.5, TotalScore: "7.5"},
				Franchise{PointScore: 7, RecordScore: 9, TotalScore: "16.0"},
				Franchise{PointScore: 2, RecordScore: 1.5, TotalScore: "3.5"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateTotalScore(tc.franchises)
			for i := range result {
				if result[i].TotalScore != tc.expected[i].TotalScore {
					t.Errorf("Mismatch in test case %s for franchise %d: Expected %s, got %s",
						tc.name, i, tc.expected[i].TotalScore, result[i].TotalScore)
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
				{RecordMagic: 8.5},
				{RecordMagic: 8.5},
				{RecordMagic: 7},
				{RecordMagic: 5},
			},
			expected: Franchises{
				{RecordMagic: 8.5, RecordScore: 3.5, RecordScoreString: "3.5"},
				{RecordMagic: 8.5, RecordScore: 3.5, RecordScoreString: "3.5"},
				{RecordMagic: 7, RecordScore: 2, RecordScoreString: "2.0"},
				{RecordMagic: 5, RecordScore: 1, RecordScoreString: "1.0"},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateRecordScore(tc.franchises)
			for i := range result {
				if result[i].RecordScore != tc.expected[i].RecordScore {
					t.Errorf("Mismatch in test case %s for franchise %d in field %s: Expected %f, got %f",
						tc.name, i, "RecordScore", tc.expected[i].RecordScore, result[i].RecordScore)
				}
				if result[i].RecordScoreString != tc.expected[i].RecordScoreString {
					t.Errorf("Mismatch in test case %s for franchise %d in field %s: Expected %s, got %s",
						tc.name, i, "RecordScoreString", tc.expected[i].RecordScoreString, result[i].RecordScoreString)
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
				{PointsFor: 15},
				{PointsFor: 15},
				{PointsFor: 10},
				{PointsFor: 5},
			},
			expected: Franchises{
				{PointsFor: 15, PointScore: 3.5, PointScoreString: "3.5"},
				{PointsFor: 15, PointScore: 3.5, PointScoreString: "3.5"},
				{PointsFor: 10, PointScore: 2, PointScoreString: "2.0"},
				{PointsFor: 5, PointScore: 1, PointScoreString: "1.0"},
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculatePointsScore(tc.franchises)
			for i := range result {
				if result[i].PointScore != tc.expected[i].PointScore {
					t.Errorf("Mismatch in test case %s for franchise %d: Expected %f, got %f",
						tc.name, i, tc.expected[i].PointScore, result[i].PointScore)
				}
			}
		})
	}
}

// func TestCalculatePointsScoreBasic(t *testing.T) {
// 	in := Franchises{
// 		{
// 			"0001",
// 			"Viagravators 3.0 Harder, Better, Faster, Stronger!",
// 			"Curry",
// 			4,
// 			1,
// 			0,
// 			538.5,
// 			"538.5",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0007",
// 			"Keep it Civil",
// 			"Falcone",
// 			4,
// 			1,
// 			0,
// 			493.6,
// 			"493.6",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0010",
// 			"FUN is a 3-letter Word",
// 			"Ian",
// 			2,
// 			3,
// 			0,
// 			479.5,
// 			"479.5",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0003",
// 			"ding dong merrily on high",
// 			"Kluck",
// 			3,
// 			2,
// 			0,
// 			399.8,
// 			"399.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0004",
// 			"Grentest Of All Time",
// 			"Grantass",
// 			2,
// 			3,
// 			0,
// 			395.2,
// 			"395.2",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0006",
// 			"Your Gun Is Digging Into My Hip",
// 			"Nolte",
// 			3,
// 			2,
// 			0,
// 			385.9,
// 			"385.9",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0009",
// 			"Sweet Chocolate James Dazzle",
// 			"James",
// 			3,
// 			2,
// 			0,
// 			375,
// 			"375.0",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0005",
// 			"Team Tang Bang",
// 			"mathew bonner",
// 			3,
// 			2,
// 			0,
// 			368.2,
// 			"368.2",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0008",
// 			"Fat Lady Sung",
// 			"Tim Kelsch",
// 			1,
// 			4,
// 			0,
// 			352,
// 			"352.0",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0002",
// 			"Touchdown My Pants",
// 			"Dustin Picasso ",
// 			0,
// 			5,
// 			0,
// 			338.8,
// 			"338.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 	}
// 	expectedJSONStr := `
// 	[
// 		{
// 			"TeamID": "0001",
// 			"TeamName": "Viagravators 3.0 Harder, Better, Faster, Stronger!",
// 			"OwnerName": "Curry",
// 			"RecordWins": 4,
// 			"RecordLosses": 1,
// 			"RecordTies": 0,
// 			"PointsFor": 538.5,
// 			"PointsForString": "538.5",
// 			"PointScore": 10,
// 			"PointScoreString": "10.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0007",
// 			"TeamName": "Keep it Civil",
// 			"OwnerName": "Falcone",
// 			"RecordWins": 4,
// 			"RecordLosses": 1,
// 			"RecordTies": 0,
// 			"PointsFor": 493.6,
// 			"PointsForString": "493.6",
// 			"PointScore": 9,
// 			"PointScoreString": "9.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0010",
// 			"TeamName": "FUN is a 3-letter Word",
// 			"OwnerName": "Ian",
// 			"RecordWins": 2,
// 			"RecordLosses": 3,
// 			"RecordTies": 0,
// 			"PointsFor": 479.5,
// 			"PointsForString": "479.5",
// 			"PointScore": 8,
// 			"PointScoreString": "8.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0003",
// 			"TeamName": "ding dong merrily on high",
// 			"OwnerName": "Kluck",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 399.8,
// 			"PointsForString": "399.8",
// 			"PointScore": 7,
// 			"PointScoreString": "7.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0004",
// 			"TeamName": "Grentest Of All Time",
// 			"OwnerName": "Grantass",
// 			"RecordWins": 2,
// 			"RecordLosses": 3,
// 			"RecordTies": 0,
// 			"PointsFor": 395.2,
// 			"PointsForString": "395.2",
// 			"PointScore": 6,
// 			"PointScoreString": "6.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0006",
// 			"TeamName": "Your Gun Is Digging Into My Hip",
// 			"OwnerName": "Nolte",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 385.9,
// 			"PointsForString": "385.9",
// 			"PointScore": 5,
// 			"PointScoreString": "5.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0009",
// 			"TeamName": "Sweet Chocolate James Dazzle",
// 			"OwnerName": "James",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 375,
// 			"PointsForString": "375.0",
// 			"PointScore": 4,
// 			"PointScoreString": "4.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0005",
// 			"TeamName": "Team Tang Bang",
// 			"OwnerName": "mathew bonner",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 368.2,
// 			"PointsForString": "368.2",
// 			"PointScore": 3,
// 			"PointScoreString": "3.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0008",
// 			"TeamName": "Fat Lady Sung",
// 			"OwnerName": "Tim Kelsch",
// 			"RecordWins": 1,
// 			"RecordLosses": 4,
// 			"RecordTies": 0,
// 			"PointsFor": 352,
// 			"PointsForString": "352.0",
// 			"PointScore": 2,
// 			"PointScoreString": "2.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0002",
// 			"TeamName": "Touchdown My Pants",
// 			"OwnerName": "Dustin Picasso ",
// 			"RecordWins": 0,
// 			"RecordLosses": 5,
// 			"RecordTies": 0,
// 			"PointsFor": 338.8,
// 			"PointsForString": "338.8",
// 			"PointScore": 1,
// 			"PointScoreString": "1.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		}
// 	]
// 	`

// 	out := calculatePointsScore(in)

// 	if !JSONCompare(t, out, expectedJSONStr) {
// 		sort.Sort(ByPointsFor{out})
// 		fmt.Println("TT", printScoringTableUncouthly(out))
// 		// fmt.Println(printTeam(out))
// 	}
// }

// func TestCalculatePointsScoreWithTies(t *testing.T) {
// 	in := Franchises{
// 		{
// 			"0001",
// 			"Viagravators 3.0 Harder, Better, Faster, Stronger!",
// 			"Curry",
// 			4,
// 			1,
// 			0,
// 			538.5,
// 			"538.5",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0007",
// 			"Keep it Civil",
// 			"Falcone",
// 			4,
// 			1,
// 			0,
// 			538.5,
// 			"538.5",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0010",
// 			"FUN is a 3-letter Word",
// 			"Ian",
// 			2,
// 			3,
// 			0,
// 			538.5,
// 			"538.5",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0003",
// 			"ding dong merrily on high",
// 			"Kluck",
// 			3,
// 			2,
// 			0,
// 			399.8,
// 			"399.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0004",
// 			"Grentest Of All Time",
// 			"Grantass",
// 			2,
// 			3,
// 			0,
// 			399.8,
// 			"399.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0006",
// 			"Your Gun Is Digging Into My Hip",
// 			"Nolte",
// 			3,
// 			2,
// 			0,
// 			399.8,
// 			"399.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0009",
// 			"Sweet Chocolate James Dazzle",
// 			"James",
// 			3,
// 			2,
// 			0,
// 			399.8,
// 			"399.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0005",
// 			"Team Tang Bang",
// 			"mathew bonner",
// 			3,
// 			2,
// 			0,
// 			399.8,
// 			"399.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0008",
// 			"Fat Lady Sung",
// 			"Tim Kelsch",
// 			1,
// 			4,
// 			0,
// 			352,
// 			"352.0",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 		{
// 			"0002",
// 			"Touchdown My Pants",
// 			"Dustin Picasso ",
// 			0,
// 			5,
// 			0,
// 			338.8,
// 			"338.8",
// 			0,
// 			"",
// 			0,
// 			0,
// 			"",
// 			"",
// 			0,
// 			0,
// 			0,
// 			"",
// 		},
// 	}
// 	expectedJSONStr := `
// 	[
// 		{
// 			"TeamID": "0001",
// 			"TeamName": "Viagravators 3.0 Harder, Better, Faster, Stronger!",
// 			"OwnerName": "Curry",
// 			"RecordWins": 4,
// 			"RecordLosses": 1,
// 			"RecordTies": 0,
// 			"PointsFor": 538.5,
// 			"PointsForString": "538.5",
// 			"PointScore": 9,
// 			"PointScoreString": "9.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0007",
// 			"TeamName": "Keep it Civil",
// 			"OwnerName": "Falcone",
// 			"RecordWins": 4,
// 			"RecordLosses": 1,
// 			"RecordTies": 0,
// 			"PointsFor": 538.5,
// 			"PointsForString": "538.5",
// 			"PointScore": 9,
// 			"PointScoreString": "9.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0010",
// 			"TeamName": "FUN is a 3-letter Word",
// 			"OwnerName": "Ian",
// 			"RecordWins": 2,
// 			"RecordLosses": 3,
// 			"RecordTies": 0,
// 			"PointsFor": 538.5,
// 			"PointsForString": "538.5",
// 			"PointScore": 9,
// 			"PointScoreString": "9.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0003",
// 			"TeamName": "ding dong merrily on high",
// 			"OwnerName": "Kluck",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 399.8,
// 			"PointsForString": "399.8",
// 			"PointScore": 5,
// 			"PointScoreString": "5.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0004",
// 			"TeamName": "Grentest Of All Time",
// 			"OwnerName": "Grantass",
// 			"RecordWins": 2,
// 			"RecordLosses": 3,
// 			"RecordTies": 0,
// 			"PointsFor": 399.8,
// 			"PointsForString": "399.8",
// 			"PointScore": 5,
// 			"PointScoreString": "5.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0006",
// 			"TeamName": "Your Gun Is Digging Into My Hip",
// 			"OwnerName": "Nolte",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 399.8,
// 			"PointsForString": "399.8",
// 			"PointScore": 5,
// 			"PointScoreString": "5.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0009",
// 			"TeamName": "Sweet Chocolate James Dazzle",
// 			"OwnerName": "James",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 399.8,
// 			"PointsForString": "399.8",
// 			"PointScore": 5,
// 			"PointScoreString": "5.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0005",
// 			"TeamName": "Team Tang Bang",
// 			"OwnerName": "mathew bonner",
// 			"RecordWins": 3,
// 			"RecordLosses": 2,
// 			"RecordTies": 0,
// 			"PointsFor": 399.8,
// 			"PointsForString": "399.8",
// 			"PointScore": 5,
// 			"PointScoreString": "5.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0008",
// 			"TeamName": "Fat Lady Sung",
// 			"OwnerName": "Tim Kelsch",
// 			"RecordWins": 1,
// 			"RecordLosses": 4,
// 			"RecordTies": 0,
// 			"PointsFor": 352,
// 			"PointsForString": "352.0",
// 			"PointScore": 2,
// 			"PointScoreString": "2.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		},
// 		{
// 			"TeamID": "0002",
// 			"TeamName": "Touchdown My Pants",
// 			"OwnerName": "Dustin Picasso ",
// 			"RecordWins": 0,
// 			"RecordLosses": 5,
// 			"RecordTies": 0,
// 			"PointsFor": 338.8,
// 			"PointsForString": "338.8",
// 			"PointScore": 1,
// 			"PointScoreString": "1.0",
// 			"RecordMagic": 0,
// 			"RecordScore": 0,
// 			"RecordScoreString": "",
// 			"TotalScore": "",
// 			"AllPlayWins": 0,
// 			"AllPlayLosses": 0,
// 			"AllPlayTies": 0,
// 			"AllPlayPercentage": ""
// 		}
// 	]
// 	`

// 	out := calculatePointsScore(in)

// 	if !JSONCompare(t, out, expectedJSONStr) {
// 		sort.Sort(ByPointsFor{out})
// 		fmt.Println(printScoringTableUncouthly(out))
// 	}
// }
