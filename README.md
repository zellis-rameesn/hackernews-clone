# Hacker News Clone

A lightweight web application that replicates the core functionality of Hacker News, built entirely with Go. It features server-side rendered HTML templates, a custom routing setup, and database integration for managing users and posts.

## 🛠️ Tech Stack

* **Backend:** [Go](https://go.dev/) (Golang)
* **Frontend:** HTML5, CSS3 (Go `html/template`)
* **Database:** SQLite (via `app.db`)

## ✨ Features

This project includes:
* **User Management:** User registration and authentication (`user_repository.go`, `middlewares.go`).
* **Post Management:** Ability to submit and retrieve posts (`post_repository.go`).
* **Server-Side Rendering:** Dynamic HTML rendering using Go's native template engine (`render.go`, `renderer.go`, `templates/`).
* **Form Validation:** Handling user input safely (`form.go`).

## 🚀 Getting Started

### Prerequisites
Make sure you have [Go](https://go.dev/doc/install) installed on your machine.

### Installation

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/zellis-rameesn/hackernews-clone.git](https://github.com/zellis-rameesn/hackernews-clone.git)
    cd hackernews-clone
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Run the application:**
    ```bash
    go run .
    ```
    *(Alternatively, build the binary with `go build` and run the executable).*

4.  **Access the app:** Open your browser and navigate to the local server address (usually `http://localhost:8080`, depending on your `main.go` configuration).

## 📂 Project Structure

* `main.go`: Application entry point.
* `routes.go` / `handlers.go`: Defines the HTTP routes and their respective handler functions.
* `models.go`: Contains the data structures used across the application.
* `*_repository.go`: Handles database operations for different entities (Users, Posts).
* `middlewares.go`: HTTP middleware for tasks like authentication or logging.
* `templates/`: Directory containing the HTML templates.
* `public/css/`: Static stylesheets.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/zellis-rameesn/hackernews-clone/issues).

## 👤 Author

**Ramees Noushad**
* GitHub: [@zellis-rameesn](https://github.com/zellis-rameesn)
