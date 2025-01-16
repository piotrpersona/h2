# h2

Heic 2 PNG converter.

I use it for personal conversion of invoices photos done by IPhone.

## Requirements

- go 1.23 for installation

## Installation

```sh
go install github.com/piotrpersona/h2@latest
```

## Usage

Start:
```sh
h2 -input '<input DIR with HEIC>' -output '<output DIR with PNG>' &
```

Stop using kill:
```sh
ps aux | grep h2
```
