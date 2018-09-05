package gojsonexplode

import (
	"testing"
)

func TestTrue(t *testing.T) {
	input := `[true]`
	output := `{"0":true}`

	out, _ := Explodejsonstr(input, ".")
	if out != output {
		t.Error("got", out)
	}
}

func TestNull(t *testing.T) {
	input := `[null]`
	output := `{"0":null}`

	out, _ := Explodejsonstr(input, ".")
	if out != output {
		t.Error("got", out)
	}
}

func TestNesting(t *testing.T) {
	input := `{"person":{"name":"Joe", "address":{"street":"123 Main St."}}}`
	output := `{"person.address.street":"123 Main St.","person.name":"Joe"}`
	out, _ := Explodejsonstr(input, ".")
	if out != output {
		t.Error("got", out)
	}
}

func TestItems(t *testing.T) {
	input := `
    [
        {
            "description": "a schema given for items",
            "schema": {
                "items": {"type": "integer"}
            },
            "tests": [
                {
                    "description": "valid items",
                    "data": [ 1, 2, 3 ],
                    "valid": true
                },
                {
                    "description": "wrong type of items",
                    "data": [1, "x"],
                    "valid": false
                },
                {
                    "description": "ignores non-arrays",
                    "data": {"foo" : "bar"},
                    "valid": true
                }
            ]
        },
        {
            "description": "an array of schemas for items",
            "schema": {
                "items": [
                    {"type": "integer"},
                    {"type": "string"}
                ]
            },
            "tests": [
                {
                    "description": "correct types",
                    "data": [ 1, "foo" ],
                    "valid": true
                },
                {
                    "description": "wrong types",
                    "data": [ "foo", 1 ],
                    "valid": false
                }
            ]
        }
    ]`
	output := `{"0.description":"a schema given for items","0.schema.items.type":"integer","0.tests.0.data.0":1,"0.tests.0.data.1":2,"0.tests.0.data.2":3,"0.tests.0.description":"valid items","0.tests.0.valid":true,"0.tests.1.data.0":1,"0.tests.1.data.1":"x","0.tests.1.description":"wrong type of items","0.tests.1.valid":false,"0.tests.2.data.foo":"bar","0.tests.2.description":"ignores non-arrays","0.tests.2.valid":true,"1.description":"an array of schemas for items","1.schema.items.0.type":"integer","1.schema.items.1.type":"string","1.tests.0.data.0":1,"1.tests.0.data.1":"foo","1.tests.0.description":"correct types","1.tests.0.valid":true,"1.tests.1.data.0":"foo","1.tests.1.data.1":1,"1.tests.1.description":"wrong types","1.tests.1.valid":false}`
	out, _ := Explodejsonstr(input, ".")
	if out != output {
		t.Error("got", out)
	}
}
