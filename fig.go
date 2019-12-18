package fig

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/jacobstr/confer"
	"github.com/pkg/errors"
	"github.com/serenize/snaker"
)

// TODO:
// - allow command line arguments to be passed to override configurations
// - eg. --APP_PORT=1234, should be bubbled to the right place
var configuration *confer.Config

const supportedProtocolPrefix = "ssm@"

var ssmMap = make(map[string]string, 0)

// FileOrder contains the name, path and order in which the configuration
// files will be read. The first file will have highest priority
var FileOrder = []string{
	// highest
	"./config/production.yaml",
	"./config/production.json",
	"production.yaml",
	"production.json",

	// medium
	"./config/dev.yaml",
	"./config/dev.json",
	"dev.yaml",
	"dev.json",

	// lowest
	"./config/config.yaml",
	"./config/config.json",
	"config.yaml",
	"config.json",
}

// ShowFiles if set to true, will output the files to StdOut
// which were read to build the overall configuration
var ShowFiles = false

func init() {
	files := make([]string, 0)

	env := os.Getenv("FG_CONFIG_ENV")

	switch {
	case env != "" && env != "dev":
		files = append(files, []string{
			fmt.Sprintf("./config/%s.yaml", env),
			fmt.Sprintf("./config/%s.json", env),
			fmt.Sprintf("%s.yaml", env),
			fmt.Sprintf("%s.json", env),
		}...)
	default:
		files = append(files, []string{
			"./config/dev.yaml",
			"./config/dev.json",
			"dev.yaml",
			"dev.json",
		}...)
	}

	files = append(files, []string{
		"./config/config.yaml",
		"./config/config.json",
		"config.yaml",
		"config.json",
	}...)

	FileOrder = files

	jumpstart()
}

// jumpstart loads the configuration
func jumpstart() {

	var filesOk = []string{}

	// loop in reverse order (lowest priority first)
	for i := len(FileOrder) - 1; i >= 0; i-- {
		tmp := confer.NewConfig()
		err := tmp.ReadPaths(FileOrder[i])
		if err == nil {
			// file found (path is good)
			abs, _ := filepath.Abs(FileOrder[i])
			filesOk = append(filesOk, abs)
		}
	}

	configuration = confer.NewConfig()
	configuration.ReadPaths(filesOk...)
	if ShowFiles {
		fmt.Println("fig configuration loaded:", strings.Join(filesOk, " → "))
	}
}

// Exists checks if the given key is present.
// It also loads configuration (if missing)
func Exists(keys ...string) bool {
	if configuration == nil {
		jumpstart()
	}

	key := strings.Join(keys, ".")
	return configuration.IsSet(key)
}

// MustExist panics if the given configuration key is missing
func MustExist(key string) {
	if !Exists(key) {
		panic("configuration key missing:" + key)
	}
}

// IntOr returns the int value at the given key.
// If key is missing it returns defaultVal
func IntOr(defaultVal int, keys ...string) int {
	key := strings.Join(keys, ".")
	if !Exists(key) {
		return defaultVal
	}

	return configuration.GetInt(key)
}

// Int returns the int value at the given key.
// Panics if the key is missing.
func Int(keys ...string) int {
	key := strings.Join(keys, ".")
	MustExist(key)
	return configuration.GetInt(key)
}

// FloatOr returns the float64 value at the given key.
// If key is missing it returns defaultVal
func FloatOr(defaultVal float64, keys ...string) float64 {
	key := strings.Join(keys, ".")
	if !Exists(key) {
		return defaultVal
	}

	return configuration.GetFloat64(key)
}

// Float returns the float value at the given key.
// Panics if the key is missing.
func Float(keys ...string) float64 {
	key := strings.Join(keys, ".")
	MustExist(key)
	return configuration.GetFloat64(key)
}

// StringOr returns the string value at the given key.
// If key is missing it returns defaultVal
func StringOr(defaultVal string, keys ...string) string {
	key := strings.Join(keys, ".")
	if !Exists(key) {
		return defaultVal
	}

	val := configuration.GetString(key)
	if _, ok := ssmMap[val]; ok {
		return ssmMap[val]
	}

	if strings.HasPrefix(val, supportedProtocolPrefix) {
		ssmVal, err := FetchFromSSM(val)
		if err != nil {
			return defaultVal
		}
		ssmMap[val] = *ssmVal
		return ssmMap[val]
	}

	ssmMap[key] = val
	return ssmMap[key]
}

