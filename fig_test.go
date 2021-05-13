package fig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	configuration = nil

	assert.Equal(t, true, Bool("boolean"))

	assert.Equal(t, true, BoolOr(true, "something"))
	assert.Equal(t, true, BoolOr(false, "boolean"))
}

func TestInt(t *testing.T) {
	configuration = nil

	assert.Equal(t, 123, Int("integer"))

	assert.Equal(t, 456, IntOr(456, "something"))
	assert.Equal(t, 123, IntOr(456, "integer"))
}

func TestString(t *testing.T) {
	configuration = nil

	assert.Equal(t, "hello", String("string"))

	assert.Equal(t, "world", StringOr("world", "something"))
	assert.Equal(t, "hello", StringOr("world", "string"))
	assert.Equal(t, "from-config", String("fg.config.env"))
}

func TestFloat(t *testing.T) {
	configuration = nil

	assert.Equal(t, 12.34, Float("floating-point"))

	assert.Equal(t, 34.12, FloatOr(34.12, "something"))
	assert.Equal(t, 12.34, FloatOr(34.12, "floating-point"))
}

func TestOverriding(t *testing.T) {

	configuration = nil
	assert.Equal(t, "json", String("tooverride"))

	// now lets change FileOrder
	oldOrder := EnvironmentOrder
	configuration = nil
	EnvironmentOrder = []string{"config", "dev"}
	Reset()
	assert.Equal(t, "yaml", String("tooverride"))
	EnvironmentOrder = oldOrder
	Reset()
}

func TestPanic(t *testing.T) {
	configuration = nil
	assert.Panics(t, func() {
		String("something")
	})
}

func TestStruct(t *testing.T) {
	configuration = nil

	type Parent struct {
		IntValue      int
		StringValue   string
		OptionalValue string `fig:"optional"`
	}

	var exact, lower, snake, url, optional Parent

	Struct(&exact, "struct", "exact")
	assert.Equal(t, 123, exact.IntValue)
	assert.Equal(t, "abc", exact.StringValue)

	Struct(&lower, "struct", "lower")
	assert.Equal(t, 123, lower.IntValue)
	assert.Equal(t, "abc", lower.StringValue)

	Struct(&snake, "struct", "snake")
	assert.Equal(t, 123, snake.IntValue)
	assert.Equal(t, "abc", snake.StringValue)

	Struct(&url, "struct", "url")
	assert.Equal(t, 123, url.IntValue)
	assert.Equal(t, "abc", url.StringValue)

	Struct(&optional, "struct", "optional")
	assert.Equal(t, 123, optional.IntValue)
	assert.Equal(t, "abc", optional.StringValue)
	assert.Equal(t, "def", optional.OptionalValue)
}

func TestMap(t *testing.T) {
	configuration = nil
	data := Map("map")
	assert.Equal(t, 1234, data["field-a"])
	assert.Equal(t, "abcd", data["field-b"])
}

func TestStringSlice(t *testing.T) {
	configuration = nil

	splice := StringSlice("string-slice")
	assert.Equal(t, []string{"str1", "str2", "123"}, splice)
}

func TestStringSliceOr(t *testing.T) {
	configuration = nil

	splice := StringSliceOr([]string{"def1"}, "string-slice-or")
	assert.Equal(t, []string{"def1"}, splice)
}

func TestValueOf(t *testing.T) {
	configuration = nil

	data := Interface("arr-json")
	arr := data.([]interface{})

	assert.Equal(t, 2, len(arr))
}
func TestSSMString(t *testing.T) {
	configuration = nil

	str := String("ssm-string.password")
	str2 := String("ssm-string.store-password")

	assert.Equal(t, "TheReaper61", str)
	assert.Equal(t, "SierraTango@61", str2)
}

func TestSSMStringOr(t *testing.T) {
	configuration = nil

	assert.Equal(t, "optional", StringOr("optional", "ssm-string.optional"))
}

func TestSSMStruct(t *testing.T) {
	configuration = nil

	type ConfigStruct struct {
		Password      string `json:"password"`
		StorePassword string `json:"store-password"`
		UserName      string `json:"username"`
	}

	cs := ConfigStruct{}

	Struct(&cs, "ssm-struct", "struct")
	assert.Equal(t, "SierraTango@61", cs.StorePassword)
	assert.Equal(t, "ssm@user", cs.UserName)
}
