package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

type RawGameSchedule struct {
	Date        string `json:"date"`
	Day         string `json:"day"`
	Time        string `json:"time"`
	Location    string `json:"location"`
	GameNo      string `json:"game_no"`
	AwayTeamURL string `json:"away_team_url"`
	AwaySP      string `json:"away_sp"`
	AwayScore   string `json:"away_score"`
	HomeTeamURL string `json:"home_team_url"`
	HomeSP      string `json:"home_sp"`
	HomeScore   string `json:"home_score"`
}

type GameSchedule struct {
	Date           string `json:"date"`
	Day            string `json:"day"`
	Time           string `json:"time"`
	Location       string `json:"location"`
	GameNo         string `json:"game_no"`
	AwayTeam       string `json:"away_team"`
	AwayTeamAbbr   string `json:"away_team_abbr"`
	AwayTeamSingle string `json:"away_team_single"`
	AwaySP         string `json:"away_sp"`
	AwayScore      string `json:"away_score"`
	HomeTeam       string `json:"home_team"`
	HomeTeamAbbr   string `json:"home_team_abbr"`
	HomeTeamSingle string `json:"home_team_single"`
	HomeScore      string `json:"home_score"`
	HomeSP         string `json:"home_sp"`
}

type RawGameBox struct {
	Date         string `json:"date"`
	Day          string `json:"day"`
	Location     string `json:"location"`
	GameNo       string `json:"game_no"`
	AwayTeamURL  string `json:"away_team_url"`
	AwayScore    string `json:"away_score"`
	HomeTeamURL  string `json:"home_team_url"`
	HomeScore    string `json:"home_score"`
	WinsPitcher  string `json:"wins_pitcher"`
	LosesPitcher string `json:"loses_pitcher"`
	SavesPitcher string `json:"saves_pitcher"`
}

type GameBox struct {
	Date           string `json:"date"`
	Day            string `json:"day"`
	Location       string `json:"location"`
	GameNo         string `json:"game_no"`
	AwayTeam       string `json:"away_team"`
	AwayTeamAbbr   string `json:"away_team_abbr"`
	AwayTeamSingle string `json:"away_team_single"`
	AwayScore      string `json:"away_score"`
	HomeTeam       string `json:"home_team"`
	HomeTeamAbbr   string `json:"home_team_abbr"`
	HomeTeamSingle string `json:"home_team_single"`
	HomeScore      string `json:"home_score"`
	WinsPitcher    string `json:"wins_pitcher"`
	LosesPitcher   string `json:"loses_pitcher"`
	SavesPitcher   string `json:"saves_pitcher"`
}

type Team struct {
	FullName string
	AbbrName string
	OneName  string
}

var (
	schedules []GameSchedule
	mu        sync.Mutex // 互斥鎖，確保數據更新時不會產生競爭
)

func getTeamName(teamURL string) Team {
	var team Team
	switch teamURL {
	case "/team/index?teamNo=ACN011":
		team = Team{FullName: "中信兄弟", AbbrName: "中信", OneName: "象"}
		return team
	case "/team/index?teamNo=AEO011":
		team = Team{FullName: "富邦悍將", AbbrName: "富邦", OneName: "悍"}
		return team
	case "/team/index?teamNo=AJL011":
		team = Team{FullName: "樂天桃猿", AbbrName: "樂天", OneName: "猿"}
		return team
	case "/team/index?teamNo=ADD011":
		team = Team{FullName: "統一獅", AbbrName: "統一", OneName: "獅"}
		return team
	case "/team/index?teamNo=AKP011":
		team = Team{FullName: "台鋼雄鷹", AbbrName: "台鋼", OneName: "鷹"}
		return team
	case "/team/index?teamNo=AAA011":
		team = Team{FullName: "味全龍", AbbrName: "味全", OneName: "龍"}
		return team
	default:
		return Team{FullName: "未知", AbbrName: "未知", OneName: "未知"}
	}
}

