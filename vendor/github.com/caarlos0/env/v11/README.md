<p align="center">
  <img alt="GoReleaser Logo" src="https://becker.software/env.png" height="140" />
  <p align="center">A simple, zero-dependencies library to parse environment variables into structs.</p>
</p>

---

## Used and supported by

<p>
  <a href="https://encore.dev">
    <img src="https://user-images.githubusercontent.com/78424526/214602214-52e0483a-b5fc-4d4c-b03e-0b7b23e012df.svg" width="120px" alt="encore icon"></img>
  </a>
  <br/>
  <br/>
  <b>Encore – the platform for building Go-based cloud backends.</b>
  <br/>
</p>

## Example

Get the module with:

```sh
go get github.com/caarlos0/env/v11
```

The usage looks like this:

```go
package main

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type config struct {
	Home         string         `env:"HOME"`
	Port         int            `env:"PORT" envDefault:"3000"`
	Password     string         `env:"PASSWORD,unset"`
	IsProduction bool           `env:"PRODUCTION"`
	Duration     time.Duration  `env:"DURATION"`
	Hosts        []string       `env:"HOSTS" envSeparator:":"`
	TempFolder   string         `env:"TEMP_FOLDER,expand" envDefault:"${HOME}/tmp"`
	StringInts   map[string]int `env:"MAP_STRING_INT"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

  // or you can use generics
  cfg, err := env.ParseAs[config]()
  if err != nil {
		fmt.Printf("%+v\n", err)
  }

	fmt.Printf("%+v\n", cfg)
}
```

You can run it like this:

```sh
$ PRODUCTION=true HOSTS="host1:host2:host3" DURATION=1s MAP_STRING_INT=k1:1,k2:2 go run main.go
{Home:/your/home Port:3000 IsProduction:true Hosts:[host1 host2 host3] Duration:1s StringInts:map[k1:1 k2:2]}
```

## Caveats

> [!CAUTION]
>
> _Unexported fields_ will be **ignored** by `env`.
> This is by design and will not change.

## Supported types and defaults

Out of the box all built-in types are supported, plus a few others that
are commonly used.

Complete list:

- `string`
- `bool`
- `int`
- `int8`
- `int16`
- `int32`
- `int64`
- `uint`
- `uint8`
- `uint16`
- `uint32`
- `uint64`
- `float32`
- `float64`
- `time.Duration`
- `encoding.TextUnmarshaler`
- `url.URL`

Pointers, slices and slices of pointers, and maps of those types are also
supported.

You can also use/define a [custom parser func](#custom-parser-funcs) for any
other type you want.

You can also use custom keys and values in your maps, as long as you provide a
parser function for them.

If you set the `envDefault` tag for something, this value will be used in the
case of absence of it in the environment.

By default, slice types will split the environment value on `,`; you can change
this behavior by setting the `envSeparator` tag. For map types, the default
separator between key and value is `:` and `,` for key-value pairs.
The behavior can be changed by setting the `envKeyValSeparator` and
`envSeparator` tags accordingly.

## Custom Parser Funcs

If you have a type that is not supported out of the box by the lib, you are able
to use (or define) and pass custom parsers (and their associated `reflect.Type`)
to the `env.ParseWithOptions()` function.

In addition to accepting a struct pointer (same as `Parse()`), this function
also accepts a `Options{}`, and you can set your custom parsers in the `FuncMap`
field.

If you add a custom parser for, say `Foo`, it will also be used to parse
`*Foo` and `[]Foo` types.

Check the examples in the [go doc](http://pkg.go.dev/github.com/caarlos0/env/v11)
for more info.

### A note about `TextUnmarshaler` and `time.Time`

Env supports by default anything that implements the `TextUnmarshaler` interface.
That includes things like `time.Time` for example.
The upside is that depending on the format you need, you don't need to change
anything.
The downside is that if you do need time in another format, you'll need to
create your own type.

Its fairly straightforward:

```go
type MyTime time.Time

func (t *MyTime) UnmarshalText(text []byte) error {
	tt, err := time.Parse("2006-01-02", string(text))
	*t = MyTime(tt)
	return err
}

