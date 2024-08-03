package main

import (
	"testing"
)

var CASBAN_BAD_ID = 7156574451
var LOLSBOT_BAD_ID = 6656436060
var GOOD_ID = 5485817729

func Test_isUserApiBanned(t *testing.T) {
	type args struct {
		userid int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"All_CasBanTrue", args{userid: CASBAN_BAD_ID}, true},
		{"All_CasBanFalse", args{userid: GOOD_ID}, false},
		{"All_LolsBanTrue", args{userid: LOLSBOT_BAD_ID}, true},
		{"All_LolsBanFalse", args{userid: GOOD_ID}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUserApiBanned(tt.args.userid); got != tt.want {
				t.Errorf("isUserApiBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isUserCasBanned(t *testing.T) {
	type args struct {
		userid int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"CasBanTrue", args{userid: CASBAN_BAD_ID}, true},
		{"CasBanFalse", args{userid: GOOD_ID}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUserCasBanned(tt.args.userid); got != tt.want {
				t.Errorf("isUserCasBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isUserLolsBanned(t *testing.T) {
	type args struct {
		userid int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"LolsBanTrue", args{userid: LOLSBOT_BAD_ID}, true},
		{"LolsBanFalse", args{userid: GOOD_ID}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUserLolsBanned(tt.args.userid); got != tt.want {
				t.Errorf("isUserLolsBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}
