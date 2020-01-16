package sharded

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/percona/percona-backup-mongodb/e2e-tests/pkg/pbm"
)

type shard struct {
	name string
	cn   *pbm.Mongo
}

func (c *Cluster) DistributedTransactions() {
	log.Println("peek a random replsets")
	var rs [2]shard
	i := 0
	for name, cn := range c.shards {
		rs[i] = shard{name: name, cn: cn}
		if i == 1 {
			break
		}
		i++
	}
	if rs[0].name == "" || rs[1].name == "" {
		log.Fatalln("no shards in cluster")
	}

	ctx := context.Background()
	_, err := rs[0].cn.Conn().Database("test").Collection("trx0").InsertOne(ctx, bson.M{"x": 0})
	if err != nil {
		log.Fatalln("ERROR: prepare collections for trx0:", err)
	}
	_, err = rs[1].cn.Conn().Database("test").Collection("trx1").InsertOne(ctx, bson.M{"y": 0})
	if err != nil {
		log.Fatalln("ERROR: prepare collections for trx1:", err)
	}

	conn := c.mongos.Conn()
	sess, err := conn.StartSession(
		options.Session().
			SetDefaultReadPreference(readpref.Primary()).
			SetCausalConsistency(true).
			SetDefaultReadConcern(readconcern.Majority()).
			SetDefaultWriteConcern(writeconcern.New(writeconcern.WMajority())),
	)
	if err != nil {
		log.Fatalln("ERROR: start session:", err)
	}
	defer sess.EndSession(ctx)

	err = mongo.WithSession(ctx, sess, func(sc mongo.SessionContext) error {
		var err error
		defer func() {
			if err != nil {
				sess.AbortTransaction(sc)
				log.Fatalln("ERROR: transaction:", err)
			}
		}()

		_, err = conn.Database("test").Collection("trx0").UpdateOne(sc, bson.D{}, bson.D{{"$set", bson.M{"x": 1}}})
		if err != nil {
			log.Fatalln("ERROR: update in transaction trx0:", err)
		}

		// !!! wait for bcp

		_, err = conn.Database("test").Collection("trx1").UpdateOne(sc, bson.D{}, bson.D{{"$set", bson.M{"y": 1}}})
		if err != nil {
			log.Fatalln("ERROR: update in transaction trx0:", err)
		}

		return sess.CommitTransaction(sc)
	})

	// conn.Database("test").Collection("trx0").Drop(ctx)
	// conn.Database("test").Collection("trx1").Drop(ctx)
}