type Config struct {
	SomeTime MyTime `env:"SOME_TIME"`
}
```

And then you can parse `Config` with `env.Parse`.

## Required fields

The `env` tag option `required` (e.g., `env:"tagKey,required"`) can be added to
ensure that some environment variable is set. In the example above, an error is
returned if the `config` struct is changed to:

```go
type config struct {
	SecretKey string `env:"SECRET_KEY,required"`
}
```

> [!NOTE]
>
> Note that being set is not the same as being empty.
> If the variable is set, but empty, the field will have its type's default
> value.
> This also means that custom parser funcs will not be invoked.

## Expand vars

If you set the `expand` option, environment variables (either in `${var}` or
`$var` format) in the string will be replaced according with the actual value
of the variable. For example:

```go
type config struct {
	SecretKey string `env:"SECRET_KEY,expand"`
}
```

This also works with `envDefault`:

```go
import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type config struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"3000"`
	Address  string `env:"ADDRESS,expand" envDefault:"$HOST:${PORT}"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", cfg)
}
```

results in this:

```sh
$ PORT=8080 go run main.go
{Host:localhost Port:8080 Address:localhost:8080}
```

## Init `nil` pointers

You can automatically initialize `nil` pointers regardless of if a variable is
set for them or not.
This behavior can be enabled by using the `init` tag option.

Example:

```go
type config struct {
	URL *url.URL `env:"URL,init"`
}
```

## Not Empty fields

While `required` demands the environment variable to be set, it doesn't check
its value. If you want to make sure the environment is set and not empty, you
need to use the `notEmpty` tag option instead (`env:"SOME_ENV,notEmpty"`).

Example:

```go
type config struct {
	SecretKey string `env:"SECRET_KEY,notEmpty"`
}
```

## Unset environment variable after reading it

The `env` tag option `unset` (e.g., `env:"tagKey,unset"`) can be added
to ensure that some environment variable is unset after reading it.

Example:

```go
type config struct {
	SecretKey string `env:"SECRET_KEY,unset"`
}
```

## From file

The `env` tag option `file` (e.g., `env:"tagKey,file"`) can be added
in order to indicate that the value of the variable shall be loaded from a
file.
The path of that file is given by the environment variable associated with it:

```go
package main

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type config struct {
	Secret      string `env:"SECRET,file"`
	Password    string `env:"PASSWORD,file"           envDefault:"/tmp/password"`
	Certificate string `env:"CERTIFICATE,file,expand" envDefault:"${CERTIFICATE_FILE}"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	fmt.Printf("%+v\n", cfg)
}
```

```sh
$ echo qwerty > /tmp/secret
$ echo dvorak > /tmp/password
$ echo coleman > /tmp/certificate

$ SECRET=/tmp/secret  \
	CERTIFICATE_FILE=/tmp/certificate \
	go run main.go
{Secret:qwerty Password:dvorak Certificate:coleman}
```

## Options

### Use field names as environment variables by default

If you don't want to set the `env` tag on every field, you can use the
`UseFieldNameByDefault` option.

It will use the field name to define the environment variable name.
So, `Foo` becomes `FOO`, `FooBar` becomes `FOO_BAR`, and so on.

Here's an example:

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username     string // will use $USERNAME
	Password     string // will use $PASSWORD
	UserFullName string // will use $USER_FULL_NAME
}

func main() {
	cfg := &Config{}
	opts := env.Options{UseFieldNameByDefault: true}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

### Environment

By setting the `Options.Environment` map you can tell `Parse` to add those
`keys` and `values` as `env` vars before parsing is done.
These `envs` are stored in the map and never actually set by `os.Setenv`.
This option effectively makes `env` ignore the OS environment variables: only
the ones provided in the option are used.

This can make your testing scenarios a bit more clean and easy to handle.

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Password string `env:"PASSWORD"`
}

func main() {
	cfg := &Config{}
	opts := env.Options{Environment: map[string]string{
		"PASSWORD": "MY_PASSWORD",
	}}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

### Changing default tag name

You can change what tag name to use for setting the env vars by setting the
`Options.TagName` variable.

For example

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Password string `json:"PASSWORD"`
}

func main() {
	cfg := &Config{}
	opts := env.Options{TagName: "json"}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

### Prefixes

You can prefix sub-structs env tags, as well as a whole `env.Parse` call.

Here's an example flexing it a bit:

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Home string `env:"HOME"`
}

type ComplexConfig struct {
	Foo   Config `envPrefix:"FOO_"`
	Clean Config
	Bar   Config `envPrefix:"BAR_"`
	Blah  string `env:"BLAH"`
}