func FetchSchedule() ([]GameSchedule, error) {
	// 建立一個新的 chromedp 瀏覽器上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	// 並透過 defer cancel() 確保任務結束時能關閉這個 context
	defer cancel()

	// 定義一個字串變數，用來接收從瀏覽器端 JavaScript 回傳的 JSON 字串
	var jsonStr string

	// 設定要瀏覽的 CPBL 官網 URL
	url := "https://cpbl.com.tw"

	// 使用 chromedp.Run 執行一系列任務，傳入前面創建的 context
	err := chromedp.Run(ctx,
		// 這一行會打開網址，也就是載入 CPBL 首頁
		chromedp.Navigate(url),
		// 等待 2 秒，給瀏覽器時間載入頁面中的 JavaScript 動態內容，尤其是比賽表格
		chromedp.Sleep(2*time.Second),
		// 這邊開始使用 Evaluate 執行一段瀏覽器中的 JavaScript，這段 JS 會在頁面 DOM 上操作並組成資料，最後回傳 JSON
		chromedp.Evaluate(`
            (() => {
                    let schedules = [];
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
                                    let home_team = gameDiv.querySelector('.home .team_name a').getAttribute('href').trim()
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
                                        'away_team_url': away_team,
                                        'away_sp': away_sp,
                                        'away_score': away_score,
                                        'home_team_url': home_team,
                                        'home_score': home_score,
                                        'home_sp': home_sp
                                    });
                            })
                    })
                    // 結束迴圈並回傳 JSON 字串
                    return JSON.stringify(schedules);
            })();
        `, &jsonStr), // 將剛剛那段 JavaScript 的回傳值存在 jsonStr 中
	)

	// 如果在 chromedp.Run 過程中有錯誤，就返回錯誤
	if err != nil {
		return nil, err
	}

	// 宣告一個 GameSchedule 結構的 slice，然後把剛剛從前端抓下來的 JSON 字串解析成 Go 的結構
	var schedules []RawGameSchedule
	if err := json.Unmarshal([]byte(jsonStr), &schedules); err != nil {
		return nil, err
	}

	var final []GameSchedule
	for _, item := range schedules {
		final = append(final, GameSchedule{
			Date:           item.Date,
			Day:            item.Day,
			Time:           item.Time,
			Location:       item.Location,
			GameNo:         item.GameNo,
			AwayTeam:       getTeamName(item.AwayTeamURL).FullName,
			AwayTeamAbbr:   getTeamName(item.AwayTeamURL).AbbrName,
			AwayTeamSingle: getTeamName(item.AwayTeamURL).OneName,
			AwayScore:      item.AwayScore,
			AwaySP:         item.AwaySP,
			HomeTeam:       getTeamName(item.HomeTeamURL).FullName,
			HomeTeamAbbr:   getTeamName(item.HomeTeamURL).AbbrName,
			HomeTeamSingle: getTeamName(item.HomeTeamURL).OneName,
			HomeScore:      item.HomeScore,
			HomeSP:         item.HomeSP,
		})
	}

	// 最後回傳整理好的比賽資料與 nil（代表沒有錯誤）
	return final, nil
}

// 定義一個名為 getScheduleHandler 的函式，w 是用來寫入 HTTP 回應的物件，r 是用來讀取請求資料的物件
func getScheduleHandler(w http.ResponseWriter, r *http.Request) {
	// 這是使用一個 sync.Mutex 來確保 schedules 這份資料在多個 Goroutine 同時讀寫時的同步安全
	mu.Lock()
	defer mu.Unlock()
	// 設定回應的 HTTP 標頭，告訴前端這是一份 JSON 格式的資料
	w.Header().Set("Content-Type", "application/json")
	// 把 schedules（通常是一個 []GameSchedule）編碼成 JSON 格式並寫入 HTTP 回應中送給前端
	json.NewEncoder(w).Encode(schedules)
}

