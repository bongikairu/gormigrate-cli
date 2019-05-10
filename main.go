/*

	gormigrate cli

	Cli for easy gormigrate integration

	Setup: Drop this file into migrations folder of your choice

		Usage:

		To create new migration file

			go run main.go --make Your Migration Name

		To apply migration

			go run main.go

		See gorm and gormigrate for more info

	Author: github.com/bongikairu
	License: MIT

*/

package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"gopkg.in/gormigrate.v1"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Migrations struct {
}

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "--make" {
		_, file, _, _ := runtime.Caller(0)
		title := "auto"
		if len(os.Args) >= 3 {
			title = strings.Join(os.Args[2:], " ")
			title = strings.ToLower(title)
			r, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
			title = r.ReplaceAllString(title, "")
			title = strings.Replace(title, " ", "_", -1)
		}
		timestamp := time.Now().Format("20060102150405")
		new_filename := fmt.Sprintf("%s_%s.go", timestamp, title)
		new_filepath := filepath.Join(filepath.Dir(file), new_filename)
		new_filedata := `
package main

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func (m Migrations) M_%s_%s(tx *gorm.DB) (*gormigrate.Migration) {
	// see example at https://github.com/go-gormigrate/gormigrate
	return &gormigrate.Migration{
		ID: "%s",
		Migrate: func(tx *gorm.DB) error {
			type NewModel struct {
				gorm.Model
				Name string
			}
			return tx.AutoMigrate(&NewModel{}).Error
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.DropTable("newmodels").Error
		},
	}
}
`
		new_filedata = fmt.Sprintf(new_filedata, timestamp, title, timestamp)
		fmt.Printf("Creating New Migration File %s\n", new_filename)
		ioutil.WriteFile(new_filepath, []byte(new_filedata), 0644)
		os.Exit(0)
		return
	}

	config := viper.New()
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetDefault("database.dialect", "mysql")
	config.SetDefault("database.url", "root:@tcp(localhost:3306)/db")
	config.SetDefault("database.debug", "1")

	db, err := gorm.Open(config.Get("database.dialect").(string), config.Get("database.url").(string))
	if err != nil {
		log.Fatalf("Can't connect to Database : %s", err)
	}
	db.LogMode(config.Get("database.debug").(string) == "1")

	migration_by_id := map[string]*gormigrate.Migration{}
	migration_id_list := []string{}

	migration := Migrations{}
	migrationType := reflect.TypeOf(migration)
	for i := 0; i < migrationType.NumMethod(); i++ {
		method := migrationType.Method(i)
		method_name := method.Name
		m := method.Func.Call([]reflect.Value{reflect.ValueOf(migration), reflect.ValueOf(db)})[0].Interface().(*gormigrate.Migration)
		if !strings.HasPrefix(method_name, fmt.Sprintf("M_%s_", m.ID)) {
			log.Fatalf("Migration %s has unmatched ID %s", method_name, m.ID)
		}
		migration_by_id[m.ID] = m
		migration_id_list = append(migration_id_list, m.ID)
	}

	if len(migration_id_list) == 0 {
		log.Fatalf("No migration available, try making one with go `run main.go --make`")
	}

	sort.Strings(migration_id_list)

	migration_list := []*gormigrate.Migration{}
	for _, m_id := range migration_id_list {
		migration_list = append(migration_list, migration_by_id[m_id])
	}

	// TODO: Bind arguments for specific migrate or rollback, fake, dryrun, list migration, and more (<3 django)
	gm := gormigrate.New(db, gormigrate.DefaultOptions, migration_list)
	if err = gm.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
	log.Printf("Migration did run successfully")
}
