# go-xsd-validator

## Overview
`go-xsd-validator` is a Go-native tool for validating XML files against XSD
schemas. It ensures that XML documents conform to the given XSD definitions with
high performance and reliability.

This tool is in its early stages of development and may contain missing features
or potential issues. I am not an expert in XML or XSD, so there might be aspects
I have overlooked. Feedback and contributions are highly appreciated to improve
its functionality.

This tool was developed out of necessity in my current job, and it remains open
to further development and enhancements.

## Prerequisites
- Go 1.23.3 or higher

## Installation
Clone the repository and navigate to the project directory:

```sh
git clone https://github.com/sergenyalcin/go-xsd-validator.git
cd go-xsd-validator
```

## Usage
To validate an XML file against an XSD schema, use the following command:

```sh
go run cmd/main.go -xml <path-to-xml> -xsd <path-to-xsd>
```

(Include a sample XML validation example here)

## Running Tests
Unit tests are included in the repository. To run them, use:

```sh
make unit-test
```

For race condition detection:

```sh
make unit-test-race
```

For test coverage analysis:

```sh
make unit-test-cover
```

## Contribution
Contributions are welcome! You can contribute by:
- Forking the repository and creating a pull request
- Opening an issue with suggestions or bug reports

## License
This project is licensed under the Apache-2.0 License.

