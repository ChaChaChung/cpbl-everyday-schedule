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

var (
	schedules []GameSchedule
	mu        sync.Mutex // 互斥鎖，確保數據更新時不會產生競爭
)

// 定義一個名為 FetchSchedule 的函式，返回值是 []GameSchedule（一個比賽排程的 slice）以及一個錯誤（error）
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
	var schedules []GameSchedule
	err = json.Unmarshal([]byte(jsonStr), &schedules)

	// 解析過程有錯誤也要返回
	if err != nil {
		return nil, err
	}

	// 最後回傳整理好的比賽資料與 nil（代表沒有錯誤）
	return schedules, nil
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

// func main() {
// 	schedules, err := FetchSchedule()
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	// 格式化 JSON 輸出
// 	jsonData, err := json.MarshalIndent(schedules, "", "    ")
// 	if err != nil {
// 		fmt.Println("Error formatting JSON:", err)
// 		return
// 	}

// 	fmt.Println(string(jsonData))
// }

func main() {
	// 啟動時先抓取一次
	var err error
	schedules, err = FetchSchedule()
	if err != nil {
		log.Println("Initial fetch failed:", err)
	}

	// 啟動 API 伺服器
	http.HandleFunc("/schedule", getScheduleHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	http.HandleFunc("/schedule/yesterday", getScheduleHandler)
	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
