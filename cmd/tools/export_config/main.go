//
// Export configuration to .env file with default values.
//

package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/labstack/gommon/log"

	"github.com/hasansino/apptemplate/internal/config"
)

type item struct {
	Comment string
	Name    string
	Default string
}

var exported = make([]item, 0)

func main() {
	export(&config.Config{}, ``)

	f, err := os.Create(`.env`)
	if err != nil {
		log.Fatalf("failed to create .env file: %v", err)
	}

	_, _ = f.WriteString("# Service default configuration")
	_, _ = f.WriteString("\n\n")

	for i := range exported {
		if len(exported[i].Comment) > 0 {
			_, _ = f.WriteString("\n" + exported[i].Comment + "\n")
		} else {
			_, _ = f.WriteString(fmt.Sprintf("%s=%s\n", exported[i].Name, exported[i].Default))
		}
	}

	if err := f.Sync(); err != nil {
		log.Fatalf("failed to sync .env file: %v", err)
	}

	fmt.Println("OK")
}

func export(i interface{}, prefix string) {
	var (
		rt = reflect.TypeOf(i)
		rv = reflect.ValueOf(i)
	)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		var (
			field     = rt.Field(i)
			value     = rv.FieldByName(field.Name)
			fieldPath = prefix + field.Name
		)

		if len(field.PkgPath) != 0 { // unexported
			continue
		}
		if field.Name == `RWMutex` {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			exported = append(exported, item{
				Comment: "# " + fieldPath,
			})
			export(value.Addr().Interface(), fieldPath+".")
		default:
			if envVarName := field.Tag.Get(`env`); len(envVarName) > 0 {
				exported = append(exported, item{
					Name: envVarName, Default: field.Tag.Get(`default`),
				})
			}
		}
	}
}
