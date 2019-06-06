# gojsonexplode


## What is gojsonexplode?
gojsonexplode go library to flatten/explode nested JSON. 

## How does it work?
```go
package main

import (
    "fmt"

    "github.com/nytlabs/gojsonexplode"
)
func main() {
    input := `{"person":{"name":"Joe", "address":{"street":"123 Main St."}}}`
    out, err := gojsonexplode.Explodejsonstr(input, ".")
    if err != nil {
        // handle error
    }
    fmt.Println(out)
}
```

should print:
```javascript
{"person.address.street":"123 Main St.","person.name":"Joe"}
```

### How are JSON arrays handled?
JSON arrays are flattned using the parent attribute concatenated with a delimiter and the respective index for each of the elements in the array
```javascript
{"list":[true, false]}
``` 
gets exploded to: 
```javascript
{"list.0": true, "list.1":false}
```
