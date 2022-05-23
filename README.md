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

| Field                                            | Meaning                                                                                                                                                                                                                                                                                                                                                                                                                   |
|--------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Source`                                         | Specifies the configuration source. Optional.<br />Used mostly for testing or templating.                                                                                                                                                                                                                                                                                                                                 |
| `EnvironmentVariable`                            | The environment variable used to lookup for the source. If specified, the source is the value of the environment variable. For example, this code:<br /><pre>opts.EnvironmentVariable = "MYSETTINGS"</pre>expects you define an environment variable like this:<br /><pre>MYSETTINGS=/tmp/settings.json</pre>so the source will be: `/tmp/settings.json`<br /><sub>**NOTE**: `Source` has priority over this field.</sub> |
| `CmdLineParameter`<br />`CmdLineParameterShort`  | Long and short command-line parameters that contains source. Set to an empty string to disable. For example, this code:<br /><pre>s := "settings"<br />opts.CmdLineParameter = &s</pre>expects you run your app like this: `yourapp --settings /tmp/settings.json`<br /><sub>**NOTE**: `EnvironmentVariable` has priority over this field.                                                                                |
| `Callback`                                       | Use a custom loader for the configuration settings. For example:<br /><pre>func (ctx context.Context, source string) (string, error) {<br />        dat, err := os.ReadFile(source)<br />        if err != nil {<br />                return "", err<br />        }<br />        return string(dat), nil<br />}</pre>                                                                                                     | 
| `Schema`                                         | Specifies an optional JSON schema to use to validate the loaded configuration. See [this page](https://json-schema.org/) for details about the schema format.                                                                                                                                                                                                                                                             |
| `ExtendedValidator`                              | Specifies a custom validator function. For example:<br /><pre>func (settings interface{}) error {<br />        s := settings.(*ConfigurationSettings)<br />        if s.IntegerValue < 0 {<br />                return errors.New("invalid integer value")<br />        }<br />        s.IntegerValue *= 2 // You can also modify them at this stage<br />        return nil<br />}</pre>                                 |
| `Context`                                        | Optional `context.Context` object to use while loading the configuration.                                                                                                                                                                                                                                                                                                                                                 |

## The source

The library contains several predefined loaders to load configuration settings from different locations. They are:

* A file path like `/tmp/settings.json` or in URL format `file:///tmp/settings.json` used to load the configuration settings from the specified file.<br /><br />
* An http or https URL like `https://configurations.company/network/myapp/settings.json`. <br /><br />
* An embedded data URL like `data://{ "integerValue": 10, .... }`. A JSON object like `{ "integerValue": 10, .... }` will be also taken as a data URL.<br /><br />
* A [Hashicorp Vault](https://www.vaultproject.io/) URL using a custom scheme: `vault://server-domain?token={access-token}&path={vault-path}` <br />In this case, the loader will try to reach the `vault-path` secret located at `http://server-domain` using the provided `access-token`.<br /><br />By default, all the keys are read and returned as a JSON object but you can additionally add the `&key={key}` query parameter in order to read the specified value.<br /><br />NOTES:<br />1. `{vault-path}` usually starts with `/secret` or `/secret/data` for KV engines v1 and v2 respectively.<br />2. Use `vaults://` to access a server using the `https` protocol. 

## Variable expansion

When data is loaded from the provided source, a macro expansion routine is executed. The following macros are processed:

* `${SRC:some-source}`: The loader will attempt to load the data located at `some-source` and replace the macro with it. `some-source` must be in any of the supported source formats.<br /><br />
* `${ENV:some-environment-variable}`: The loader will replace the macro with the content of the environment variable named `some-environment-variable`.

You can also embed macros inside other macros, for example:

`${SRC:vault://my-vault-server.network?token=${ENV:VAULT_ACCESS_TOKEN}&path=/secret/data/mysecret}`

## LICENSE

See `LICENSE` file for details.
