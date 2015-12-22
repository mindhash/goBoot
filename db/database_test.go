package db

import ("testing"
		"log"
		"runtime"
		"strings"
		"fmt"
	//	"encoding/json"
	//	"github.com/mindhash/go.assert"
		"github.com/mindhash/goBoot/base"
		"gopkg.in/mgo.v2/bson"
)
 
const (
	_dbserver string = "127.0.0.1"
	_dbname string = "test"
)

type  user struct {
	Name string
	Email string
	Active bool
}

func setupTestDB(t *testing.T) *DatabaseContext {
	dstore,err := ConnectToDataStore(base.DStoreSpec{
		Hostaddr: _dbserver, Dbname:_dbname,
 		})
	 
	if err != nil {
			log.Fatalf("Couldn't connect to DB: %v", err)
	}
		
	db, err := NewDatabaseContext(dstore.GetName(), dstore)
	assertNoError(t, err, "Couldn't create context for database 'db'")
	return db
}

func tearDownTestDB(t *testing.T, db *DatabaseContext) {
	db.Close()
}

func TestDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer tearDownTestDB(t, db)

	
	log.Printf("Create test data...")
	 body := bson.M{"name": "unknown", "data": "1234"}
	
	 err := db.Datastore.Insert("mytable",&body)
	assertNoError(t, err, "Couldn't insert data ")
	log.Printf("Inserted data successfully..")

	//body1 := Body{}
	
	//log.Printf("Retrieve data...")
	//doc,err := db.GetDoc(docid1)
    
    //err= FindAll ("FirstTable", key string, &body1  )

	//assertNoError(t, err, "Couldn't get Doc Body")
	//log.Printf("Retrieved data successfully..", &body1)
	 
	//assert.DeepEquals(t, body1, body) 

}

func TestObjectDML(t *testing.T) {
	db := setupTestDB(t)
	defer tearDownTestDB(t, db)

	
	log.Printf("Create User test data...")
	 u:= &user {"Moak","None",true}
	err := db.Datastore.Insert("users",u)


	assertNoError(t, err, "Couldn't insert data ")
	log.Printf("Inserted User successfully..")

	//body1 := Body{}
	
	//log.Printf("Retrieve data...")
	//doc,err := db.GetDoc(docid1)
    
    //err= FindAll ("FirstTable", key string, &body1  )

	//assertNoError(t, err, "Couldn't get Doc Body")
	//log.Printf("Retrieved data successfully..", &body1)
	 
	//assert.DeepEquals(t, body1, body) 

}



//////// HELPERS:

func assertFailed(t *testing.T, message string) {
	_, file, line, ok := runtime.Caller(2) // assertFailed + assertNoError + public function.
	if ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = 1
	}
	t.Fatalf("%s:%d: %s", file, line, message)
}

func assertNoError(t *testing.T, err error, message string) {
	if err != nil {
		assertFailed(t, fmt.Sprintf("%s: %v", message, err))
	}
}

func assertTrue(t *testing.T, success bool, message string) {
	if !success {
		assertFailed(t, message)
	}
}

func assertFalse(t *testing.T, failure bool, message string) {
	if failure {
		assertFailed(t, message)
	}
}