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
	rootCmd.Execute()
}
