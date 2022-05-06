# go-config-reader

Simple configuration reader and validator that accepts a variety of sources.

## How to use

1. Import the library

```golang
import (
    cf "github.com/randlabs/go-config-reader"
)
```

2. Create a struct that defines your configuration settings and add the required JSON tags to it.

```golang
type ConfigurationSettings struct {
    Name         string  `json:"name"`
    IntegerValue int     `json:"integerValue"`
    FloatValue   float64 `json:"floatValue"`
}
```

3. Define the variable that will hold your settings

```golang
settings := ConfigurationSettings{}
```

4. Setup the configuration load options

```golang
opts := cf.Options{
	Source: "/tmp/settings.json"",
}
```

5. And execute load 

```golang
err := cf.Load(opts, &settings)
if err != nil {
	return err
}
```

## Loader options

The `Options` struct accept several modifiers that affects the load operation:

| Field                                            | Meaning                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|--------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Source`                                         | Specifies the configuration source. Optional.<br />Used mostly for testing or templating.                                                                                                                                                                                                                                                                                                                                                                                      |
| `EnvironmentVariable`                            | The environment variable used to lookup for the source. If specified, the source is the value of the environment variable. For example, this code:<br /><br />&nbsp;&nbsp;&nbsp;`opts.EnvironmentVariable = "MYSETTINGS"`<br /><br />expects you define an environment variable like this:<br /><br />&nbsp;&nbsp;&nbsp;`MYSETTINGS=/tmp/settings.json`<br /><br />so the source will be `/tmp/settings.json`<br /><sub>**NOTE**: `Source` has priority over this field.</sub> |
| `CmdLineParameter`<br />`CmdLineParameterShort`  | Long and short command-line parameters that contains source. Set to an empty string to disable. For example, this code:<br /><br />&nbsp;&nbsp;&nbsp;`s := "settings"`<br />&nbsp;&nbsp;&nbsp;`opts.CmdLineParameter = &s`<br /><br />expects you run your app like this: `yourapp --settings /tmp/settings.json`<br /><sub>**NOTE**: `EnvironmentVariable` has priority over this field.                                                                                      |
| `Callback`                                       | Use a custom loader for the configuration settings. For example:<br /><br />`func (ctx context.Context, source string) (string, error) {`<br />`&nbsp;&nbsp;dat, err := os.ReadFile(source)`<br />`&nbsp;&nbsp;if err != nil {`<br />`&nbsp;&nbsp;&nbsp;&nbsp;return "", err`<br />`&nbsp;&nbsp;}`<br />`&nbsp;&nbsp;return string(dat), nil`<br />`}`                                                                                                                         | 
| `Schema`                                         | Specifies an optional JSON schema to use to validate the loaded configuration. See [this page](https://json-schema.org/) for details about the schema format.                                                                                                                                                                                                                                                                                                                  |
| `ExtendedValidator`                              | Specifies a custom validator function. For example:<br /><br />`func (settings interface{}) error {`<br />`&nbsp;&nbsp;s := settings.(*ConfigurationSettings)`<br />`&nbsp;&nbsp;if s.IntegerValue < 0 {`<br />`&nbsp;&nbsp;&nbsp;&nbsp;return errors.New("invalid integer value")`<br />`&nbsp;&nbsp;}`<br />`&nbsp;&nbsp;s.IntegerValue *= 2 // You can also modify them at this stage`<br />`&nbsp;&nbsp;return nil`<br />`}`                                               |
| `Context`                                        | Optional `context.Context` object to use while loading the configuration.                                                                                                                                                                                                                                                                                                                                                                                                      |

## The source

The library contains several predefined loaders to load configuration settings from different locations. They are:

* A file path like `/tmp/settings.json` or in URL format `file:///tmp/settings.json` used to load the configuration settings from the specified file.
* An http or https URL like `https://configurations.company/network/myapp/settings.json`.
* An embedded data URL like `data://{ "integerValue": 10, .... }`. A JSON object like `{ "integerValue": 10, .... }` will be also taken as a data URL.
* A non-standard Hashicorp Vault URL with the following format: `vault://server-domain?token={access-token}&path={vault-path}`<br />In this case, the loader will try to reach the `vault-path` secret located at `http://server-domain` using the provided `access-token`.<br /><sub>NOTE: Use `vaults://` to access a server using the `https` protocol.</sub>

## Variable expansion

When data is loaded from the provided source, a macro expansion routine is executed. The following macros are processed:

* `${SRC:some-source}`: The loader will attempt to load the data located at `some-source` and replace the macro with it. `some-source` must be in any of the supported source formats.
* `${ENV:some-environment-variable}`: The loader will replace the macro with the content of the environment variable named `some-environment-variable`.

## LICENSE

See `LICENSE` file for details.