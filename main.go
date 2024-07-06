package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	Routers()
}


func Routers() {
	InitDB()
	defer db.Close()
	router := mux.NewRouter()
	router.HandleFunc("/users",
		GetUsers).Methods("GET")
	router.HandleFunc("/users",
		CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}",
		GetUser).Methods("GET")
	router.HandleFunc("/users/{id}",
		UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}",
		DeleteUser).Methods("DELETE")
	http.ListenAndServe(":3000",
		&CORSRouterDecorator{router})
}


// Task 3: Write your code here
func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

    var users []User
    result, err := db.Query("SELECT `id`, `first_name`, `middle_name`, `last_name`, `email`" +
        ", `gender`, `civil_status`, `birthday`, `contact`, `address`,floor(datediff (now(), birthday)/365) as age FROM users")

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer result.Close()

    for result.Next() {
        var user User
        err := result.Scan(&user.ID, &user.FirstName, &user.MiddleName,
            &user.LastName, &user.Email, &user.Gender, &user.CivilStatus, &user.Birthday, &user.Contact, &user.Address, &user.Age)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        users = append(users, user)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(users)
}

// Task 4: write your code for create user here 
func CreateUser(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("INSERT INTO users (first_name, middle_name" +
	", last_name, email, gender, civil_status, birthday, contact, address) VALUES (?,?,?,?,?,?,?,?,?)" )

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	var keyVal map[string]string
	err = json.Unmarshal(body, &keyVal)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	firstName := keyVal["firstName"]
    middleName := keyVal["middleName"]
    lastName := keyVal["lastName"]
    email := keyVal["email"]
    gender := keyVal["gender"]
    civilStatus := keyVal["civilStatus"]
    birthday := keyVal["birthday"]
    contact := keyVal["contact"]
    address := keyVal["address"]

	_, err = stmt.Exec(firstName, middleName, lastName, email, gender, civilStatus, birthday, contact, address)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "New user was created")

}

// Task 5: Write code for get user here
func GetUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    result, err := db.Query("SELECT id, first_name, middle_name, last_name, email"+
        ", gender, civil_status, birthday, contact, address, floor(datediff (now(), birthday)/365) as age FROM users WHERE id = ?", params["id"])
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer result.Close()

    var user User
    userFound := false
    for result.Next() {
        err := result.Scan(&user.ID, &user.FirstName, &user.MiddleName,
            &user.LastName, &user.Email, &user.Gender, &user.CivilStatus, &user.Birthday, &user.Contact, &user.Address, &user.Age)

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        userFound = true
    }

    if err := result.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if !userFound {
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "User not found with ID: %s", params["id"])
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(user)
}

// Task 6: write code for update user here
func UpdateUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    stmt, err := db.Prepare("UPDATE users SET first_name = ?," +
        "middle_name=?,last_name= ?, email=?, gender=?,civil_status=?,birthday=?,contact=?,address=? WHERE id = ?")
    if err != nil {
        panic(err.Error())
    }
    defer stmt.Close()

    var userUpdate User 
    if err := json.NewDecoder(r.Body).Decode(&userUpdate); err != nil {
        panic(err.Error())
    }

    result, err := stmt.Exec(
        userUpdate.FirstName,
        userUpdate.MiddleName,
        userUpdate.LastName,
        userUpdate.Email,
        userUpdate.Gender,
        userUpdate.CivilStatus,
        userUpdate.Birthday,
        userUpdate.Contact,
        userUpdate.Address,
        params["id"],
    )
    if err != nil {
        panic(err.Error())
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        panic(err.Error())
    }

    if rowsAffected == 0 {
        w.WriteHeader(http.StatusNotFound)
        fmt.Fprintf(w, "No user found with ID = %s", params["id"])
        return
    }

    fmt.Fprintf(w, "User with ID = %s was updated", params["id"])
}

// Task 7: Write code for delete user here

func DeleteUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    stmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer stmt.Close()

    result, err := stmt.Exec(params["id"])
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if rowsAffected == 0 {
        http.Error(w, fmt.Sprintf("No user found with ID = %s", params["id"]), http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "User with ID = %s was deleted", params["id"])
}


// Task 2: Write your code here

type User struct {
    ID          string `json:"id"`
    FirstName   string `json:"firstName"`
    MiddleName  string `json:"middleName"`
    LastName    string `json:"lastName"`
    Email       string `json:"email"`
    Gender      string `json:"gender"`
    CivilStatus string `json:"civilStatus"`
    Birthday    string `json:"birthday"`
    Contact     string `json:"contact"`
    Address     string `json:"address"`
    Age         string `json:"age"`
}

var db *sql.DB
var err error

// Task 1: Write code for DB initialization here

func InitDB()  {
	db, err = sql.Open("mysql", "user:user@tcp(127.0.0.1:3306)/userdb")
	if err != nil {
        panic(err.Error())
    }
}

type CORSRouterDecorator struct {
	R *mux.Router
}

func (c *CORSRouterDecorator) ServeHTTP(rw http.ResponseWriter,
	req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods",
			"POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Accept-Language,"+
				" Content-Type, YourOwnHeader")
	}

	if req.Method == "OPTIONS" {
		return
	}

	c.R.ServeHTTP(rw, req)
}