// 定義一個名為 FetchYesterdaySchedule 的函式，返回值是 []GameSchedule（一個比賽排程的 slice）以及一個錯誤（error）
func FetchYesterdaySchedule() ([]GameBox, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var jsonStr string
	url := "https://cpbl.com.tw"

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.prev`, chromedp.ByQuery), // 等待「前一天」按鈕出現
		chromedp.Click(`.prev`, chromedp.ByQuery),       // 點擊「前一天」按鈕
		chromedp.Sleep(2*time.Second),                   // 給 JS 時間載入內容
		chromedp.Evaluate(`( () => {
			let schedules = [];
			const a = document.querySelectorAll('.major');
			a.forEach((dateDiv) => {
				const b = dateDiv.querySelectorAll('.game_item')
				const date = document.querySelector('.date').innerText.trim()
				const day = document.querySelector('.day').innerText.trim()
				b.forEach((gameDiv) => {
					const time = gameDiv.querySelector('.time')?.innerText.trim()
					const location = gameDiv.querySelector('.place').innerText.trim()
					const game_no = gameDiv.querySelector('.game_no a').innerText.trim()
					let away_team = gameDiv.querySelector('.away .team_name a').getAttribute('href').trim()
					let home_team = gameDiv.querySelector('.home .team_name a').getAttribute('href').trim()
					const away_score = gameDiv.querySelector('.score_wrap .away')?.innerText.trim() || "0"
					const home_score = gameDiv.querySelector('.score_wrap .home')?.innerText.trim() || "0"
					const wins_pitcher = gameDiv.querySelector('.wins .name a')?.innerText.trim()
					const loses_pitcher = gameDiv.querySelector('.loses .name a')?.innerText.trim()
					const saves_pitcher = gameDiv.querySelector('.saves .name a')?.innerText.trim()
					schedules.push({ 
						'date': date,
						'day': day,
						'time': time,
						'location': location,
						'game_no': game_no,
						'away_team_url': away_team,
						'away_score': away_score,
						'home_team_url': home_team,
						'home_score': home_score,
						'wins_pitcher': wins_pitcher,
						'loses_pitcher': loses_pitcher,
                        'saves_pitcher': saves_pitcher
					});
				})
			})
			return JSON.stringify(schedules);
		})()`, &jsonStr),
	)

	if err != nil {
		return nil, err
	}

	var schedulesYesterday []RawGameBox
	if err := json.Unmarshal([]byte(jsonStr), &schedulesYesterday); err != nil {
		return nil, err
	}

	var final []GameBox
	for _, item := range schedulesYesterday {
		final = append(final, GameBox{
			Date:           item.Date,
			Day:            item.Day,
			Location:       item.Location,
			GameNo:         item.GameNo,
			AwayTeam:       getTeamName(item.AwayTeamURL).FullName,
			AwayTeamAbbr:   getTeamName(item.AwayTeamURL).AbbrName,
			AwayTeamSingle: getTeamName(item.AwayTeamURL).OneName,
			AwayScore:      item.AwayScore,
			HomeTeam:       getTeamName(item.HomeTeamURL).FullName,
			HomeTeamAbbr:   getTeamName(item.HomeTeamURL).AbbrName,
			HomeTeamSingle: getTeamName(item.HomeTeamURL).OneName,
			HomeScore:      item.HomeScore,
			WinsPitcher:    item.WinsPitcher,
			LosesPitcher:   item.LosesPitcher,
			SavesPitcher:   item.SavesPitcher,
		})
	}

	return final, nil
}

// 定義一個名為 getYesterdayScheduleHandler 的函式，w 是用來寫入 HTTP 回應的物件，r 是用來讀取請求資料的物件
func getYesterdayScheduleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	schedulesYesterday, err := FetchYesterdaySchedule()
	if err != nil {
		http.Error(w, "Failed to fetch schedule", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(schedulesYesterday)
}

// 定義一個名為 FetchTomorrowSchedule 的函式，返回值是 []GameSchedule（一個比賽排程的 slice）以及一個錯誤（error）
func FetchTomorrowSchedule() ([]GameSchedule, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var jsonStr string
	url := "https://cpbl.com.tw"

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`.next`, chromedp.ByQuery), // 等待「後一天」按鈕出現
		chromedp.Click(`.next`, chromedp.ByQuery),       // 點擊「後一天」按鈕
		chromedp.Sleep(2*time.Second),                   // 給 JS 時間載入內容
		chromedp.Evaluate(`( () => {
			let schedules = [];
			const a = document.querySelectorAll('.major');
			a.forEach((dateDiv) => {
				const b = dateDiv.querySelectorAll('.game_item')
				const date = document.querySelector('.date').innerText.trim()
				const day = document.querySelector('.day').innerText.trim()
				b.forEach((gameDiv) => {
					const time = gameDiv.querySelector('.time')?.innerText.trim()
					const location = gameDiv.querySelector('.place').innerText.trim()
					const game_no = gameDiv.querySelector('.game_no a').innerText.trim()
					let away_team = gameDiv.querySelector('.away .team_name a').getAttribute('href').trim()
					let home_team = gameDiv.querySelector('.home .team_name a').getAttribute('href').trim()
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
						'away_team_url': away_team,
						'away_score': away_score,
						'home_team_url': home_team,
						'home_score': home_score,
						'away_sp': away_sp,
						'home_sp': home_sp
					});
				})
			})
			return JSON.stringify(schedules);
		})()`, &jsonStr),
	)

	if err != nil {
		return nil, err
	}

	var schedulesTomorrow []RawGameSchedule
	if err := json.Unmarshal([]byte(jsonStr), &schedulesTomorrow); err != nil {
		return nil, err
	}

	var final []GameSchedule
	for _, item := range schedulesTomorrow {
		final = append(final, GameSchedule{
			Date:           item.Date,
			Day:            item.Day,
			Time:           item.Time,
			Location:       item.Location,
			GameNo:         item.GameNo,
			AwayTeam:       getTeamName(item.AwayTeamURL).FullName,
			AwayTeamAbbr:   getTeamName(item.AwayTeamURL).AbbrName,
			AwayTeamSingle: getTeamName(item.AwayTeamURL).OneName,
			AwayScore:      item.AwayScore,
			AwaySP:         item.AwaySP,
			HomeTeam:       getTeamName(item.HomeTeamURL).FullName,
			HomeTeamAbbr:   getTeamName(item.HomeTeamURL).AbbrName,
			HomeTeamSingle: getTeamName(item.HomeTeamURL).OneName,
			HomeScore:      item.HomeScore,
			HomeSP:         item.HomeSP,
		})
	}

	return final, nil
}

// 定義一個名為 getTomorrowScheduleHandler 的函式，w 是用來寫入 HTTP 回應的物件，r 是用來讀取請求資料的物件
func getTomorrowScheduleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	schedulesTomorrow, err := FetchTomorrowSchedule()
	if err != nil {
		http.Error(w, "Failed to fetch schedule", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(schedulesTomorrow)
}

func main() {
	// 啟動時先抓取一次
	var err error
	schedules, err = FetchSchedule()
	if err != nil {
		log.Println("Initial fetch failed:", err)
	}

	// 啟動 API 伺服器
	http.HandleFunc("/schedule", getScheduleHandler)
	http.HandleFunc("/schedule/yesterday", getYesterdayScheduleHandler)
	http.HandleFunc("/schedule/tomorrow", getTomorrowScheduleHandler)
	fmt.Println("Server running on :14888")
	log.Fatal(http.ListenAndServe(":14888", nil))
}
