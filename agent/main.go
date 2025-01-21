package main

func main() {
	if err := CreateDatabaseFileIfNotExist(); err != nil {
		panic(err)
	}
	if err := InitiateDatabaseInstances(); err != nil {
		panic(err)
	}
	if err := MigrateDatabase(); err != nil {
		panic(err)
	}
	go startHttpServer()
	go startDnsServer()
	<-make(chan struct{})
}
