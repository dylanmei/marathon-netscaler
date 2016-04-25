package main

type App struct {
	ID           string
	ServiceGroup string
	Port         int

	Agents []Agent
}

type Agent struct {
	ID   string
	Host string
}
