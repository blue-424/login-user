package main

import (
	"7userWallet/model"
	"7userWallet/repository"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	jsonFile, err := os.Open("cnf/sql.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	data, err := ioutil.ReadFile("cnf/sql.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var obj model.Configuration
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println(err)
		return
	}
	db := repository.OpenConnection(obj)
	defer db.Close()
	http.HandleFunc("/user-wallet", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "خطا در خواندن درخواست", http.StatusInternalServerError)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "خطا در خواندن درخواست", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		//login//
		var data map[string]string
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "خطا در پارس کردن JSON", http.StatusInternalServerError)
			return
		}
		username, ok := data["username"]
		if !ok {
			http.Error(w, "مقدار username یافت نشد", http.StatusBadRequest)
			return
		}
		password, ok := data["password"]
		if !ok {
			http.Error(w, "مقدار password یافت نشد", http.StatusBadRequest)
			return
		}
		valid, name := checkCredentials(username, password, obj)
		if valid {
			fmt.Printf("Hello %s\n", name)
		} else {
			http.Error(w, "نام کاربری یا رمز عبور اشتباه است", http.StatusUnauthorized)
			return
		}
	})
	log.Println("Starting server...")
	l, err := net.Listen("tcp", "localhost:8083")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(l, nil))
	log.Println("Sending request...")
	res, err := http.Get("http://localhost:8083/user-wallet")
	if err := http.ListenAndServe(":8083", nil); err != nil {
		log.Fatal(err)
	}
	log.Println("Reading response...")
	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get("https://jsonplaceholder.typicode.com/todos/1")
	if err != nil {
		log.Printf("Request Failed: %s", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return
	}
	bodyString := string(body)
	log.Print(bodyString)

}
func checkCredentials(username, password string, obj model.Configuration) (bool, string) {
	db := repository.OpenConnection(obj)
	defer db.Close()
	query := "SELECT name FROM tbl_users WHERE username = $1 AND password = $2"
	row := db.QueryRow(query, username, password)
	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		return false, ""
	} else if err != nil {
		log.Println(err)
		return false, ""
	}
	return true, name
}
