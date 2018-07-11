package main

type configuration struct {
	Mailserver  string   `json:"mailserver"`
	Mailport    int      `json:"mailport"`
	Mailfrom    string   `json:"mailfrom"`
	Mailonerror bool     `json:"mailonerror"`
	Mailtoerror string   `json:"mailtoerror"`
	Mailto      string   `json:"mailto"`
	Mailsubject string   `json:"mailsubject"`
	Keywords    []string `json:"keywords"`
}
