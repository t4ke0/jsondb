[![Go Report Card](https://goreportcard.com/badge/github.com/t4ke0/jsondb)](https://goreportcard.com/report/github.com/t4ke0/jsondb)

## jsondb


use json file as your database.


### Install

```bash
go get -v github.com/t4ke0/jsondb
```


### Usage

```golang
package main

import "github.com/t4ke0/jsondb"

const filename string = "db.json"

type Table struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
   conn, err := jsondb.Connect[Table](filename)
   if err != nil {
        ...
   }
   defer conn.Close()

   if err := conn.Init(); err != nil {
        ...
   }

   content, err := conn.ReadFromDB()
   if err != nil {
        ...
   }

   if err := conn.WriteToDB(Table{Id: 1, Name: "jsondb"}); err != nil {
        ...
   }
}
```

for more examples see `examples` folder.
