package handler

import "net/http"

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	appSession, _ := store.Get(r, "app-session")
	appSession.Options.MaxAge = -1
	appSession.Save(r, w)

	eqSession, _ := store.Get(r, "equations-session")
	eqSession.Options.MaxAge = -1
	eqSession.Save(r, w)

	http.SetCookie(w, &http.Cookie{
		Name:     "app-session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "equations-session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
