package evaluator

import (
	"base/object"
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "modernc.org/sqlite"
)

var dbConnections = make(map[string]interface{})

func RegisterDBBuiltins() {
	builtins["db.connect"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			aliasObj, ok1 := args[0].(*object.String)
			driverObj, ok2 := args[1].(*object.String)
			dsnObj, ok3 := args[2].(*object.String)

			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `db.connect` must be (STRING, STRING, STRING)")
			}

			alias := aliasObj.Value
			driver := driverObj.Value
			dsn := dsnObj.Value

			if driver == "mongodb" {
				client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dsn))
				if err != nil {
					return newError("mongodb connection error: %s", err.Error())
				}
				dbConnections[alias] = client
			} else {
				db, err := sql.Open(driver, dsn)
				if err != nil {
					return newError("sql connection error: %s", err.Error())
				}
				dbConnections[alias] = db
			}
			return TRUE
		},
	}

	builtins["db.exec"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2+", len(args))
			}
			aliasObj, ok1 := args[0].(*object.String)
			queryObj, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("first two arguments to `db.exec` must be STRING")
			}
			alias := aliasObj.Value
			query := queryObj.Value
			conn, exists := dbConnections[alias]
			if !exists {
				return newError("no connection found: %s", alias)
			}
			if c, ok := conn.(*sql.DB); ok {
				goArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					goArgs[i] = baseObjectToGoType(arg)
				}
				res, err := c.Exec(query, goArgs...)
				if err != nil {
					return newError("sql exec error: %s", err.Error())
				}
				affected, _ := res.RowsAffected()
				return &object.Integer{Value: affected}
			}
			return newError("db.exec is only for SQL databases")
		},
	}

	builtins["db.insert"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			aliasObj, ok1 := args[0].(*object.String)
			targetObj, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("first two arguments to `db.insert` must be STRING")
			}
			alias := aliasObj.Value
			target := targetObj.Value
			data := args[2]

			conn, exists := dbConnections[alias]
			if !exists {
				return newError("no connection for alias: %s", alias)
			}

			switch c := conn.(type) {
			case *sql.DB:
				hash, ok := data.(*object.Hash)
				if !ok {
					return newError("data for SQL insert must be HASH")
				}
				keys := []string{}
				placeholders := []string{}
				vals := []interface{}{}
				for k, v := range hash.Pairs {
					keys = append(keys, k)
					placeholders = append(placeholders, "?")
					vals = append(vals, baseObjectToGoType(v))
				}
				query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", target, stringsJoin(keys, ","), stringsJoin(placeholders, ","))
				_, err := c.Exec(query, vals...)
				if err != nil {
					return newError("sql insert error: %s", err.Error())
				}
			case *mongo.Client:
				coll := c.Database("test").Collection(target)
				_, err := coll.InsertOne(context.TODO(), baseObjectToGoType(data))
				if err != nil {
					return newError("mongodb insert error: %s", err.Error())
				}
			}
			return TRUE
		},
	}

	builtins["db.query"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2+", len(args))
			}
			aliasObj, ok1 := args[0].(*object.String)
			if !ok1 {
				return newError("first argument to `db.query` must be STRING")
			}
			alias := aliasObj.Value
			conn, exists := dbConnections[alias]
			if !exists {
				return newError("no connection for alias: %s", alias)
			}

			switch c := conn.(type) {
			case *sql.DB:
				queryObj, ok := args[1].(*object.String)
				if !ok {
					return newError("second argument to SQL `db.query` must be STRING")
				}
				query := queryObj.Value
				goArgs := make([]interface{}, len(args)-2)
				for i, arg := range args[2:] {
					goArgs[i] = baseObjectToGoType(arg)
				}
				rows, err := c.Query(query, goArgs...)
				if err != nil {
					return newError("sql query error: %s", err.Error())
				}
				defer rows.Close()
				cols, _ := rows.Columns()
				results := []object.Object{}
				for rows.Next() {
					vals := make([]interface{}, len(cols))
					ptrs := make([]interface{}, len(cols))
					for i := range vals {
						ptrs[i] = &vals[i]
					}
					rows.Scan(ptrs...)
					row := make(map[string]object.Object)
					for i, col := range cols {
						row[col] = goTypeToBaseObject(vals[i])
					}
					results = append(results, &object.Hash{Pairs: row})
				}
				return &object.Array{Elements: results}
			case *mongo.Client:
				coll := c.Database("test").Collection(args[1].(*object.String).Value)
				filter := make(map[string]interface{})
				if len(args) > 2 {
					filter = baseObjectToGoType(args[2]).(map[string]interface{})
				}
				cursor, err := coll.Find(context.TODO(), filter)
				if err != nil {
					return newError("mongodb find error: %s", err.Error())
				}
				var results []interface{}
				cursor.All(context.TODO(), &results)
				return goTypeToBaseObject(results)
			}
			return NULL
		},
	}

	builtins["db.update"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 4 {
				return newError("wrong number of arguments. got=%d, want=4", len(args))
			}
			aliasObj, ok1 := args[0].(*object.String)
			targetObj, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("first two arguments to `db.update` must be STRING")
			}
			alias := aliasObj.Value
			target := targetObj.Value
			match := baseObjectToGoType(args[2])
			update := baseObjectToGoType(args[3])

			conn, _ := dbConnections[alias]
			switch c := conn.(type) {
			case *sql.DB:
				return newError("SQL update via HASH not implemented; use db.exec for now")
			case *mongo.Client:
				coll := c.Database("test").Collection(target)
				_, err := coll.UpdateMany(context.TODO(), match, update)
				if err != nil {
					return newError("mongodb update error: %s", err.Error())
				}
			}
			return TRUE
		},
	}

	builtins["db.delete"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			alias := args[0].(*object.String).Value
			target := args[1].(*object.String).Value
			match := baseObjectToGoType(args[2])

			conn, _ := dbConnections[alias]
			switch c := conn.(type) {
			case *sql.DB:
				return newError("SQL delete via HASH not implemented; use db.exec for now")
			case *mongo.Client:
				coll := c.Database("test").Collection(target)
				_, err := coll.DeleteMany(context.TODO(), match)
				if err != nil {
					return newError("mongodb delete error: %s", err.Error())
				}
			}
			return TRUE
		},
	}

	builtins["db.insert_many"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			aliasObj, ok1 := args[0].(*object.String)
			targetObj, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("first two arguments to `db.insert_many` must be STRING")
			}
			alias := aliasObj.Value
			target := targetObj.Value
			arr, ok := args[2].(*object.Array)
			if !ok {
				return newError("third argument to `db.insert_many` must be ARRAY")
			}

			conn, _ := dbConnections[alias]
			switch c := conn.(type) {
			case *sql.DB:
				for _, el := range arr.Elements {
					builtins["db.insert"].Fn(env, args[0], args[1], el)
				}
			case *mongo.Client:
				coll := c.Database("test").Collection(target)
				var docs []interface{}
				for _, el := range arr.Elements {
					docs = append(docs, baseObjectToGoType(el))
				}
				_, err := coll.InsertMany(context.TODO(), docs)
				if err != nil {
					return newError("mongodb insert_many error: %s", err.Error())
				}
			}
			return TRUE
		},
	}

	builtins["db.aggregate"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			alias := args[0].(*object.String).Value
			target := args[1].(*object.String).Value
			pipeline := baseObjectToGoType(args[2])

			conn, _ := dbConnections[alias]
			if c, ok := conn.(*mongo.Client); ok {
				coll := c.Database("test").Collection(target)
				cursor, err := coll.Aggregate(context.TODO(), pipeline)
				if err != nil {
					return newError("mongodb aggregate error: %s", err.Error())
				}
				var res []interface{}
				cursor.All(context.TODO(), &res)
				return goTypeToBaseObject(res)
			}
			return newError("aggregate is only supported for NoSQL (MongoDB) at the moment")
		},
	}
}

func stringsJoin(s []string, sep string) string {
	res := ""
	for i, v := range s {
		res += v
		if i < len(s)-1 {
			res += sep
		}
	}
	return res
}
