package main

import (
	
	"sync"

	router "github.com/jamesmukumu/guzman/work/Router"

	"github.com/jamesmukumu/guzman/work/db"
)


var wg = &sync.WaitGroup{}
func main(){



wg.Add(2)

defer wg.Wait()
go func(){
defer wg.Done()
router.ServerSetup()
}()
go func(){
defer wg.Done()
db.Db_connection()
}()


}