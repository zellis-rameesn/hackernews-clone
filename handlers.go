package main

import (
	"database/sql"
	"errors"
	"net/http"
)

const loggedInUserEmail = "logged_in_user_email"

type AppOperations interface {
	Home(w http.ResponseWriter, r *http.Request)
	About(w http.ResponseWriter, r *http.Request)
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	filter := Filter{
		Query:    r.URL.Query().Get("q"),
		OrderBy:  r.URL.Query().Get("order_by"),
		Page:     app.getFilterFromUrl(r, "page", 1),
		PageSize: app.getFilterFromUrl(r, "page_size", 3),
	}
	posts, meta, err := app.postRepo.GetAll(filter)
	if err != nil {
		app.errorLog.Printf("Failed to get posts: %s", err.Error())
		app.session.Put(r, "flash", "Failed to fetch posts")
		app.render(w, r, "index.html", nil)
		return
	}
	app.render(w, r, "index.html", &TemplateData{Posts: posts, Metadata: meta})
}

func (app *Application) Vote(w http.ResponseWriter, r *http.Request) {
	postId := app.getFilterFromUrl(r, "post_id", 0)
	user := app.getUserFromContext(r)

	_, err := app.postRepo.AddVote(user.Id, postId)
	if err != nil {
		if err.Error() == ErrDuplicateVote.Error() {
			app.session.Put(r, "flash", "Duplicate vote!")
		} else {
			app.session.Put(r, "flash", "Failed to vote!")
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.session.Put(r, "flash", "Added vote!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) Comments(w http.ResponseWriter, r *http.Request) {
	postId := app.getFilterFromUrl(r, "post_id", 0)
	user := app.getUserFromContext(r)

	post, err := app.postRepo.GetByID(postId)

	if err != nil {
		app.serverError(w, errors.New("Failed to fetch post"))
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		form := NewForm(r.PostForm)
		// form.MaxLength("comment", 60)
		// isValid := form.isValid()
		// if !isValid {
		// 	form.Errors.Add("generic", "Invalid comment!")
		// 	return
		// }
		comment := form.Get("comment")
		_, err := app.postRepo.AddComment(comment, user.Id, postId)

		if err != nil {
			app.session.Put(r, "flash", "Failed to add comment!")
			return
		}

		app.session.Put(r, "flash", "Added comment!")
	}

	comments, err := app.postRepo.GetComments(postId)

	if err != nil {
		app.serverError(w, errors.New("Failed to fetch comments"))
		return
	}

	app.render(w, r, "comments.html", &TemplateData{Post: post, Comments: comments})
}
func (app *Application) About(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("loggedInUserEmail:", app.session.GetString(r, loggedInUserEmail))
	app.infoLog.Println("Session data:", app.session.GetString(r, "flash"))
	app.render(w, r, "about.html", nil)
}

func (app *Application) Submit(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		form := NewForm(r.PostForm)
		form.Required("url", "title").validateUrl("url").MaxLength("title", 40)
		isValid := form.isValid()
		if !isValid {
			form.Errors.Add("generic", "Invalid form!")
			app.render(w, r, "submit.html", &TemplateData{Form: form})
			return
		}
		url := form.Get("url")
		title := form.Get("title")
		user := app.getUserFromContext(r)
		app.infoLog.Println("User fetched from context!!", user.Id)

		postID, err := app.postRepo.CreatePost(url, title, user.Id)
		app.infoLog.Println("postID", postID)
		if err != nil {
			app.errorLog.Printf("post create error %s", err.Error())
			// do not sent err to frontend since it may contain db related stuffs
			if err.Error() == ErrDuplicatePostTitle.Error() {
				form.Errors.Add("generic", "Duplicate post title!")
			} else {
				form.Errors.Add("generic", "Failed to create post!")
			}

			app.render(w, r, "submit.html", &TemplateData{Form: form})
			return
		}
		app.session.Put(r, "flash", "Post created successfully!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.render(w, r, "submit.html", &TemplateData{Form: NewForm(r.PostForm)})
}

func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, loggedInUserEmail)
	app.session.Put(r, "flash", "Logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		form := NewForm(r.PostForm)
		form.Required("email", "password").MaxLength("password", 30).MinLength("email", 10).MinLength("password", 6).validateMail("email")
		isValid := form.isValid()
		if !isValid {
			app.errorLog.Printf("Validation errors %v", form.Errors)
			form.Errors.Add("generic", "Invalid form!")
			app.render(w, r, "login.html", &TemplateData{Form: form})
			return
		}
		email := form.Get("email")
		password := form.Get("password")
		userEmail, err := app.userRepo.Authenticate(email, password)

		if err != nil {
			if err == sql.ErrNoRows {
				form.Errors.Add("generic", "No user found!")
			} else {
				form.Errors.Add("generic", err.Error())
			}
			app.render(w, r, "login.html", &TemplateData{Form: form})
			return
		}
		app.session.Put(r, loggedInUserEmail, userEmail)
		app.session.Put(r, "flash", "User logged in!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	app.render(w, r, "login.html", &TemplateData{Form: NewForm(r.PostForm)})
}

func (app *Application) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		form := NewForm(r.PostForm)
		form.Required("email", "password", "name").MaxLength("password", 30).MaxLength("name", 30).MinLength("email", 10).MinLength("password", 6).validateMail("email")
		isValid := form.isValid()
		if !isValid {
			app.errorLog.Printf("Validation errors %v", form.Errors)
			form.Errors.Add("generic", "Invalid form!")
			app.render(w, r, "register.html", &TemplateData{Form: form})
			return
		}
		email := form.Get("email")
		password := form.Get("password")
		name := form.Get("name")
		avatar := form.Get("avatar")
		userEmail, err := app.userRepo.InsertUser(name, email, password, avatar)
		if err != nil {
			form.Errors.Add("generic", err.Error())
			app.render(w, r, "login.html", &TemplateData{Form: form})
			return

		}
		app.session.Put(r, "flash", "User registered!")
		app.session.Put(r, loggedInUserEmail, userEmail)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	app.render(w, r, "register.html", &TemplateData{Form: NewForm(r.PostForm)})
}

func (app *Application) Contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "contact.html", nil)
}
