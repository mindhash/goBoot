package base



import (
"time"
"log"
"gopkg.in/mgo.v2"
//"gopkg.in/mgo.v2/bson"
)
 

type Datastore interface {
	Insert (colName string, val interface{}) error 
	//FindByKey (colName string, key string, result interface{}  )   error
	//FindAll(colName string, result interface{} ) error  
	//FindByValue (colName string, query interface{}, result interface{} ) error
	//UpdateByKey(colName string, key string, val interface{}) error
	Close() 
}

type DStoreSpec struct {
	Hostaddr, Dbname, Dbuser, Dbpwd string
}

type  Mongostore struct {
	session *mgo.Session 
	dbspec DStoreSpec
}

//To Do: Capture error in log
func (ds *Mongostore) Insert (colName string, val interface{}) error {
	collection := ds.session.DB(ds.dbspec.Dbname).C("colName")
	err := collection.Insert(val)
	return err
}

func (ds *Mongostore) getName() string {
	return ds.dbspec.Dbname
}

func (ds *Mongostore) Close()  {
	ds.session.Close()
}

//returns mongo DB Session 
func GetDatastore (spec DStoreSpec) (ds Datastore, err error ) {

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{spec.Hostaddr},
		Timeout:  60 * time.Second,
		Database: spec.Dbname,
		Username: spec.Dbuser,
		Password: spec.Dbpwd,
	}
	
	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatalf("CreateSession: %s\n", err)
		return nil, err
	}

	mongoSession.SetMode(mgo.Monotonic, true)
	return &Mongostore{session: mongoSession, dbspec:spec}, nil
}
