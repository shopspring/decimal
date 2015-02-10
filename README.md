# decimal [![Build Status](https://travis-ci.org/jellolabs/decimal.png?branch=master)](https://travis-ci.org/jellolabs/decimal)

Arbitrary-precision fixed-point decimal numbers in go.

NOTE: can "only" represent numbers with a maximum of 2^31 digits after the decmial point.

## Features

 * addition, subtraction, multiplication with no loss of precision
 * division with specified precision
 * database/sql serialization/deserialization
 * json and xml serialization/deserialization

## Notable users

This is currently being used in production by [Spring](https://shopspring.com/), and has been since August 14, 2014.

## Usage

```go
package main

import (
    "fmt"
    "github.com/jellolabs/decimal"
)

func main() {
	price, err := decimal.NewFromString("136.02")
    if err != nil {
        panic(err)
    }

	quantity := decimal.NewFromFloat(3)

	fee, _ := decimal.NewFromString(".035")
	taxRate, _ := decimal.NewFromString(".08875")

    subtotal := price.Mul(quantity)

    preTax := subtotal.Mul(fee.Add(decimal.NewFromFloat(1)))

    total := preTax.Mul(taxRate.Add(decimal.NewFromFloat(1)))

	fmt.Println("Subtotal:", subtotal)
	fmt.Println("Pre-tax:", preTax)
    fmt.Println("Taxes:", total.Sub(preTax))
	fmt.Println("Total:", total)
	fmt.Println("Tax rate:", total.Sub(preTax).Div(preTax))
}
```

## Documentation

http://godoc.org/github.com/jellolabs/decimal

## License

The MIT License (MIT)
