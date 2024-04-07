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

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Friends []string `json:"friends"`
}

var users = make(map[string]User)

var userMutex = &sync.Mutex{}
var nextID = 1

func zapis() {
	var formattedUsers []string

	for _, user := range users {
		formattedUser := fmt.Sprintf("ID: %s | Name: %s | Age: %d | Friends: %v", user.ID, user.Name, user.Age, user.Friends)
		formattedUsers = append(formattedUsers, formattedUser)
	}

	file, err := os.OpenFile("server.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Не смогли откыть файл, ", err)
		return
	}
	defer file.Close()
	for _, user := range formattedUsers {
		file.WriteString(fmt.Sprintf("Server 1: %s\n", user))
	}
	file.WriteString("===========================================================\n")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userMutex.Lock()
	newUser.ID = strconv.Itoa(nextID)
	nextID++
	userMutex.Unlock()

	users[newUser.ID] = newUser

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": newUser.ID})
	zapis()
}

//Задание 2

type MakeFriendsRequest struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

func makeFriendsHandler(w http.ResponseWriter, r *http.Request) {
	var request MakeFriendsRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sourceUser, sourceExists := users[request.SourceID]
	targetUser, targetExists := users[request.TargetID]

	if !sourceExists || !targetExists {
		http.Error(w, "One or both users do not exist", http.StatusNotFound)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, targetUser.Name)
	targetUser.Friends = append(targetUser.Friends, sourceUser.Name)

	users[request.SourceID] = sourceUser
	users[request.TargetID] = targetUser

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Статус %d %s и %s теперь друзья", http.StatusOK, sourceUser.Name, targetUser.Name)
	zapis()
}

// Задание 3
type DeleteUserRequest struct {
	TargetID string `json:"target_id"`
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var request DeleteUserRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users[request.TargetID]
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	for range user.Friends {
		for id, friend := range users {
			for i, name := range friend.Friends {
				if name == user.Name {
					friend.Friends = append(friend.Friends[:i], friend.Friends[i+1:]...)
					users[id] = friend
					break
				}
			}
		}
	}

	delete(users, request.TargetID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Статус %d Удалён пользователь: %s", http.StatusOK, user.Name)
	zapis()
}

// Задание 4

func getUserFriendsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/friends/"):]

	user, exists := users[userID]
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user.Friends)
	zapis()
}

// Задание 5

type UpdateUserAgeRequest struct {
	NewAge int `json:"new_age"`
}

func updateUserAgeHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Path[len("/usera/"):]

	var request UpdateUserAgeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, exists := users[userID]
	if !exists {
		http.Error(w, "User does not exist", http.StatusNotFound)
		return
	}

	user.Age = request.NewAge
	users[userID] = user

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "возраст пользователя успешно обновлён")
	zapis()
}

//вывод данных

func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	var formattedUsers []string

	for _, user := range users {
		formattedUser := fmt.Sprintf("ID: %s | Name: %s | Age: %d | Friends: %v", user.ID, user.Name, user.Age, user.Friends)
		formattedUsers = append(formattedUsers, formattedUser)
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, user := range formattedUsers {
		fmt.Fprintln(w, user)
	}
}

func main() {
	http.HandleFunc("/create", createUserHandler)        // Задание 1
	http.HandleFunc("/make_friends", makeFriendsHandler) // Задание 2
	http.HandleFunc("/user", deleteUserHandler)          // Задание 3
	http.HandleFunc("/friends/", getUserFriendsHandler)  // Задание 4
	http.HandleFunc("/usera/", updateUserAgeHandler)     // Задание 5
	http.HandleFunc("/", getAllUsersHandler)             // Вывод данных

	fmt.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
