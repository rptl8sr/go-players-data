package model

import "time"

type Player struct {
	Number       int       `json:"number"`
	ID           int       `json:"ID"`
	GroupName    string    `json:"groupName"`
	PlayerName   string    `json:"panelName"`
	Tags         []string  `json:"tags"`
	ScheduleName string    `json:"scheduleName"`
	TimeZoneDiff int       `json:"timeZoneDiff"`
	LastOnline   time.Time `json:"lastOnline"`
	Serial       string    `json:"serial"`
	MAC          string    `json:"MAC"`
	IP           string    `json:"IP"`
	Type         string    `json:"type"`
	Model        string    `json:"model"`
	Version      string    `json:"version"`
	StoreNumber  int       `json:"storeNumber"`
	CompanyName  string    `json:"companyName"`
}

type PlayerReceive struct {
	Number       int    `json:"number"`
	ID           string `json:"id"`
	GroupName    string `json:"group_name"`
	PlayerName   string `json:"panel_name"`
	Tags         string `json:"f_tag"`
	ScheduleName string `json:"schedule_name"`
	TimeZoneDiff string `json:"timezone_diff"`
	LastOnline   string `json:"last_online"`
	Serial       string `json:"serial"`
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Type         string `json:"type"`
	Model        string `json:"model"`
	Version      string `json:"v"`
}
