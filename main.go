package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

type GameSchedule struct {
	Date      string `json:"date"`
	Day       string `json:"day"`
	Time      string `json:"time"`
	Location  string `json:"location"`
	GameNo    string `json:"game_no"`
	AwayTeam  string `json:"away_team"`
	AwaySP    string `json:"away_sp"`
	AwayScore string `json:"away_score"`
	HomeTeam  string `json:"home_team"`
	HomeScore string `json:"home_score"`
	HomeSP    string `json:"home_sp"`
}

// FetchSchedule fetches the schedule from CPBL website based on the year, month, and game type
func FetchSchedule() ([]GameSchedule, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var jsonStr string

	// Define the URL
	url := "https://cpbl.com.tw"

	// Run chromedp tasks
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second), // Wait for table to load
		chromedp.Evaluate(`
				(() => {
						let schedules = [];
						function getTeamName (team_url) {
							switch(team_url) {
								case '/team/index?teamNo=ACN011':
									return '中信兄弟'
									break
								case '/team/index?teamNo=AEO011':
									return '富邦悍將'
									break
								case '/team/index?teamNo=AJL011':
									return '樂天桃猿'
									break
								case '/team/index?teamNo=ADD011':
									return '統一獅'
									break
								case '/team/index?teamNo=AKP011':
									return '台鋼雄鷹'
									break
								case '/team/index?teamNo=AAA011':
									return '味全龍'
									break
								default:
									break;
							}
						}
						const a = document.querySelectorAll('.major');
						a.forEach((dateDiv) => {
								const b = dateDiv.querySelectorAll('.game_item')
								const date = document.querySelector('.date').innerText.trim()
								const day = document.querySelector('.day').innerText.trim()
								b.forEach((gameDiv) => {
										const time = gameDiv.querySelector('.time').innerText.trim()
										const location = gameDiv.querySelector('.place').innerText.trim()
										const game_no = gameDiv.querySelector('.game_no a').innerText.trim()
										let away_team = gameDiv.querySelector('.away .team_name a').getAttribute('href').trim()
										away_team = getTeamName(away_team)
										let home_team = gameDiv.querySelector('.home .team_name a').getAttribute('href').trim()
										home_team = getTeamName(home_team)
										const away_score = gameDiv.querySelector('.score_wrap .away')?.innerText.trim() || "0"
										const home_score = gameDiv.querySelector('.score_wrap .home')?.innerText.trim() || "0"
										const away_sp = gameDiv.querySelector('.away_sp .name a')?.innerText.trim()
										const home_sp = gameDiv.querySelector('.home_sp .name a')?.innerText.trim()
										schedules.push({ 
											'date': date,
											'day': day,
											'time': time,
											'location': location,
											'game_no': game_no,
											'away_team': away_team,
											'away_sp': away_sp,
											'away_score': away_score,
											'home_team': home_team,
											'home_score': home_score,
											'home_sp': home_sp
										});
								})
						})
						return JSON.stringify(schedules);
				})();
		`, &jsonStr),
	)

	if err != nil {
		return nil, err
	}

	var schedules []GameSchedule
	err = json.Unmarshal([]byte(jsonStr), &schedules)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func main() {
	schedules, err := FetchSchedule()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 格式化 JSON 輸出
	jsonData, err := json.MarshalIndent(schedules, "", "    ")
	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return
	}

	fmt.Println(string(jsonData))
}
