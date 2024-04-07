package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Задание 1

type User2 struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Friends []string `json:"friends"`
}

var users2 = make(map[string]User2)

var userMutex2 = &sync.Mutex{}
var nextId int = 1

func zapis2() {
	var formattedUsers []string

	for _, user := range users2 {
		formattedUser := fmt.Sprintf("ID: %s | Name: %s | Age: %d | Friends: %v", user.ID, user.Name, user.Age, user.Friends)
		formattedUsers = append(formattedUsers, formattedUser)
	}

	file, err := os.OpenFile("server.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Не смогли открыть файл, ", err)
		return
	}
	defer file.Close()
	for _, user := range formattedUsers {
		file.WriteString(fmt.Sprintf("Server 2: %s\n", user))
	}
	file.WriteString("===========================================================\n")
}

func createUserHandler2(w http.ResponseWriter, r *http.Request) {
	var newUser User2
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userMutex2.Lock()
	newUser.ID = strconv.Itoa(nextId)
	nextId++
	userMutex2.Unlock()

	users2[newUser.ID] = newUser

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": newUser.ID})
	zapis2()
}

//Задание 2

type MakeFriendsRequest2 struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

func makeFriendsHandler2(w http.ResponseWriter, r *http.Request) {
	var request MakeFriendsRequest2

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceUser, sourceExists := users2[request.SourceID]
	targetUser, targetExists := users2[request.TargetID]

	if !sourceExists || !targetExists {
		http.Error(w, "One or both users do not exist", http.StatusNotFound)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, targetUser.Name)
	targetUser.Friends = append(targetUser.Friends, sourceUser.Name)

	users2[request.SourceID] = sourceUser
	users2[request.TargetID] = targetUser

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Статус %d %s и %s теперь друзья", http.StatusOK, sourceUser.Name, targetUser.Name)
	zapis2()
}

// Задание 3
type DeleteUserRequest2 struct {
	TargetID string `json:"target_id"`
}

func deleteUserHandler2(w http.ResponseWriter, r *http.Request) {
	var request DeleteUserRequest2

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users2[request.TargetID]
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	for range user.Friends {
		for id, friend := range users2 {
			for i, name := range friend.Friends {
				if name == user.Name {
					friend.Friends = append(friend.Friends[:i], friend.Friends[i+1:]...)
					users2[id] = friend
					break
				}
			}
		}
	}

	delete(users2, request.TargetID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Статус %d Удалён пользователь: %s", http.StatusOK, user.Name)
	zapis2()
}

// Задание 4

func getUserFriendsHandler2(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/friends/"):]

	user, exists := users2[userID]
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.Friends)
	zapis2()
}

// Задание 5

type UpdateUserAgeRequest2 struct {
	NewAge int `json:"new_age"`
}

func updateUserAgeHandler2(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/usera/"):]

	var request UpdateUserAgeRequest2
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users2[userID]
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	user.Age = request.NewAge
	users2[userID] = user

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "возраст пользователя успешно обновлён")
	zapis2()
}

//вывод данных

func getAllUsersHandler2(w http.ResponseWriter, r *http.Request) {
	var formattedUsers []string

	for _, user := range users2 {
		formattedUser := fmt.Sprintf("ID: %s | Name: %s | Age: %d | Friends: %v", user.ID, user.Name, user.Age, user.Friends)
		formattedUsers = append(formattedUsers, formattedUser)
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, user := range formattedUsers {
		fmt.Fprintln(w, user)
	}
}

func main() {
	http.HandleFunc("/create", createUserHandler2)        // Задание 1
	http.HandleFunc("/make_friends", makeFriendsHandler2) // Задание 2
	http.HandleFunc("/user", deleteUserHandler2)          // Задание 3
	http.HandleFunc("/friends/", getUserFriendsHandler2)  // Задание 4
	http.HandleFunc("/usera/", updateUserAgeHandler2)     // Задание 5
	http.HandleFunc("/", getAllUsersHandler2)             // Вывод данных

	fmt.Println("Starting server on :8082...")
	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}
