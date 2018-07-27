package models

import (
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	mgofun "gopkg.in/lujiacn/mgofun.v3"
)

const (
	Active   = iota
	InActive = iota
	NoExist  = iota

	//default avatar if no thumb for user
	DefaultAvatar string = `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`
)

type User struct {
	mgofun.BaseModel `bson:",inline"`
	Name             string `bson:"Name,omitempty"`
	First            string `bson:"First,omitempty"`
	Last             string `bson:"Last,omitempty"`
	Mail             string `bson:"Mail,omitempty"`
	Depart           string `bson:"Depart,omitempty"`
	Avatar           string `bson:"Avatar,omitempty"`
	Identity         string `bson:"Identity,omitempty"` //if ldap is SAMAccount
}

func (c *User) GetAvatar() string {
	if c.Avatar != "" {
		return c.Avatar

	}
	return DefaultAvatar
}

//Save authorized saUser to local User
func (c *User) SaveUser(s *mgo.Session) error {
	//check if user exist
	u := new(User)
	q := bson.M{"Identity": c.Identity}
	do := NewDo(s, u)
	do.Query = q
	err := do.GetByQ()
	if err != nil && err != mgo.ErrNotFound {
		return err
	}

	//if user not exist create new ObjectId
	if !u.Id.Valid() {
		c.Id = bson.NewObjectId()
	} else {
		//if user exist, just update Id
		c.Id = u.Id
	}

	//Save User
	do = NewDo(s, c)
	err = do.Save()
	if err != nil {
		return err
	}
	return nil
}
