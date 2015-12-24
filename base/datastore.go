package base



import (
"time"
"log"
"gopkg.in/mgo.v2"
_ "gopkg.in/mgo.v2/bson"
)
 

type Datastore interface {
	Insert (collection string, val interface{}) error 
	//FindByKey (collection string, key string, result interface{}  )   error
	//FindAll(collection string, result interface{} ) error  
	FindByValue (collection string, query interface{}, result interface{} ) error
	//UpdateByKey(collection string, key string, val interface{}) error
	Close() 
	GetName() string
}

type DStoreSpec struct {
	Hostaddr, Dbname, Dbuser, Dbpwd string
}

type  Mongostore struct {
	session *mgo.Session 
	dbspec DStoreSpec
}

func (ds *Mongostore) FindByValue(collection string, query interface{}, result interface{} ) error {
	col := ds.session.DB(ds.dbspec.Dbname).C(collection)
	err := col.Find(query).One(&result)
	if err != nil {
                log.Fatal(err)
    }
    return err
}

//To Do: Capture error in log
func (ds *Mongostore) Insert(collection string, val interface{}) error {
	col := ds.session.DB(ds.dbspec.Dbname).C(collection)
	err := col.Insert(val)
	return err
}

func (ds *Mongostore) GetName() string {
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

	//c := mongoSession.DB("test").C("mytable")
    //    err = c.Insert(bson.M{"name":"Asta1"})

    //    if err != nil {
    //            log.Fatal(err)
    //    }
 

	return &Mongostore{session: mongoSession, dbspec:spec}, nil
}