func main() {
	cfg := &ComplexConfig{}
	opts := env.Options{
		Prefix: "T_",
		Environment: map[string]string{
			"T_FOO_HOME": "/foo",
			"T_BAR_HOME": "/bar",
			"T_BLAH":     "blahhh",
			"T_HOME":     "/clean",
		},
	}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

### Complex objects inside array (slice)

You can set sub-struct field values inside a slice by naming the environment variables with sequential numbers starting from 0 (without omitting numbers in between) and an underscore.
It is possible to use prefix tag too.

Here's an example with and without prefix tag:

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Test struct {
	Str string `env:"STR"`
	Num int    `env:"NUM"`
}
type ComplexConfig struct {
	Baz []Test  `env:",init"`
	Bar []Test  `envPrefix:"BAR"`
	Foo *[]Test `envPrefix:"FOO_"`
}

func main() {
	cfg := &ComplexConfig{}
	opts := env.Options{
		Environment: map[string]string{
			"0_STR":     "bt",
			"1_NUM":     "10",

			"FOO_0_STR": "b0t",
			"FOO_1_STR": "b1t",
			"FOO_1_NUM": "212",

			"BAR_0_STR": "f0t",
			"BAR_0_NUM": "101",
			"BAR_1_STR": "f1t",
			"BAR_1_NUM": "111",
		},
	}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

### On set hooks

You might want to listen to value sets and, for example, log something or do
some other kind of logic.
You can do this by passing a `OnSet` option:

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username string `env:"USERNAME" envDefault:"admin"`
	Password string `env:"PASSWORD"`
}

func main() {
	cfg := &Config{}
	opts := env.Options{
		OnSet: func(tag string, value interface{}, isDefault bool) {
			fmt.Printf("Set %s to %v (default? %v)\n", tag, value, isDefault)
		},
	}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

## Making all fields to required

You can make all fields that don't have a default value be required by setting
the `RequiredIfNoDef: true` in the `Options`.

For example

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username string `env:"USERNAME" envDefault:"admin"`
	Password string `env:"PASSWORD"`
}

func main() {
	cfg := &Config{}
	opts := env.Options{RequiredIfNoDef: true}

	// Load env vars.
	if err := env.ParseWithOptions(cfg, opts); err != nil {
		log.Fatal(err)
	}

	// Print the loaded data.
	fmt.Printf("%+v\n", cfg)
}
```

## Defaults from code

You may define default value also in code, by initialising the config data
before it's filled by `env.Parse`.
Default values defined as struct tags will overwrite existing values during
Parse.

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username string `env:"USERNAME" envDefault:"admin"`
	Password string `env:"PASSWORD"`
}

func main() {
	cfg := Config{
		Username: "test",
		Password: "123456",
	}

	if err := env.Parse(&cfg); err != nil {
		fmt.Println("failed:", err)
	}

	fmt.Printf("%+v", cfg) // {Username:admin Password:123456}
}
```

## Error handling

You can handle the errors the library throws like so:

```go
package main

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Username string `env:"USERNAME" envDefault:"admin"`
	Password string `env:"PASSWORD"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if e, ok := err.(*env.AggregateError); ok {
		for _, er := range e.Errors {
			switch v := er.(type) {
			case env.ParseError:
				// handle it
			case env.NotStructPtrError:
				// handle it
			case env.NoParserError:
				// handle it
			case env.NoSupportedTagOptionError:
				// handle it
			default:
				fmt.Printf("Unknown error type %v", v)
			}
		}
	}

	fmt.Printf("%+v", cfg) // {Username:admin Password:123456}
}
```

> **Info**
>
> If you want to check if an specific error is in the chain, you can also use
> `errors.Is()`.

## Badges

[![Release](https://img.shields.io/github/release/caarlos0/env.svg?style=for-the-badge)](https://github.com/goreleaser/goreleaser/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Build status](https://img.shields.io/github/actions/workflow/status/caarlos0/env/build.yml?style=for-the-badge&branch=main)](https://github.com/caarlos0/env/actions?workflow=build)
[![Codecov branch](https://img.shields.io/codecov/c/github/caarlos0/env/main.svg?style=for-the-badge)](https://codecov.io/gh/caarlos0/env)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](http://godoc.org/github.com/caarlos0/env/v11)
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)

## Related projects

- [envdoc](https://github.com/g4s8/envdoc) - generate documentation for environment variables from `env` tags

## Stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/env.svg)](https://starchart.cc/caarlos0/env)