// String returns the string value at the given key.
// Panics if the key is missing.
func String(keys ...string) string {
	key := strings.Join(keys, ".")
	MustExist(key)

	val := configuration.GetString(key)
	if _, ok := ssmMap[val]; ok {
		return ssmMap[val]
	}

	if strings.HasPrefix(val, supportedProtocolPrefix) {
		ssmVal, err := FetchFromSSM(val)
		if err != nil {
			fmt.Printf("cannot fetch value from SSM for key: %s", key)
			return val
		}

		ssmMap[val] = *ssmVal
		return ssmMap[val]
	}

	ssmMap[key] = val
	return ssmMap[key]
}

// BoolOr returns the bool value at the given key.
// If key is missing it returns defaultVal
func BoolOr(defaultVal bool, keys ...string) bool {
	key := strings.Join(keys, ".")
	if !Exists(key) {
		return defaultVal
	}

	return configuration.GetBool(key)
}

// Bool returns the bool value at the given key.
// Panics if the key is missing.
func Bool(keys ...string) bool {
	key := strings.Join(keys, ".")
	MustExist(key)
	return configuration.GetBool(key)
}

// Map returns a map a the given key.
// Panics if the key is missing.
func Map(keys ...string) map[string]interface{} {
	key := strings.Join(keys, ".")
	MustExist(key)
	return configuration.GetStringMap(key)
}

// StringSlice returns a splice of string at the given key.
// Panics if the key is not present.
func StringSlice(keys ...string) []string {
	key := strings.Join(keys, ".")
	MustExist(key)
	return configuration.GetStringSlice(key)
}

// StringSliceOr returns a splice of string at the given key.
// If key is missing it returns defaultVal.
func StringSliceOr(defaultVal []string, keys ...string) []string {
	key := strings.Join(keys, ".")
	if !Exists(key) {
		return defaultVal
	}
	return configuration.GetStringSlice(key)
}

// Struct is used to parse and load simple structures. Most common use is
// reading connection strings.
// Note that it does not work for nested structs or arrays
func Struct(addr interface{}, keys ...string) {

	container := strings.Join(keys, ".")

	// addr's underlying type should be a struct address
	utype := reflect.TypeOf(addr)
	if utype.Kind() != reflect.Ptr || utype.Elem().Kind() != reflect.Struct {
		panic("Struct() method expects address of a struct")
	}

	rtype := reflect.TypeOf(addr).Elem()
	rvalue := reflect.ValueOf(addr).Elem()

	// loop: all the fields of this type
	for i := 0; i < rtype.NumField(); i++ {
		ftype := rtype.Field(i)
		fname := ftype.Name // field name

		// lookup the key in a number of formats
		// eg, a field name called NumItems could be read as:
		//  - NumItems  [exact]
		//  - numitems  [lowercase]
		//  - num_items [snakecase]
		//  - num-items [urlcase]
		lookup := []string{fname, strings.ToLower(fname), snaker.CamelToSnake(fname), strings.Replace(snaker.CamelToSnake(fname), "_", "-", -1)}
		found := false
		for _, amatch := range lookup {
			//fmt.Println("checking:", container, amatch)
			if Exists(container, amatch) {
				found = true
				switch fmt.Sprintf("%s", ftype.Type) {
				case "string":
					val := String(container, amatch)
					rvalue.Field(i).SetString(val)
				case "int":
					val := Int(container, amatch)
					rvalue.Field(i).SetInt(int64(val))
				default:
					panic("Can not read any value other than string|int")
				}
			}

		}
		if !found {
			// if the field was not an optional one, then throw error
			if !strings.Contains(ftype.Tag.Get("fig"), "optional") {
				panic("Can not find value for: " + container + ".(" + strings.Join(lookup, "|") + ")")
			}
		}
	}
}

// FetchFromSSM looks up the AWS SSM for the given key and returns the result.
func FetchFromSSM(ssmString string) (*string, error) {

	// Separate the look-up key from the ssmString
	// e.g. ssm@key
	key := ssmString[strings.Index(ssmString, "@")+1:]

	awsRegion := os.Getenv("AWS_REGION")
	// use the default aws region
	if awsRegion == "" {
		awsRegion = "ap-south-1"
	}

	config := &aws.Config{
		Region: aws.String(awsRegion),
	}

	awsSess, err := session.NewSessionWithOptions(session.Options{
		Config:            *config,
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		panic(errors.Wrap(err, "cannot create aws session for fetching creds form SSM"))
	}

	ssmSvc := ssm.New(awsSess, config)

	val, err := ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {

			return nil, errors.Wrapf(awsErr, "cannot fetch value from SSM corresponding to key %s", key)
		}
	}

	return val.Parameter.Value, nil
}
