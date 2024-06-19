// Pages 225-226
// Listing 10-3: Creating a TOML configuration adapter and registering it with Caddy.
package tomladapter

import (
	"encoding/json"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/pelletier/go-toml"
)

// Registers the configuration adapter with Caddy.
// For this, we include a call to caddyconfig.RegisterAdapter and pass it the
// adapter's type and adapter object implementing the caddyconfig.Adapter interface.
// When we import this module from Caddy's main.go file, the configuration
// adapter registers itself with Caddy, adding support for parsing the TOML
// configuration file.
func init() {
	caddyconfig.RegisterAdapter("toml", Adapter{})
}

// Adapter converts a TOML Caddy configuration to JSON.
type Adapter struct{}

// Adapt the TOML body to JSON.
func (a Adapter) Adapt(body []byte, _ map[string]interface{}) (
	[]byte, []caddyconfig.Warning, error) {
	// Using Thomas Pelletier's go-toml library to parse the configuration file contents.
	tree, err := toml.LoadBytes(body)
	if err != nil {
		return nil, nil, err
	}

	// Then convert the parsed TOML into a map and marshal the map to JSON.
	b, err := json.Marshal(tree.ToMap())

	return b, nil, err
}
