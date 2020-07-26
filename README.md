# Bedis

starting a Redis server and using it as the "Builtin Redis"

## Usage

```golang
package main

import (
    "fmt"

    "github.com/wrfly/bedis"
)

func main() {
    builtin, err := bedis.New(bedis.Option{
        Memory: "3GB",
    })
    if err != nil {
        panic(err)
    }
    defer builtin.StopAndClose()

    client, err := builtin.DefaultClient()
    if err != nil {
        panic(err)
    }

    // set key
    if err := client.Set("key", 12345, -1).Err(); err != nil {
        panic(err)
    }
    // get key
    if v, err := client.Get("key").Result(); err == nil {
        fmt.Println("redis get value", v)
    }
}
```
