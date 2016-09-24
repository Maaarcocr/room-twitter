// room-twitter
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type Booking struct {
	Username string `json:"username"`
	Notes    string `json:"notes"`
	Start    string `json:"start"`
	End      string `json:"end"`
}

type RespLogin struct {
	Societies [][]string
	Email     string
	Token     string
	Quota     int
	Groups    []string
}

type Room struct {
	Water            bool   `json:"water_fountain"`
	Printers         bool   `json:"printers"`
	Name             string `json:"room_name"`
	ID               string `json:"room_id"`
	Coffee           bool   `json:"coffee"`
	Capacity         int    `json:"capacity"`
	IndividualAccess bool   `json:"individual_access"`
}

type SimpleRoom struct {
	Name string
	Free bool
}

func login() string {
	resp, err := http.PostForm("http://localhost:80/api/v1/login", url.Values{"username": {"emily"}, "password": {"emilypassword"}})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyRaw, _ := ioutil.ReadAll(resp.Body)
	var body RespLogin
	json.Unmarshal(bodyRaw, &body)
	//fmt.Println(body, string(bodyRaw))
	token := body.Token
	return token
}

func getRoomsList(token string) []Room {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:80/api/v1/get_list_of_rooms", nil)
	req.Header.Add("Authorization", "Token "+token)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyRaw, _ := ioutil.ReadAll(resp.Body)
	body := make([]Room, 0)
	json.Unmarshal(bodyRaw, &body)
	return body
}

func getDayString(now time.Time) string {
	return now.Format("20060102")
}

func getRoomStatusResponse(now time.Time, token string, room Room) []Booking {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:80/api/v1/get_room_bookings", nil)
	req.Header.Add("Authorization", "Token "+token)
	q := req.URL.Query()
	q.Add("room_id", room.ID)
	q.Add("date", getDayString(now))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	bodyRaw, _ := ioutil.ReadAll(resp.Body)
	body := make([]Booking, 0)
	json.Unmarshal(bodyRaw, &body)
	return body
}

func readTime(str string) int {
	r := regexp.MustCompile("[^:]*")
	result := r.FindString(str)
	num, err := strconv.Atoi(result)
	if err != nil {
		panic(err)
	}
	return num
}

func checkStatus(now time.Time, booking Booking) bool {
	endNum := readTime(booking.End)
	startNum := readTime(booking.Start)
	nowHour := now.Hour()
	return (nowHour < endNum && nowHour >= startNum)
}

func getRoomsStatus(token string, rooms []Room, now time.Time) []SimpleRoom {
	simplerooms := make([]SimpleRoom, 0)
	for _, room := range rooms {
		//fmt.Println(room.ID)
		bookings := getRoomStatusResponse(now, token, room)
		if len(bookings) > 0 {
			free := true
			for _, booking := range bookings {
				if checkStatus(now, booking) {
					free = false
					break
				}
			}
			simplerooms = append(simplerooms, SimpleRoom{Name: room.Name, Free: free})
		} else {
			simplerooms = append(simplerooms, SimpleRoom{Name: room.Name, Free: true})
		}
	}
	return simplerooms
}

func main() {
	now := time.Now().Add(time.Hour * 14)
	token := login()
	rooms := getRoomsList(token)
	simplerooms := getRoomsStatus(token, rooms, now)
	fmt.Println(simplerooms)
}
