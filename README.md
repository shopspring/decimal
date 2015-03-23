# decimal [![Build Status](https://travis-ci.org/shopspring/decimal.png?branch=master)](https://travis-ci.org/shopspring/decimal)

Arbitrary-precision fixed-point decimal numbers in go.

NOTE: can "only" represent numbers with a maximum of 2^31 digits after the decmial point.

## Features

 * the zero-value is 0, and is safe to use without initialization
 * addition, subtraction, multiplication with no loss of precision
 * division with specified precision
 * database/sql serialization/deserialization
 * json and xml serialization/deserialization

## Install

Run `go get github.com/shopspring/decimal`

## Usage

```go
package main

import (
    "fmt"
    "github.com/shopspring/decimal"
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
    
	fmt.Println("Subtotal:", subtotal)                      // => Subtotal: 408.06
	fmt.Println("Pre-tax:", preTax)                         // => Pre-tax: 422.3421
    fmt.Println("Taxes:", total.Sub(preTax))                // => Taxes: 37.482861375
	fmt.Println("Total:", total)                            // => Total: 459.824961375
	fmt.Println("Tax rate:", total.Sub(preTax).Div(preTax)) // => Tax rate: 0.08875
}
```

## Documentation

http://godoc.org/github.com/shopspring/decimal

## Production Usage

* [Spring](https://shopspring.com/), since August 14, 2014.
* If you are using this in production, please let us know!

## License

The MIT License (MIT)

This is a heavily modified fork of [fpd.Decimal](https://github.com/oguzbilgic/fpd), which was also released under the MIT License.
