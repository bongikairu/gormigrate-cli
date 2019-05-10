# Gormigrate CLI

Gormigrate CLI is a helper command file allowing you to easily integrate migration workflow to your gorm-based project. It aims to mimic how popular framework in other language does migration CLI.

## Installation

Just drop main.go into migrations folder of your choice

You also need (of course) [gorm](https://gorm.io/), [gormigrate](https://github.com/go-gormigrate/gormigrate), and [viper](https://github.com/spf13/viper)

## Usage

To create new migration file

    go run main.go --make Your Migration Name

To apply migration

    go run main.go
    