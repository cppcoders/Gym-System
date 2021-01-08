package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gorilla/mux"
)

var db *sql.DB

// Member Struct
type Member struct {
	Name              string        `json:"name"`
	ID                string        `json:"id"`
	Phone             string        `json:"phone"`
	Address           string        `json:"address"`
	Gender            string        `json:"gender"`
	MembershipType    string        `json:"membershiptype"`
	Period            string        `json:"period"`
	RegisterationDate string        `json:"registerationdate"`
	StartDate         string        `json:"startdate"`
	EndDate           string        `json:"enddate"`
	IsStudent         bool          `json:"isstudent"`
	Student           StudentExtras `json:"student"`
	Staff             StaffExtras   `json:"staff"`
}

// StudentExtras is Extra Member Fields for Student
type StudentExtras struct {
	Semester   int    `json:"semester"`
	College    string `json:"college"`
	Department string `json:"department"`
}

// StaffExtras is Extra Member Fields for Staff
type StaffExtras struct {
	Position string `json:"position"`
}

// Members is All Members we Have
var Members []Member

func getMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Members)
}
func getMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	for _, member := range Members {
		if member.ID == params["id"] {
			json.NewEncoder(w).Encode(member)
			return
		}
	}
	json.NewEncoder(w).Encode(&Member{})

}
func createMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var member Member
	_ = json.NewDecoder(r.Body).Decode(&member)
	member.RegisterationDate = time.Now().Format("2006-01-02 15:04:05")
	temp, _ := time.Parse("2006-01-02", member.StartDate)
	if member.MembershipType == "Daily" {
		member.EndDate = temp.AddDate(0, 0, 1).Format("2006-01-02")
	} else if member.MembershipType == "Monthly" {
		member.EndDate = temp.AddDate(0, 1, 0).Format("2006-01-02")
	} else {
		member.EndDate = temp.AddDate(0, 4, 0).Format("2006-01-02")
	}
	if member.IsStudent == true {
		member.Staff = StaffExtras{}
	} else {
		member.Student = StudentExtras{}
	}
	Members = append(Members, member)
	if member.IsStudent == false {
		_, err := db.Exec("INSERT INTO staff VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", member.Name, member.ID, member.Address, member.Gender, member.MembershipType, member.Period, member.StartDate, member.EndDate, member.Phone, member.Staff.Position)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Server error, unable to create.", 500)
			return
		}
	} else {
		_, err := db.Exec("INSERT INTO student VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", member.Name, member.ID, member.Address, member.Gender, member.MembershipType, member.Period, member.StartDate, member.EndDate, member.Phone, member.Student.Semester, member.Student.College, member.Student.Department)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Server error, unable to create.", 500)
			return
		}
	}
	w.Write([]byte("Member Registered Successfully!"))
}
func updateMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, member := range Members {
		if member.ID == params["id"] {
			var temp Member
			_ = json.NewDecoder(r.Body).Decode(&temp)
			member.StartDate = temp.StartDate
			member.MembershipType = temp.MembershipType
			member.Period = temp.Period
			temp2, _ := time.Parse("2006-01-02", member.StartDate)
			if member.MembershipType == "Daily" {
				member.EndDate = temp2.AddDate(0, 0, 1).Format("2006-01-02")
			} else if member.MembershipType == "Monthly" {
				member.EndDate = temp2.AddDate(0, 1, 0).Format("2006-01-02")
			} else {
				member.EndDate = temp2.AddDate(0, 4, 0).Format("2006-01-02")
			}
			Members[index] = member
			json.NewEncoder(w).Encode(member)
			return
		}
	}

}
func main() {
	db, _ = sql.Open("sqlite3", "./db.db")
	defer db.Close()

	//------------------------------------
	rows, _ := db.Query("select * from student")
	var member Member
	for rows.Next() {
		member.IsStudent = true
		member.Student = StudentExtras{}
		member.Staff = StaffExtras{}
		rows.Scan(&member.Name, &member.ID, &member.Address, &member.Gender, &member.MembershipType, &member.Period, &member.StartDate, &member.EndDate, &member.Phone, &member.Student.Semester, &member.Student.College, &member.Student.Department)
		Members = append(Members, member)
	}
	rows2, _ := db.Query("select * from staff")
	for rows2.Next() {
		member.IsStudent = false
		member.Student = StudentExtras{}
		member.Staff = StaffExtras{}
		rows2.Scan(&member.Name, &member.ID, &member.Address, &member.Gender, &member.MembershipType, &member.Period, &member.StartDate, &member.EndDate, &member.Phone, &member.Staff.Position)
		Members = append(Members, member)
	}
	//------------------------------------
	r := mux.NewRouter()
	r.HandleFunc("/members", getMembers).Methods("GET")
	r.HandleFunc("/members", createMember).Methods("POST")
	r.HandleFunc("/members/{id}", getMember).Methods("GET")
	r.HandleFunc("/members/{id}", updateMember).Methods("PUT")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "Home.html") })
	r.HandleFunc("/Home", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "Home.html") })
	r.HandleFunc("/renew", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "renew.html") })
	r.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "view.html") })
	r.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "check.html") })
	r.HandleFunc("/registeration", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "registeration.html") })
	fmt.Println("http://localhost:8009/")
	log.Fatal(http.ListenAndServe(":8009", r))

}
